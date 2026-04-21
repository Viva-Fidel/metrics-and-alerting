package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/model"
	"resty.dev/v3"
)

func ReportMetrics(ctx context.Context, client *resty.Client, metrics *Metrics) {
	batch := make([]models.Metrics, 0, len(metrics.Gauge)+len(metrics.Counter))

	for name, value := range metrics.Gauge {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}

	for name, value := range metrics.Counter {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &v,
		})
	}

	if len(batch) == 0 {
		return
	}

	if sendBatchMetrics(ctx, client, batch) {
		return
	}

	sendMetricsLegacy(ctx, client, metrics)
}

func sendMetricsLegacy(ctx context.Context, client *resty.Client, metrics *Metrics) {
	for name, value := range metrics.Gauge {
		_, err := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "text/plain").
			Post(fmt.Sprintf(
				"/update/gauge/%s/%s",
				name,
				strconv.FormatFloat(value, 'f', -1, 64),
			))
		if err != nil {
			continue
		}
	}

	for name, value := range metrics.Counter {
		_, err := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "text/plain").
			Post(fmt.Sprintf(
				"/update/counter/%s/%d",
				name,
				value,
			))
		if err != nil {
			continue
		}
	}
}

func sendBatchMetrics(ctx context.Context, client *resty.Client, batch []models.Metrics) bool {
	var compressed bytes.Buffer
	zipWriter := gzip.NewWriter(&compressed)
	if err := json.NewEncoder(zipWriter).Encode(batch); err != nil {
		_ = zipWriter.Close()
		return false
	}
	if err := zipWriter.Close(); err != nil {
		return false
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressed.Bytes()).
		Post("/updates/")
	if err != nil {
		return false
	}

	return resp.IsSuccess()
}
