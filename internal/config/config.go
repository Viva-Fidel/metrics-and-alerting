package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Agent AgentConfig
	LogMetrics LogMetricsConfig
	Db DbConfig
}

type AgentConfig struct {
	Address     string `env:"ADDRESS"`
	ReportInterval int64 `env:"REPORT_INTERVAL"`
	PollInterval   int64 `env:"POLL_INTERVAL"`
}

type ServerConfig struct {
	Address *string `env:"ADDRESS"`
}

type LogMetricsConfig struct {
	StoreInterval *int64  `env:"STORE_INTERVAL"`
	FilePath      *string `env:"FILE_STORAGE_PATH"`
	Restore       *bool   `env:"RESTORE"`
}

type DbConfig struct {
	DATABASE_DSN *string `env:"DATABASE_DSN"`
}

func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}