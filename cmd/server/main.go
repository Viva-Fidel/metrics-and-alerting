package main

import (
	"log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	serverapp "github.com/Viva-Fidel/metrics-and-alerting/internal/server"
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
	metricsRepository := repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)

	db, err := db.NewDB(flags.Db)
    if err != nil {
    	logger.Error("failed to run init db", slog.Any("error", err))
    }

	// Собираем HTTP-роутер
	r := serverapp.NewRouter(metricsRepository, logging.SlogMiddleware(logger), db)

	// Запускаем HTTP-сервер
	if err := r.Run(flags.RunAddr); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
		os.Exit(1)
	}
}
