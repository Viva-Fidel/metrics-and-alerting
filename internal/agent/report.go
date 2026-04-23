package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/model"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"resty.dev/v3"
)

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

// ReportMetrics отправляет собранные метрики на сервер
func ReportMetrics(ctx context.Context, client *resty.Client, metrics *Metrics, hashKey string) {
	gauge, counter := metrics.Snapshot()
	batch := make([]models.Metrics, 0, len(gauge)+len(counter))

	for name, value := range gauge {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}

	for name, value := range counter {
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

	if sendBatchMetrics(ctx, client, batch, hashKey) {
		return
	}

	sendMetricsLegacy(ctx, client, gauge, counter, hashKey)
}

// sendMetricsLegacy отправляет метрики по одному, используя старый формат REST-запросов.
func sendMetricsLegacy(
	ctx context.Context,
	client *resty.Client,
	gauge map[string]float64,
	counter map[string]int64,
	hashKey string,
) {
	for name, value := range gauge {
		err := withRetry(ctx, func() error {
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
			_, reqErr := req.Post(path)
			return reqErr
		}, isRetriableAgentError)
		if err != nil {
			continue
		}
	}

	for name, value := range counter {
		err := withRetry(ctx, func() error {
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
			_, reqErr := req.Post(path)
			return reqErr
		}, isRetriableAgentError)
		if err != nil {
			continue
		}
	}
}

// sendBatchMetrics отправляет пакет метрик в формате JSON c использованием gzip-сжатия.
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
	err := withRetry(ctx, func() error {
		var reqErr error
		req := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(compressed.Bytes())
		if hashKey != "" {
			req.SetHeader(security.HashHeader, security.BuildHash(compressed.Bytes(), hashKey))
		}
		resp, reqErr = req.Post("/updates/")
		return reqErr
	}, isRetriableAgentError)
	if err != nil {
		return false
	}

	return resp.IsSuccess()
}

// withRetry выполняет fn с повторными попытками, если ошибка подходит для повторения
func withRetry(ctx context.Context, fn func() error, canRetry func(error) bool) error {
	for attempt := 0; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if attempt >= len(retryDelays) || !canRetry(err) {
			return err
		}

		timer := time.NewTimer(retryDelays[attempt])
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// isRetriableAgentError определяет, является ли ошибка временной
func isRetriableAgentError(err error) bool {
	if err == nil {
		return false
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	type temporary interface {
		Temporary() bool
	}
	var tempErr temporary
	return errors.As(err, &tempErr) && tempErr.Temporary()
}
