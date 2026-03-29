package main

import (
	"context"
	"fmt"
	"log"
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
	conf, err := config.LoadConfig()
	if err != nil {
		log.Printf("failed to load config: %v", err)
	}

	// Считываем флаги и переопределяем конфигом
	flags := parseFlags(conf)

	// Инициализация метрик
	metrics := agent.NewMetrics()
	elapsedTime := int64(0)

	// Создание клиента resty
	client := resty.New().SetBaseURL("http://" + flags.FlagRunAddr)
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

		elapsedTime += flags.PollInterval

		// Отправка метрик на сервер
		if elapsedTime >= flags.ReportInterval {
			fmt.Println("Отправка метрик")
			agent.ReportMetrics(ctx, client, metrics)
			elapsedTime = 0
		}

		time.Sleep(time.Duration(flags.PollInterval) * time.Second)
	}
}


