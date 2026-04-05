package server

import (
	"database/sql"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func NewRouter(metricsRepository *repository.MemRepository, middleware gin.HandlerFunc, db *sql.DB) *gin.Engine {
	router := gin.New()

	router.Use(middleware)
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithDecompressFn(gzip.DefaultDecompressHandle),
	))

	metricsService := service.NewMetricsService(metricsRepository)
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: metricsService,
		DB:             db,
	})

	return router
}
