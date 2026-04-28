package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server     ServerConfig
	Agent      AgentConfig
	LogMetrics LogMetricsConfig
	Db         DbConfig
}

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	Key            string `env:"KEY"`
}

type ServerConfig struct {
	Address *string `env:"ADDRESS"`
	Key     string  `env:"KEY"`
}

type LogMetricsConfig struct {
	StoreInterval *int64  `env:"STORE_INTERVAL"`
	FilePath      *string `env:"FILE_STORAGE_PATH"`
	Restore       *bool   `env:"RESTORE"`
}

type DbConfig struct {
	DATABASE_DSN       *string `env:"DATABASE_DSN"`
	MaxOpenConns       int     `env:"DATABASE_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns       int     `env:"DATABASE_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetimeSec int64   `env:"DATABASE_CONN_MAX_LIFETIME_SEC" envDefault:"300"`
	ConnMaxIdleTimeSec int64   `env:"DATABASE_CONN_MAX_IDLE_TIME_SEC" envDefault:"60"`
}

func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
