package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

var flagRunAddr string

func parseFlags(conf *config.Config) {

    // Флаги командной строки
    flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address")

    // Считываем флаги
    flag.Parse()

    // Переопределяем флаги значениями из конфигурации, если они указаны
    if conf.Server.Address != "" {
        flagRunAddr = conf.Server.Address
    }
}