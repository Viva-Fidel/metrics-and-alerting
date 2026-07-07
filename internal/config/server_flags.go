package config

import (
	"flag"
	"os"
)

// ServerFlags содержит итоговые параметры запуска сервера
type ServerFlags struct {
	RunAddr              string
	FilePath             string
	DB                   string
	Key                  string
	CryptoKey            string
	AuditFile            string
	AuditURL             string
	StoreInt             int64
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeSec int64
	DBConnMaxIdleTimeSec int64
	RestoreData          bool
}

// LoadServerFlags загружает параметры сервера из файла, флагов и переменных окружения
func LoadServerFlags() (*ServerFlags, error) {
	configPath := resolveConfigPath()

	flags := &ServerFlags{
		RunAddr:              "localhost:8080",
		StoreInt:             300,
		FilePath:             "metrics.json",
		DBMaxOpenConns:       10,
		DBMaxIdleConns:       5,
		DBConnMaxLifetimeSec: 300,
		DBConnMaxIdleTimeSec: 60,
	}

	if configPath != "" {
		if err := applyServerFileConfig(configPath, flags); err != nil {
			return nil, err
		}
	}

	configFlag := &configPathFlag{value: configPath}
	flag.Var(configFlag, "c", "path to JSON configuration file")
	flag.Var(configFlag, "config", "path to JSON configuration file")
	flag.StringVar(&flags.RunAddr, "a", flags.RunAddr, "server address")
	flag.Int64Var(&flags.StoreInt, "i", flags.StoreInt, "interval in seconds to save server metrics (0 = synchronous)")
	flag.StringVar(&flags.FilePath, "f", flags.FilePath, "path to file where server metrics are stored")
	flag.BoolVar(&flags.RestoreData, "r", flags.RestoreData, "load previously saved metrics on startup")
	flag.StringVar(&flags.DB, "d", flags.DB, "database DSN")
	flag.StringVar(&flags.Key, "k", flags.Key, "hash key")
	flag.StringVar(&flags.CryptoKey, "crypto-key", flags.CryptoKey, "path to private key file")
	flag.StringVar(&flags.AuditFile, "audit-file", flags.AuditFile, "path to audit log file")
	flag.StringVar(&flags.AuditURL, "audit-url", flags.AuditURL, "URL to send audit logs")

	flag.Parse()

	conf, err := LoadConfig()
	if err != nil {
		return nil, err
	}

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
	if conf.DB.DatabaseDSN != nil {
		flags.DB = *conf.DB.DatabaseDSN
	}
	if _, ok := os.LookupEnv("DATABASE_MAX_OPEN_CONNS"); ok {
		flags.DBMaxOpenConns = conf.DB.MaxOpenConns
	}
	if _, ok := os.LookupEnv("DATABASE_MAX_IDLE_CONNS"); ok {
		flags.DBMaxIdleConns = conf.DB.MaxIdleConns
	}
	if _, ok := os.LookupEnv("DATABASE_CONN_MAX_LIFETIME_SEC"); ok {
		flags.DBConnMaxLifetimeSec = conf.DB.ConnMaxLifetimeSec
	}
	if _, ok := os.LookupEnv("DATABASE_CONN_MAX_IDLE_TIME_SEC"); ok {
		flags.DBConnMaxIdleTimeSec = conf.DB.ConnMaxIdleTimeSec
	}
	if conf.Server.Key != "" {
		flags.Key = conf.Server.Key
	}
	if conf.Server.CryptoKey != "" {
		flags.CryptoKey = conf.Server.CryptoKey
	}
	if conf.Audit.FilePath != nil {
		flags.AuditFile = *conf.Audit.FilePath
	}
	if conf.Audit.URL != nil {
		flags.AuditURL = *conf.Audit.URL
	}

	return flags, nil
}
