package main

import (
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/gin-gonic/gin"
)


func setupRouter() *gin.Engine {

	router := gin.Default()

    // storages
    memStorage := repository.NewMemStorage()

	// handlers
    handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MemStorage: memStorage,
	})

	return router
}

func main() {
	parseFlags()
	r := setupRouter()
    r.Run(flagRunAddr)
}