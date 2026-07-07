package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type serverFileConfig struct {
	Address                *string `json:"address"`
	Restore                *bool   `json:"restore"`
	StoreInterval          *string `json:"store_interval"`
	StoreFile              *string `json:"store_file"`
	DatabaseDSN            *string `json:"database_dsn"`
	CryptoKey              *string `json:"crypto_key"`
	Key                    *string `json:"key"`
	AuditFile              *string `json:"audit_file"`
	AuditURL               *string `json:"audit_url"`
	DatabaseMaxOpenConns   *int    `json:"database_max_open_conns"`
	DatabaseMaxIdleConns   *int    `json:"database_max_idle_conns"`
	DatabaseConnMaxLifeSec *int64  `json:"database_conn_max_lifetime_sec"`
	DatabaseConnMaxIdleSec *int64  `json:"database_conn_max_idle_time_sec"`
}

type agentFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
}

type configPathFlag struct {
	value string
}

func (f *configPathFlag) String() string {
	return f.value
}

func (f *configPathFlag) Set(s string) error {
	f.value = s
	return nil
}

func resolveConfigPath() string {
	path := os.Getenv("CONFIG")
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "-c", arg == "-config":
			if i+1 < len(os.Args) {
				path = os.Args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-c="):
			path = strings.TrimPrefix(arg, "-c=")
		case strings.HasPrefix(arg, "-config="):
			path = strings.TrimPrefix(arg, "-config=")
		}
	}
	return path
}

func loadFileConfig(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}

	return nil
}

func parseDurationSeconds(raw string) (int64, error) {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", raw, err)
	}

	return int64(d.Seconds()), nil
}

func applyServerFileConfig(path string, flags *ServerFlags) error {
	var fileCfg serverFileConfig
	if err := loadFileConfig(path, &fileCfg); err != nil {
		return err
	}

	if fileCfg.Address != nil {
		flags.RunAddr = *fileCfg.Address
	}
	if fileCfg.Restore != nil {
		flags.RestoreData = *fileCfg.Restore
	}
	if fileCfg.StoreInterval != nil {
		storeInterval, err := parseDurationSeconds(*fileCfg.StoreInterval)
		if err != nil {
			return err
		}
		flags.StoreInt = storeInterval
	}
	if fileCfg.StoreFile != nil {
		flags.FilePath = *fileCfg.StoreFile
	}
	if fileCfg.DatabaseDSN != nil {
		flags.DB = *fileCfg.DatabaseDSN
	}
	if fileCfg.CryptoKey != nil {
		flags.CryptoKey = *fileCfg.CryptoKey
	}
	if fileCfg.Key != nil {
		flags.Key = *fileCfg.Key
	}
	if fileCfg.AuditFile != nil {
		flags.AuditFile = *fileCfg.AuditFile
	}
	if fileCfg.AuditURL != nil {
		flags.AuditURL = *fileCfg.AuditURL
	}
	if fileCfg.DatabaseMaxOpenConns != nil {
		flags.DBMaxOpenConns = *fileCfg.DatabaseMaxOpenConns
	}
	if fileCfg.DatabaseMaxIdleConns != nil {
		flags.DBMaxIdleConns = *fileCfg.DatabaseMaxIdleConns
	}
	if fileCfg.DatabaseConnMaxLifeSec != nil {
		flags.DBConnMaxLifetimeSec = *fileCfg.DatabaseConnMaxLifeSec
	}
	if fileCfg.DatabaseConnMaxIdleSec != nil {
		flags.DBConnMaxIdleTimeSec = *fileCfg.DatabaseConnMaxIdleSec
	}

	return nil
}

func applyAgentFileConfig(path string, flags *AgentFlags) error {
	var fileCfg agentFileConfig
	if err := loadFileConfig(path, &fileCfg); err != nil {
		return err
	}

	if fileCfg.Address != nil {
		flags.RunAddr = *fileCfg.Address
	}
	if fileCfg.ReportInterval != nil {
		reportInterval, err := parseDurationSeconds(*fileCfg.ReportInterval)
		if err != nil {
			return err
		}
		flags.ReportInterval = reportInterval
	}
	if fileCfg.PollInterval != nil {
		pollInterval, err := parseDurationSeconds(*fileCfg.PollInterval)
		if err != nil {
			return err
		}
		flags.PollInterval = pollInterval
	}
	if fileCfg.CryptoKey != nil {
		flags.CryptoKey = *fileCfg.CryptoKey
	}
	if fileCfg.Key != nil {
		flags.Key = *fileCfg.Key
	}
	if fileCfg.RateLimit != nil {
		flags.RateLimit = *fileCfg.RateLimit
	}

	return nil
}
