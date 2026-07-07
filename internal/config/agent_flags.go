package config

import "flag"

// AgentFlags содержит итоговые параметры запуска агента
type AgentFlags struct {
	RunAddr        string
	Key            string
	CryptoKey      string
	ReportInterval int64
	PollInterval   int64
	RateLimit      int
}

// LoadAgentFlags загружает параметры запуска агента из файла, флагов и переменных окружения
func LoadAgentFlags() (*AgentFlags, error) {
	configPath := resolveConfigPath()

	flags := &AgentFlags{
		RunAddr:        "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      1,
	}

	if configPath != "" {
		if err := applyAgentFileConfig(configPath, flags); err != nil {
			return nil, err
		}
	}

	configFlag := &configPathFlag{value: configPath}
	flag.Var(configFlag, "c", "path to JSON configuration file")
	flag.Var(configFlag, "config", "path to JSON configuration file")
	flag.StringVar(&flags.RunAddr, "a", flags.RunAddr, "server address")
	flag.Int64Var(&flags.ReportInterval, "r", flags.ReportInterval, "report sending interval")
	flag.Int64Var(&flags.PollInterval, "p", flags.PollInterval, "report collecting interval")
	flag.IntVar(&flags.RateLimit, "l", flags.RateLimit, "max number of concurrent outgoing requests")
	flag.StringVar(&flags.Key, "k", flags.Key, "hash key")
	flag.StringVar(&flags.CryptoKey, "crypto-key", flags.CryptoKey, "path to public key file")

	flag.Parse()

	conf, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if conf.Agent.Address != "" {
		flags.RunAddr = conf.Agent.Address
	}
	if conf.Agent.ReportInterval != 0 {
		flags.ReportInterval = conf.Agent.ReportInterval
	}
	if conf.Agent.PollInterval != 0 {
		flags.PollInterval = conf.Agent.PollInterval
	}
	if conf.Agent.RateLimit != 0 {
		flags.RateLimit = conf.Agent.RateLimit
	}
	if conf.Agent.Key != "" {
		flags.Key = conf.Agent.Key
	}
	if conf.Agent.CryptoKey != "" {
		flags.CryptoKey = conf.Agent.CryptoKey
	}

	return flags, nil
}
