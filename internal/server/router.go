package server

import (
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func NewRouter(metricsRepository service.MetricsRepository, middleware gin.HandlerFunc, hashKey string) *gin.Engine {
	router := gin.New()

	router.Use(middleware)
	router.Use(HashMiddleware(hashKey))
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
