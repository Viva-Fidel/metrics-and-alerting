package main

import (
	"log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	serverapp "github.com/Viva-Fidel/metrics-and-alerting/internal/server"
)

func main() {
	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем параметры запуска
	flags, err := config.LoadServerFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	// Инициализируем репозиторий метрик с настройками
	metricsRepository := repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)

	// Собираем HTTP-роутер
	r := serverapp.NewRouter(metricsRepository, logging.SlogMiddleware(logger))

	// Запускаем HTTP-сервер
	if err := r.Run(flags.RunAddr); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
		os.Exit(1)
	}
}
