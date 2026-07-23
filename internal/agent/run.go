package agent

import (
	"context"
	"crypto/rsa"
	"log/slog"
	"sync"
	"time"

	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"resty.dev/v3"
)

// Run запускает агента, собирает метрики и отправляет их на сервер
func Run(
	ctx context.Context,
	logger *slog.Logger,
	client *resty.Client,
	grpcClient pb.MetricsClient,
	hostIP string,
	metrics *Metrics,
	pollInterval int64,
	reportInterval int64,
	rateLimit int,
	hashKey string,
	publicKey *rsa.PublicKey,
) {
	logger.Info("Запуск агента")

	if rateLimit < 1 {
		rateLimit = 1
	}

	// Отдельный контекст отправки: не отменяется по сигналу,
	// чтобы данные в процессе обработки были успешно переданы на сервер.
	// При завершении заменяется на контекст с таймаутом, чтобы не зависнуть
	// при недоступном сервере.
	sendCtx := context.Background()

	report := func(ctx context.Context) {
		if grpcClient != nil {
			ReportMetricsGRPC(ctx, grpcClient, metrics, hostIP)
			return
		}
		ReportMetrics(ctx, client, metrics, hashKey, publicKey)
	}

	// Канал для отправки метрик на сервер
	reportJobs := make(chan struct{}, rateLimit)
	var workersWG sync.WaitGroup

	for i := range rateLimit {
		workerID := i + 1
		workersWG.Go(func() {
			for range reportJobs {
				logger.Info("Отправка метрик", slog.Int("worker_id", workerID))
				report(sendCtx)
			}
		})
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
	logger.Info("Получен сигнал завершения")

	var cancel context.CancelFunc
	sendCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	loopsWG.Wait()
	close(reportJobs)
	workersWG.Wait()

	// Финальная отправка: передаём на сервер метрики,
	// собранные, но ещё не отправленные к моменту получения сигнала
	logger.Info("Финальная отправка метрик перед завершением")
	report(sendCtx)

	logger.Info("Остановка агента")
}
