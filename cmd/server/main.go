package main

import (
	
    "log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logging"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)


func setupRouter(metricsRepository *repository.MemRepository, logger *slog.Logger) *gin.Engine {
	router := gin.New()

	router.Use(logging.SlogMiddleware(logger))
	router.Use(gin.Recovery())

	router.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithDecompressFn(gzip.DefaultDecompressHandle),
	))

	metricsService := service.NewMetricsService(metricsRepository)

	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: metricsService,
	})

	return router
}


func main() {
	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// conf
	conf, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	// flags
	flags := parseFlags(conf)

	// repo
	metricsRepository := repository.NewMemRepository(logger, flags.FilePath, flags.StoreInt, flags.RestoreData)


	// router
	r := setupRouter(metricsRepository, logger)

	// server run
	if err := r.Run(flags.RunAddr); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
		os.Exit(1)
	}
}