package main

import (
	"fmt"
	"net/http"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/pkg/storage"
)


func main() {

    router := http.NewServeMux()

    // storage
    memStorage := storage.NewMemStorage()


	// handlers
    handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		Storage: memStorage,
	})


    server := http.Server{
		Addr: ":8080",
		Handler: router,
	}
    
    fmt.Println("Сервер запушен")
	server.ListenAndServe()
}
