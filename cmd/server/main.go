package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	serverapp "github.com/Viva-Fidel/metrics-and-alerting/internal/server"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/Viva-Fidel/metrics-and-alerting/pkg/db"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const shutdownTimeout = 10 * time.Second

func main() {
	printBuildInfo()

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем параметры запуска
	flags, err := config.LoadServerFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	var metricsRepository service.MetricsRepository
	var memRepo *repository.MemRepository

	useMemoryStorage := func() {
		memRepo = repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)
		metricsRepository = memRepo
	}

	// Если указан DSN - подключаемся к Postgres, при ошибке остаёмся на in-memory
	var database *sql.DB
	if dsn := strings.TrimSpace(flags.DB); dsn != "" {
		database, err = db.NewDB(dsn, db.Options{
			MaxOpenConns:    flags.DBMaxOpenConns,
			MaxIdleConns:    flags.DBMaxIdleConns,
			ConnMaxLifetime: time.Duration(flags.DBConnMaxLifetimeSec) * time.Second,
			ConnMaxIdleTime: time.Duration(flags.DBConnMaxIdleTimeSec) * time.Second,
		})
		if err != nil {
			logger.Error("failed to run init db", slog.Any("error", err))
			logger.Warn("using in-memory metrics storage")
			useMemoryStorage()
		} else if err := db.RunMigrations(database); err != nil {
			_ = database.Close()
			database = nil
			logger.Error("failed to run migrations", slog.Any("error", err))
			logger.Warn("using in-memory metrics storage")
			useMemoryStorage()
		} else {
			metricsRepository = repository.NewDBRepository(database)
		}
	} else {
		useMemoryStorage()
	}

	auditPublisher := audit.NewPublisher(logger, flags.AuditFile, flags.AuditURL)
	defer func() {
		if err := auditPublisher.Close(); err != nil {
			logger.Error("failed to close audit publisher", slog.Any("error", err))
		}
	}()

	var privateKey *rsa.PrivateKey
	if flags.CryptoKey != "" {
		privateKey, err = security.LoadPrivateKey(flags.CryptoKey)
		if err != nil {
			logger.Error("failed to load private key", slog.Any("error", err))
			os.Exit(1)
		}
	}

	trustedSubnetMiddleware, err := serverapp.TrustedSubnetMiddleware(flags.TrustedSubnet)
	if err != nil {
		logger.Error("failed to init trusted subnet middleware", slog.Any("error", err))
		os.Exit(1)
	}

	metricsService := service.NewMetricsService(metricsRepository)

	// Собираем HTTP-роутер
	r := serverapp.NewRouter(metricsRepository, logging.SlogMiddleware(logger), flags.Key, privateKey, auditPublisher, trustedSubnetMiddleware)

	grpcServer, err := serverapp.NewGRPCServer(flags.GRPCAddress, metricsService, auditPublisher, flags.TrustedSubnet, logger)
	if err != nil {
		logger.Error("failed to init gRPC server", slog.Any("error", err))
		os.Exit(1)
	}

	// Graceful shutdown по сигналам SIGINT, SIGTERM, SIGQUIT
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Запускаем HTTP-сервер
	srv := &http.Server{
		Addr:    flags.RunAddr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to run server", slog.Any("error", err))
			stop() // при ошибке запуска инициируем завершение
		}
	}()

	if grpcServer != nil {
		go func() {
			if err := grpcServer.Serve(); err != nil {
				logger.Error("failed to run gRPC server", slog.Any("error", err))
				stop()
			}
		}()
	}

	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Дожидаемся завершения активных запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server gracefully", slog.Any("error", err))
	}

	if grpcServer != nil {
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-shutdownCtx.Done():
			grpcServer.Stop()
		}
	}

	// Сохраняем все несохранённые данные in-memory хранилища
	if memRepo != nil {
		if err := memRepo.Close(); err != nil {
			logger.Error("failed to save metrics on shutdown", slog.Any("error", err))
		}
	}

	// Закрываем соединение с БД
	if database != nil {
		if err := database.Close(); err != nil {
			logger.Error("failed to close database", slog.Any("error", err))
		}
	}

	logger.Info("server stopped")
}
