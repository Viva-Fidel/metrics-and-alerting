package config

import "flag"

type AgentFlags struct {
	RunAddr        string
	ReportInterval int64
	PollInterval   int64
	Key            string
}

func LoadAgentFlags() (*AgentFlags, error) {
	conf, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	flags := &AgentFlags{}

	flag.StringVar(&flags.RunAddr, "a", "localhost:8080", "server address")
	flag.Int64Var(&flags.ReportInterval, "r", 10, "report sending interval")
	flag.Int64Var(&flags.PollInterval, "p", 2, "report collecting interval")
	flag.StringVar(&flags.Key, "k", "", "hash key")
	flag.Parse()

	if conf.Agent.Address != "" {
		flags.RunAddr = conf.Agent.Address
	}
	if conf.Agent.ReportInterval != 0 {
		flags.ReportInterval = conf.Agent.ReportInterval
	}
	if conf.Agent.PollInterval != 0 {
		flags.PollInterval = conf.Agent.PollInterval
	}
	if conf.Agent.Key != "" {
		flags.Key = conf.Agent.Key
	}

	return flags, nil
}
