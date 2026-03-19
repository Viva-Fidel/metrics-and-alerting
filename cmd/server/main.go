package main

import (
	"fmt"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/gin-gonic/gin"
)


func setupRouter() *gin.Engine {

	router := gin.New()
	router.Use(gin.Recovery())

	// кастомный логгер gin
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf("[%s] %s %s -> %d %d bytes in %v\n",
        param.TimeStamp.Format(time.RFC3339),
        param.Method,
        param.Path,
        param.StatusCode,
        param.BodySize,
        param.Latency,
    )}))
	

    // storages
    memStorage := repository.NewMemStorage()

	// handlers
    handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MemStorage: memStorage,
	})

	return router
}


func main() {
	conf, _ := config.LoadConfig()

	parseFlags(conf)
	r := setupRouter()
    r.Run(flagRunAddr)
}