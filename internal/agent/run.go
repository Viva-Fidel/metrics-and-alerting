package agent

import (
	"context"
	"log/slog"
	"time"

	"resty.dev/v3"
)

func Run(
	ctx context.Context,
	logger *slog.Logger,
	client *resty.Client,
	metrics *Metrics,
	pollInterval int64,
	reportInterval int64,
) {
	elapsedTime := int64(0)

	logger.Info("Запуск агента")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Остановка агента")
			return
		default:
		}

		PollMetrics(metrics)
		elapsedTime += pollInterval

		if elapsedTime >= reportInterval {
			logger.Info("Отправка метрик")
			ReportMetrics(ctx, client, metrics)
			elapsedTime = 0
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}
