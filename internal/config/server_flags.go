package config

import "os"

// ServerFlags содержит итоговые параметры запуска сервера
type ServerFlags struct {
	RunAddr              string
	FilePath             string
	DB                   string
	Key                  string
	CryptoKey            string
	TrustedSubnet        string
	AuditFile            string
	AuditURL             string
	StoreInt             int64
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeSec int64
	DBConnMaxIdleTimeSec int64
	RestoreData          bool
}

// LoadServerFlags загружает параметры сервера из файла, переменных окружения и флагов
func LoadServerFlags() (*ServerFlags, error) {
	configPath, err := resolveConfigPath(os.Args[1:])
	if err != nil {
		return nil, err
	}

	fileLayer := &ServerFileConfig{}
	if err := loadJSONConfig(configPath, fileLayer); err != nil {
		return nil, err
	}

	envLayer, err := loadServerEnvLayer()
	if err != nil {
		return nil, err
	}

	flagLayer, err := parseServerFlagLayer()
	if err != nil {
		return nil, err
	}

	merged, err := mergeOverlays(fileLayer, envLayer, flagLayer)
	if err != nil {
		return nil, err
	}

	return serverFileConfigToFlags(merged), nil
}

func parseServerFlagLayer() (*ServerFileConfig, error) {
	var configPath string
	var runAddr, filePath, db, key, cryptoKey, trustedSubnet, auditFile, auditURL string
	var storeInt int64
	var restoreData bool

	fs := newFlagSet()
	fs.StringVar(&configPath, "c", "", "path to configuration file")
	fs.StringVar(&configPath, "config", "", "path to configuration file")
	fs.StringVar(&runAddr, "a", "localhost:8080", "server address")
	fs.Int64Var(&storeInt, "i", 300, "interval in seconds to save server metrics (0 = synchronous)")
	fs.StringVar(&filePath, "f", "metrics.json", "path to file where server metrics are stored")
	fs.BoolVar(&restoreData, "r", false, "load previously saved metrics on startup")
	fs.StringVar(&db, "d", "", "database DSN")
	fs.StringVar(&key, "k", "", "hash key")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to private key file")
	fs.StringVar(&trustedSubnet, "t", "", "trusted subnet in CIDR notation")
	fs.StringVar(&auditFile, "audit-file", "", "path to audit log file")
	fs.StringVar(&auditURL, "audit-url", "", "URL to send audit logs")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return &ServerFileConfig{
		Address:       &runAddr,
		StoreInterval: &storeInt,
		StoreFile:     &filePath,
		Restore:       &restoreData,
		DatabaseDSN:   &db,
		Key:           &key,
		CryptoKey:     &cryptoKey,
		TrustedSubnet: &trustedSubnet,
		AuditFile:     &auditFile,
		AuditURL:      &auditURL,
	}, nil
}
