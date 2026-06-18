package agent

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"resty.dev/v3"
)

// Run запускает агента, собирает метрики и отправляет их на сервер
func Run(
	ctx context.Context,
	logger *slog.Logger,
	client *resty.Client,
	metrics *Metrics,
	pollInterval int64,
	reportInterval int64,
	rateLimit int,
	hashKey string,
) {
	logger.Info("Запуск агента")

	if rateLimit < 1 {
		rateLimit = 1
	}

	// Канал для отправки метрик на сервер
	reportJobs := make(chan struct{}, rateLimit)
	var workersWG sync.WaitGroup

	for i := 0; i < rateLimit; i++ {
		workerID := i + 1
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for range reportJobs {
				logger.Info("Отправка метрик", slog.Int("worker_id", workerID))
				ReportMetrics(ctx, client, metrics, hashKey)
			}
		}()
	}

	// WaitGroup для ожидания завершения всех циклов
	var loopsWG sync.WaitGroup
	loopsWG.Add(3)

	go func() {
		defer loopsWG.Done()
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				PollMetrics(metrics)
			}
		}
	}()

	// Цикл для сбора системных метрик
	go func() {
		defer loopsWG.Done()
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := PollSystemMetrics(metrics); err != nil {
					logger.Warn("Не удалось собрать системные метрики", slog.Any("error", err))
				}
			}
		}
	}()

	// Цикл для отправки метрик на сервер
	go func() {
		defer loopsWG.Done()
		ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case reportJobs <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Ожидание завершения всех циклов
	<-ctx.Done()
	loopsWG.Wait()
	close(reportJobs)
	workersWG.Wait()
	logger.Info("Остановка агента")
}
