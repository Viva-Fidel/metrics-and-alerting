package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	models "github.com/Viva-Fidel/metrics-and-alerting/internal/model"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"resty.dev/v3"
)

const metricTypeGauge = "gauge"
const metricTypeCounter = "counter"

// ReportMetrics отправляет собранные метрики на сервер
func ReportMetrics(ctx context.Context, client *resty.Client, metrics *Metrics, hashKey string) {
	gauge, counter := metrics.Snapshot()
	batch := make([]models.Metrics, 0, len(gauge)+len(counter))

	for name, value := range gauge {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: metricTypeGauge,
			Value: &v,
		})
	}

	for name, value := range counter {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: metricTypeCounter,
			Delta: &v,
		})
	}

	if len(batch) == 0 {
		return
	}

	if sendBatchMetrics(ctx, client, batch, hashKey) {
		return
	}

	sendMetricsLegacy(ctx, client, gauge, counter, hashKey)
}

// sendMetricsLegacy отправляет метрики по одному, используя старый формат REST-запросов
func sendMetricsLegacy(
	ctx context.Context,
	client *resty.Client,
	gauge map[string]float64,
	counter map[string]int64,
	hashKey string,
) {
	for name, value := range gauge {
		req := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "text/plain")
		path := fmt.Sprintf(
			"/update/gauge/%s/%s",
			name,
			strconv.FormatFloat(value, 'f', -1, 64),
		)
		if hashKey != "" {
			req.SetHeader(security.HashHeader, security.BuildHash(nil, hashKey))
		}
		_, err := req.Post(path)
		if err != nil {
			continue
		}
	}

	for name, value := range counter {
		req := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "text/plain")
		path := fmt.Sprintf(
			"/update/counter/%s/%d",
			name,
			value,
		)
		if hashKey != "" {
			req.SetHeader(security.HashHeader, security.BuildHash(nil, hashKey))
		}
		_, err := req.Post(path)
		if err != nil {
			continue
		}
	}
}

// sendBatchMetrics отправляет пакет метрик в формате JSON c использованием gzip-сжатия
func sendBatchMetrics(ctx context.Context, client *resty.Client, batch []models.Metrics, hashKey string) bool {
	var compressed bytes.Buffer
	zipWriter := gzip.NewWriter(&compressed)
	if err := json.NewEncoder(zipWriter).Encode(batch); err != nil {
		_ = zipWriter.Close()
		return false
	}
	if err := zipWriter.Close(); err != nil {
		return false
	}

	var resp *resty.Response
	req := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressed.Bytes())
	if hashKey != "" {
		req.SetHeader(security.HashHeader, security.BuildHash(compressed.Bytes(), hashKey))
	}
	resp, err := req.Post("/updates")
	if err != nil {
		return false
	}

	return resp.IsSuccess()
}
