package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func resolveConfigPath(args []string) (string, error) {
	var fromFlags string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-c" || arg == "-config":
			if i+1 >= len(args) {
				return "", fmt.Errorf("flag %s requires value", arg)
			}
			fromFlags = strings.TrimSpace(args[i+1])
			i++
		case strings.HasPrefix(arg, "-c="):
			fromFlags = strings.TrimSpace(strings.TrimPrefix(arg, "-c="))
		case strings.HasPrefix(arg, "-config="):
			fromFlags = strings.TrimSpace(strings.TrimPrefix(arg, "-config="))
		}
	}

	if fromFlags != "" {
		return fromFlags, nil
	}

	if value, ok := os.LookupEnv("CONFIG"); ok {
		return strings.TrimSpace(value), nil
	}

	return "", nil
}

func loadJSONConfig(path string, cfg any) error {
	if path == "" {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}
	if err := json.Unmarshal(content, cfg); err != nil {
		return fmt.Errorf("decode config file %q: %w", path, err)
	}

	return nil
}

func envString(name string) *string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func envInt64(name string) (*int64, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	return &parsed, nil
}

func envInt(name string) (*int, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return nil, nil
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	return &parsed, nil
}

func envBool(name string) (*bool, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	return &parsed, nil
}
