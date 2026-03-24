package repository

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
)

type LogRepository struct {
	filePath string
}

func NewLogRepository(path string) *LogRepository {
	return &LogRepository{filePath: path}
}

func (r *LogRepository) Save(metrics []payload.MetricsJSON) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0o644)
}

func (r *LogRepository) Load() ([]payload.MetricsJSON, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var metrics []payload.MetricsJSON
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}