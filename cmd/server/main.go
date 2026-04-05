package main

import (
	"log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logger"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
)


func setupRouter() *gin.Engine {

	// Инициализация логгера
	newLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router := gin.New()

	// logger
    router.Use(logger.SlogMiddleware(newLogger))
    router.Use(gin.Recovery())

	// gzip
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	// repos
	metricsRepository := repository.NewMemRepository()

	// services
	metricsService := service.NewMetricsService(metricsRepository)
	
	// handlers
    handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: metricsService,
	})

	return router
}


func main() {
	conf, _ := config.LoadConfig()

	parseFlags(conf)
	r := setupRouter()
    r.Run(flagRunAddr)
}