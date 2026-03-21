package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	"resty.dev/v3"
)


func main() {
	// Загружаем конфигурацию из env
	conf, _ := config.LoadConfig()

	// Считываем флаги и переопределяем конфигом
	parseFlags(conf)

	// Инициализация метрик
	metrics := agent.NewMetrics()
	elapsedTime := int64(0)

	// Создание клиента resty
	client := resty.New().SetBaseURL("http://" + flagRunAddr)
	defer client.Close()

	// Контекст для корректного закрытия
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

    fmt.Println("Агент запущен")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Завершение работы агента")
			return
		default:
		}

		// Забор метрик
		agent.PollMetrics(metrics)

		elapsedTime += pollInterval

		// Отправка метрик на сервер
		if elapsedTime >= reportInterval {
			fmt.Println("Отправка метрик")
			agent.ReportMetrics(ctx, client, metrics)
			elapsedTime = 0
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}


