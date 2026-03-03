package main

import (
	"fmt"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
)


func main() {

	parseFlags()

	metrics := agent.NewMetrics()

	elapsedTime := int64(0)

    fmt.Println("Агент запущен")

	for {
		// Забор метрик
		agent.PollMetrics(metrics)

		elapsedTime += pollInterval

		// Отправка метрик на сервер
		if elapsedTime >= reportInterval {
			fmt.Println("Отправка метрик")
			agent.ReportMetrics(metrics, flagRunAddr)
			elapsedTime = 0
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}


