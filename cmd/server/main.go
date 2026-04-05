package main

import (
	"log/slog"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/logger"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/gin-gonic/gin"
)


func setupRouter() *gin.Engine {

	newLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))


	router := gin.New()

	// logger
    router.Use(logger.SlogMiddleware(newLogger))
    router.Use(gin.Recovery())
	

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