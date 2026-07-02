package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	serverapp "github.com/Viva-Fidel/metrics-and-alerting/internal/server"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/Viva-Fidel/metrics-and-alerting/pkg/db"
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
	flags, err := config.LoadServerFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	// Инициализируем репозиторий метрик с настройками
	var metricsRepository service.MetricsRepository = repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)

	// Если указан DSN - подключаемся к Postgres, при ошибке остаёмся на in-memory
	if dsn := strings.TrimSpace(flags.DB); dsn != "" {
		database, err := db.NewDB(dsn, db.Options{
			MaxOpenConns:    flags.DBMaxOpenConns,
			MaxIdleConns:    flags.DBMaxIdleConns,
			ConnMaxLifetime: time.Duration(flags.DBConnMaxLifetimeSec) * time.Second,
			ConnMaxIdleTime: time.Duration(flags.DBConnMaxIdleTimeSec) * time.Second,
		})
		if err != nil {
			logger.Error("failed to run init db", slog.Any("error", err))
			logger.Warn("using in-memory metrics storage")
		} else {
			if err := db.RunMigrations(database); err != nil {
				_ = database.Close()
				logger.Error("failed to run migrations", slog.Any("error", err))
				logger.Warn("using in-memory metrics storage")
			} else {
				metricsRepository = repository.NewDBRepository(database)
			}
		}
	}

	auditPublisher := audit.NewPublisher(logger, flags.AuditFile, flags.AuditURL)
	defer func() {
		if err := auditPublisher.Close(); err != nil {
			logger.Error("failed to close audit publisher", slog.Any("error", err))
		}
	}()

	// Собираем HTTP-роутер
	r := serverapp.NewRouter(metricsRepository, logging.SlogMiddleware(logger), flags.Key, auditPublisher)

	// Запускаем HTTP-сервер
	if err := r.Run(flags.RunAddr); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
	}
}
