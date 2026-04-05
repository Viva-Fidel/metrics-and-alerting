package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

var (
	flagRunAddr   string
	reportInterval int64
	pollInterval   int64
)

func parseFlags(conf *config.Config) {

	// Флаги командной строки
    flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&reportInterval, "r", 10, "report sending interval")
	flag.Int64Var(&pollInterval, "p", 2, "report collecting interval")

    // Считываем флаги
	flag.Parse()

	// Переопределяем флаги значениями из конфигурации, если они указаны
	if envRunAddr := conf.Agent.Address; envRunAddr != "" {
        flagRunAddr = envRunAddr
    }

	if envReportInterval := conf.Agent.ReportInterval; envReportInterval != 0 {
        reportInterval = envReportInterval
    }

	if envPollInterval := conf.Agent.PollInterval; envPollInterval != 0 {
        pollInterval = envPollInterval
    }
}