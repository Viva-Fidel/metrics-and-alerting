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

    flag.StringVar(&flagRunAddr, "a", "localhost:8081", "server address")
	flag.Int64Var(&reportInterval, "r", 10, "report sending interval")
	flag.Int64Var(&pollInterval, "p", 2, "report collecting interval")

    flag.Parse()

	if envRunAddr := conf.Agent.ADDRESS; envRunAddr != "" {
        flagRunAddr = envRunAddr
    }

	if envReportInterval := conf.Agent.REPORT_INTERVAL; envReportInterval != 0 {
        reportInterval = envReportInterval
    }

	if envPollInterval := conf.Agent.POLL_INTERVAL; envPollInterval != 0 {
        pollInterval = envPollInterval
    }
}