package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"resty.dev/v3"
)

func main() {
	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем параметры запуска
	flags, err := config.LoadAgentFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	// Инициализируем хранилище
	metrics := agent.NewMetrics()

	// Конфигурируем HTTP-клиент
	client := resty.New().SetBaseURL("http://" + flags.RunAddr)
	defer client.Close()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем основной цикл сбора и отправки метрик
	agent.Run(ctx, logger, client, metrics, flags.PollInterval, flags.ReportInterval, flags.Key)
}
