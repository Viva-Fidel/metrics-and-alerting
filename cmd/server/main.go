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


func setupRouter(metricsRepository *repository.MemRepository) *gin.Engine {

	// Инициализация логгера
	newLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router := gin.New()

	// logger
    router.Use(logger.SlogMiddleware(newLogger))
    router.Use(gin.Recovery())

	// gzip
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	// services
	metricsService := service.NewMetricsService(metricsRepository)
	
	// handlers
    handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: metricsService,
	})

	return router
}


func main() {
	// Загружаем конфигурацию из env
	conf, _ := config.LoadConfig()

	// Считываем флаги и переопределяем конфигом
	parseFlags(conf)

	metricsRepository := repository.NewMemRepository()
	metricsRepository.ConfigureStorage(flagFilePath, flagStoreInt)
	if flagRestoreData {
		if err := metricsRepository.LoadFromFile(); err != nil {
			panic(err)
		}
	}
	metricsRepository.StartSaver()

	r := setupRouter(metricsRepository)
	r.Run(flagRunAddr)
}