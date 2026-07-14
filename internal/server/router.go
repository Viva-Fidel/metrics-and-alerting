// Package server собирает HTTP-роутер сервиса сбора метрик
package server

import (
	"crypto/rsa"
	"net/http"
	_ "net/http/pprof"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт и настраивает gin.Engine с маршрутами метрик, middleware и pprof
func NewRouter(metricsRepository service.MetricsRepository, middleware gin.HandlerFunc, hashKey string, privateKey *rsa.PrivateKey, auditPublisher *audit.Publisher) *gin.Engine {
	router := gin.New()

	router.Use(middleware)
	router.Use(HashMiddleware(hashKey))
	if privateKey != nil {
		router.Use(CryptoMiddleware(privateKey))
	}
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithDecompressFn(gzip.DefaultDecompressHandle),
	))

	metricsService := service.NewMetricsService(metricsRepository)
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: metricsService,
		AuditPublisher: auditPublisher,
	})

	registerPprof(router)

	return router
}

func registerPprof(router *gin.Engine) {
	router.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
}
