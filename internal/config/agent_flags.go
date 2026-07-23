package config

import "os"

// AgentFlags содержит итоговые параметры запуска агента
type AgentFlags struct {
	RunAddr        string
	GRPCAddress    string
	Key            string
	CryptoKey      string
	ReportInterval int64
	PollInterval   int64
	RateLimit      int
}

// LoadAgentFlags загружает параметры запуска агента из файла, переменных окружения и флагов
func LoadAgentFlags() (*AgentFlags, error) {
	configPath, err := resolveConfigPath(os.Args[1:])
	if err != nil {
		return nil, err
	}

	fileLayer := &AgentFileConfig{}
	if err := loadJSONConfig(configPath, fileLayer); err != nil {
		return nil, err
	}

	envLayer, err := loadAgentEnvLayer()
	if err != nil {
		return nil, err
	}

	flagLayer, err := parseAgentFlagLayer()
	if err != nil {
		return nil, err
	}

	merged, err := mergeOverlays(fileLayer, envLayer, flagLayer)
	if err != nil {
		return nil, err
	}

	return agentFileConfigToFlags(merged), nil
}

func parseAgentFlagLayer() (*AgentFileConfig, error) {
	var configPath string
	var addr, grpcAddress string
	var reportInterval, pollInterval int64
	var rateLimit int
	var key, cryptoKey string

	fs := newFlagSet()
	fs.StringVar(&configPath, "c", "", "path to configuration file")
	fs.StringVar(&configPath, "config", "", "path to configuration file")
	fs.StringVar(&addr, "a", "localhost:8080", "server address")
	fs.StringVar(&grpcAddress, "grpc-address", "", "gRPC server address")
	fs.Int64Var(&reportInterval, "r", 10, "report sending interval")
	fs.Int64Var(&pollInterval, "p", 2, "report collecting interval")
	fs.IntVar(&rateLimit, "l", 1, "max number of concurrent outgoing requests")
	fs.StringVar(&key, "k", "", "hash key")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to public key file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return &AgentFileConfig{
		Address:        &addr,
		GRPCAddress:    &grpcAddress,
		ReportInterval: &reportInterval,
		PollInterval:   &pollInterval,
		RateLimit:      &rateLimit,
		Key:            &key,
		CryptoKey:      &cryptoKey,
	}, nil
}
