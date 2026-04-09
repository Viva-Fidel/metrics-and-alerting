package main

import (
	"log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	serverapp "github.com/Viva-Fidel/metrics-and-alerting/internal/server"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/Viva-Fidel/metrics-and-alerting/pkg/db"
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
	var metricsRepository service.MetricsRepository = repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)

	// Если указан DSN для подключения к базе данных, инициализируем БД, применяем миграции и пересоздаём репозиторий для работы с Postgres.
	database, err := db.NewDB(flags.Db)
	if err != nil {
		logger.Error("failed to run init db", slog.Any("error", err))
	} else {
		if err := db.RunMigrations(database); err != nil {
			logger.Error("failed to run migrations", slog.Any("error", err))
		} else {
			metricsRepository = repository.NewDBRepository(database)
		}
	}

	// Собираем HTTP-роутер
	r := serverapp.NewRouter(metricsRepository, logging.SlogMiddleware(logger), database)

	// Запускаем HTTP-сервер
	if err := r.Run(flags.RunAddr); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
		os.Exit(1)
	}
}
