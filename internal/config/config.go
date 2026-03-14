package config

import (


	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Agent AgentConfig
}

type AgentConfig struct {
	ADDRESS     string `env:"ADDRESS" envDefault:"localhost:8081"`
	REPORT_INTERVAL int64 `env:"REPORT_INTERVAL" envDefault:"10"`
	POLL_INTERVAL   int64 `env:"POLL_INTERVAL" envDefault:"2"`
}

type ServerConfig struct {
	ADDRESS string `env:"ADDRESS" envDefault:"localhost:8080"`
}


func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}