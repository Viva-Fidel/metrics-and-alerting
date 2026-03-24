package repository

import (
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
    f, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        return err
    }
    defer f.Close()

    for _, m := range metrics {
        data, err := json.Marshal(m)
        if err != nil {
            return err
        }
        if _, err := f.Write(append(data, '\n')); err != nil {
            return err
        }
    }

    return nil
}

func (r *LogRepository) Load() ([]payload.MetricsJSON, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var metrics []payload.MetricsJSON
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}