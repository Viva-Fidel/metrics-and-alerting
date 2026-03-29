package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

type ServerFlags struct {
	RunAddr     string
	StoreInt    int64
	FilePath    string
	RestoreData bool
}


func parseFlags(conf *config.Config) *ServerFlags {
	flags := &ServerFlags{}

	// Флаги командной строки
	flag.StringVar(&flags.RunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&flags.StoreInt, "i", 300, "interval in seconds to save server metrics (0 = synchronous)")
	flag.StringVar(&flags.FilePath, "f", "metrics.json", "path to file where server metrics are stored")
	flag.BoolVar(&flags.RestoreData, "r", false, "load previously saved metrics on startup")

	// Считываем флаги
	flag.Parse()

	// Переопределяем флаги значениями из конфигурации, если они указаны
	if conf.Server.Address != nil {
		flags.RunAddr = *conf.Server.Address
	}

	if conf.LogMetrics.StoreInterval != nil {
		flags.StoreInt = *conf.LogMetrics.StoreInterval
	}

	if conf.LogMetrics.FilePath != nil {
		flags.FilePath = *conf.LogMetrics.FilePath
	}

	if conf.LogMetrics.Restore != nil {
		flags.RestoreData = *conf.LogMetrics.Restore
	}

	return flags
}