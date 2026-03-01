package main

import (
	"fmt"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
)


func main() {
	metrics := agent.NewMetrics()

	pollInterval := 2 // Время обновления метрик
	reportInterval := 10 // Отправка метрик с заданной частотой
	elapsedTime := 0

    fmt.Println("Агент запущен")

	for {
		// Забор метрик
		agent.PollMetrics(metrics)

		elapsedTime += pollInterval

		// Отправка метрик на сервер
		if elapsedTime >= reportInterval {
			fmt.Println("Отправка метрик")
			agent.ReportMetrics(metrics)
			elapsedTime = 0
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}


