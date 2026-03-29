package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

type AgentFlags struct {
	FlagRunAddr     string
	ReportInterval int64
	PollInterval   int64
}

func parseFlags(conf *config.Config) *AgentFlags{
	flags := &AgentFlags{}

	// Флаги командной строки
    flag.StringVar(&flags.FlagRunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&flags.ReportInterval, "r", 10, "report sending interval")
	flag.Int64Var(&flags.PollInterval, "p", 2, "report collecting interval")

    // Считываем флаги
	flag.Parse()

	// Переопределяем флаги значениями из конфигурации, если они указаны
	if envRunAddr := conf.Agent.Address; envRunAddr != "" {
        flags.FlagRunAddr = envRunAddr
    }

	if envReportInterval := conf.Agent.ReportInterval; envReportInterval != 0 {
        flags.ReportInterval = envReportInterval
    }

	if envPollInterval := conf.Agent.PollInterval; envPollInterval != 0 {
        flags.PollInterval = envPollInterval
    }

	return flags
}