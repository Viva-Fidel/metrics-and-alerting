package main

import (
	"context"
	"crypto/rsa"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"resty.dev/v3"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем параметры запуска
	flags, err := config.LoadAgentFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	var publicKey *rsa.PublicKey
	if flags.CryptoKey != "" {
		publicKey, err = security.LoadPublicKey(flags.CryptoKey)
		if err != nil {
			logger.Error("failed to load public key", slog.Any("error", err))
			os.Exit(1)
		}
	}

	// Инициализируем хранилище
	metrics := agent.NewMetrics()

	hostIP, err := agent.HostIP()
	if err != nil {
		logger.Error("failed to resolve host IP", slog.Any("error", err))
		os.Exit(1)
	}

	// Конфигурируем HTTP-клиент
	client := resty.New().
		SetBaseURL("http://" + flags.RunAddr).
		SetHeader(security.RealIPHeader, hostIP).
		SetRetryCount(3).
		SetRetryWaitTime(time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		AddRetryConditions(func(_ *resty.Response, err error) bool {
			return err != nil
		})
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error("failed to close http client", slog.Any("error", err))
		}
	}()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Запускаем основной цикл сбора и отправки метрик
	agent.Run(ctx, logger, client, metrics, flags.PollInterval, flags.ReportInterval, flags.RateLimit, flags.Key, publicKey)
}
