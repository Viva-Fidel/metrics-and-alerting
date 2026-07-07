package config

import "flag"

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

// LoadServerFlags загружает параметры сервера из флагов и переменных окружения
func LoadServerFlags() (*ServerFlags, error) {
	conf, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	flags := &ServerFlags{}

	flag.StringVar(&flags.RunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&flags.StoreInt, "i", 300, "interval in seconds to save server metrics (0 = synchronous)")
	flag.StringVar(&flags.FilePath, "f", "metrics.json", "path to file where server metrics are stored")
	flag.BoolVar(&flags.RestoreData, "r", false, "load previously saved metrics on startup")
	flag.StringVar(&flags.DB, "d", "", "database DSN")
	flag.StringVar(&flags.Key, "k", "", "hash key")
	flag.StringVar(&flags.CryptoKey, "crypto-key", "", "path to private key file")
	flag.StringVar(&flags.AuditFile, "audit-file", "", "path to audit log file")
	flag.StringVar(&flags.AuditURL, "audit-url", "", "URL to send audit logs")

	flag.Parse()

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
	flags.DBMaxOpenConns = conf.DB.MaxOpenConns
	flags.DBMaxIdleConns = conf.DB.MaxIdleConns
	flags.DBConnMaxLifetimeSec = conf.DB.ConnMaxLifetimeSec
	flags.DBConnMaxIdleTimeSec = conf.DB.ConnMaxIdleTimeSec
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
