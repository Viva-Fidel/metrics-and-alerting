package config

import (


	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Agent AgentConfig
}

type AgentConfig struct {
	Address     string `env:"ADDRESS"`
	ReportInterval int64 `env:"REPORT_INTERVAL"`
	PollInterval   int64 `env:"POLL_INTERVAL"`
}

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}


func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}