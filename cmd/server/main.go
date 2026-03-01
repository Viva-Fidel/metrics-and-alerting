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


    // server
	server := http.Server{
		Addr: ":8080",
		Handler: router,
	}
    
    fmt.Println("Сервер запущен")
	err := server.ListenAndServe()
	
	// Вывод ошибок, если сервер не запустился
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
