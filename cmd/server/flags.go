package main

import (
	"flag"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
)

var (
	flagRunAddr     string
	flagStoreInt    int64
	flagFilePath    string
	flagRestoreData bool
)

// parseFlags разбирает флаги и накладывает конфиг из окружения (через config.LoadConfig).
// Приоритет: если переменная окружения задана — значение из conf; иначе остаётся значение флага (или дефолт флага).
func parseFlags(conf *config.Config) {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&flagStoreInt, "i", 300, "interval in seconds to save server metrics (0 = synchronous)")
	flag.StringVar(&flagFilePath, "f", "metrics.json", "path to file where server metrics are stored")
	flag.BoolVar(&flagRestoreData, "r", false, "load previously saved metrics on startup")

	flag.Parse()

	if conf.Server.Address != nil {
		flagRunAddr = *conf.Server.Address
	}

	if conf.LogMetrics.StoreInterval != nil {
		flagStoreInt = *conf.LogMetrics.StoreInterval
	}

	if conf.LogMetrics.FilePath != nil {
		flagFilePath = *conf.LogMetrics.FilePath
	}

	if conf.LogMetrics.Restore != nil {
		flagRestoreData = *conf.LogMetrics.Restore
	}
}