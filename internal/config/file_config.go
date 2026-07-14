package config

// AgentFileConfig описывает параметры агента в JSON-файле конфигурации.
type AgentFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *int64  `json:"report_interval"`
	PollInterval   *int64  `json:"poll_interval"`
	RateLimit      *int    `json:"rate_limit"`
	Key            *string `json:"key"`
	CryptoKey      *string `json:"crypto_key"`
}

// ServerFileConfig описывает параметры сервера в JSON-файле конфигурации.
type ServerFileConfig struct {
	Address                    *string `json:"address"`
	StoreInterval              *int64  `json:"store_interval"`
	StoreFile                  *string `json:"store_file"`
	Restore                    *bool   `json:"restore"`
	DatabaseDSN                *string `json:"database_dsn"`
	Key                        *string `json:"key"`
	CryptoKey                  *string `json:"crypto_key"`
	AuditFile                  *string `json:"audit_file"`
	AuditURL                   *string `json:"audit_url"`
	DatabaseMaxOpenConns       *int    `json:"database_max_open_conns"`
	DatabaseMaxIdleConns       *int    `json:"database_max_idle_conns"`
	DatabaseConnMaxLifetimeSec *int64  `json:"database_conn_max_lifetime_sec"`
	DatabaseConnMaxIdleTimeSec *int64  `json:"database_conn_max_idle_time_sec"`
}

func loadAgentEnvLayer() (*AgentFileConfig, error) {
	reportInterval, err := envInt64("REPORT_INTERVAL")
	if err != nil {
		return nil, err
	}
	pollInterval, err := envInt64("POLL_INTERVAL")
	if err != nil {
		return nil, err
	}
	rateLimit, err := envInt("RATE_LIMIT")
	if err != nil {
		return nil, err
	}

	return &AgentFileConfig{
		Address:        envString("ADDRESS"),
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		RateLimit:      rateLimit,
		Key:            envString("KEY"),
		CryptoKey:      envString("CRYPTO_KEY"),
	}, nil
}

func loadServerEnvLayer() (*ServerFileConfig, error) {
	storeInterval, err := envInt64("STORE_INTERVAL")
	if err != nil {
		return nil, err
	}
	restore, err := envBool("RESTORE")
	if err != nil {
		return nil, err
	}
	maxOpenConns, err := envInt("DATABASE_MAX_OPEN_CONNS")
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := envInt("DATABASE_MAX_IDLE_CONNS")
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := envInt64("DATABASE_CONN_MAX_LIFETIME_SEC")
	if err != nil {
		return nil, err
	}
	connMaxIdleTime, err := envInt64("DATABASE_CONN_MAX_IDLE_TIME_SEC")
	if err != nil {
		return nil, err
	}

	return &ServerFileConfig{
		Address:                    envString("ADDRESS"),
		StoreInterval:              storeInterval,
		StoreFile:                  envString("FILE_STORAGE_PATH"),
		Restore:                    restore,
		DatabaseDSN:                envString("DATABASE_DSN"),
		Key:                        envString("KEY"),
		CryptoKey:                  envString("CRYPTO_KEY"),
		AuditFile:                  envString("AUDIT_FILE"),
		AuditURL:                   envString("AUDIT_URL"),
		DatabaseMaxOpenConns:       maxOpenConns,
		DatabaseMaxIdleConns:       maxIdleConns,
		DatabaseConnMaxLifetimeSec: connMaxLifetime,
		DatabaseConnMaxIdleTimeSec: connMaxIdleTime,
	}, nil
}

func agentFileConfigToFlags(cfg *AgentFileConfig) *AgentFlags {
	return &AgentFlags{
		RunAddr:        derefString(cfg.Address, "localhost:8080"),
		ReportInterval: derefInt64(cfg.ReportInterval, 10),
		PollInterval:   derefInt64(cfg.PollInterval, 2),
		RateLimit:      derefInt(cfg.RateLimit, 1),
		Key:            derefString(cfg.Key, ""),
		CryptoKey:      derefString(cfg.CryptoKey, ""),
	}
}

func serverFileConfigToFlags(cfg *ServerFileConfig) *ServerFlags {
	return &ServerFlags{
		RunAddr:              derefString(cfg.Address, "localhost:8080"),
		StoreInt:             derefInt64(cfg.StoreInterval, 300),
		FilePath:             derefString(cfg.StoreFile, "metrics.json"),
		RestoreData:          derefBool(cfg.Restore, false),
		DB:                   derefString(cfg.DatabaseDSN, ""),
		Key:                  derefString(cfg.Key, ""),
		CryptoKey:            derefString(cfg.CryptoKey, ""),
		AuditFile:            derefString(cfg.AuditFile, ""),
		AuditURL:             derefString(cfg.AuditURL, ""),
		DBMaxOpenConns:       derefInt(cfg.DatabaseMaxOpenConns, 10),
		DBMaxIdleConns:       derefInt(cfg.DatabaseMaxIdleConns, 5),
		DBConnMaxLifetimeSec: derefInt64(cfg.DatabaseConnMaxLifetimeSec, 300),
		DBConnMaxIdleTimeSec: derefInt64(cfg.DatabaseConnMaxIdleTimeSec, 60),
	}
}

func derefString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func derefInt64(value *int64, fallback int64) int64 {
	if value == nil {
		return fallback
	}
	return *value
}

func derefInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func derefBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
