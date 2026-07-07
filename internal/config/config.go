// Package config загружает параметры сервера и агента из переменных окружения и флагов
package config

import (
	"github.com/caarlos0/env/v11"
)

// Config объединяет настройки всех компонентов приложения
type Config struct {
	LogMetrics LogMetricsConfig
	Audit      AuditConfig
	Agent      AgentConfig
	Server     ServerConfig
	DB         DBConfig
}

// AgentConfig содержит параметры агента сбора метрик
type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// ServerConfig содержит параметры HTTP-сервера
type ServerConfig struct {
	Address   *string `env:"ADDRESS"`
	Key       string  `env:"KEY"`
	CryptoKey string  `env:"CRYPTO_KEY"`
}

// LogMetricsConfig содержит параметры файлового хранилища метрик
type LogMetricsConfig struct {
	StoreInterval *int64  `env:"STORE_INTERVAL"`
	FilePath      *string `env:"FILE_STORAGE_PATH"`
	Restore       *bool   `env:"RESTORE"`
}

// DBConfig содержит параметры подключения к PostgreSQL
type DBConfig struct {
	DatabaseDSN        *string `env:"DATABASE_DSN"`
	MaxOpenConns       int     `env:"DATABASE_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns       int     `env:"DATABASE_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetimeSec int64   `env:"DATABASE_CONN_MAX_LIFETIME_SEC" envDefault:"300"`
	ConnMaxIdleTimeSec int64   `env:"DATABASE_CONN_MAX_IDLE_TIME_SEC" envDefault:"60"`
}

// AuditConfig содержит параметры аудита изменений метрик
type AuditConfig struct {
	FilePath *string `env:"AUDIT_FILE"`
	URL      *string `env:"AUDIT_URL"`
}

// LoadConfig читает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
