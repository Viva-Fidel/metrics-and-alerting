package agent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/pkg/compress"
	"resty.dev/v3"
)

func ReportMetrics(ctx context.Context, client *resty.Client, metrics *Metrics) {
	// Отправка gauge
	for name, value := range metrics.Gauge {
		payload := strconv.FormatFloat(value, 'f', -1, 64)
		gzipped, err := compress.CompressBytes([]byte(payload))
		if err != nil {
			continue
		}

		_, err = client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/octet-stream").
			SetBody(gzipped).
			Post(fmt.Sprintf("/update/gauge/%s", name))
		if err != nil {
			continue
		}
	}

	// Отправка counter
	for name, value := range metrics.Counter {
		payload := strconv.FormatInt(value, 10)
		gzipped, err := compress.CompressBytes([]byte(payload))
		if err != nil {
			continue
		}

		_, err = client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/octet-stream").
			SetBody(gzipped).
			Post(fmt.Sprintf("/update/counter/%s", name))
		if err != nil {
			continue
		}
	}
}