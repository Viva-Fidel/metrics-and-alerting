package agent

import (
	"context"
	"fmt"
	"strconv"

	"resty.dev/v3"
)

func ReportMetrics(ctx context.Context, client *resty.Client, metrics *Metrics) {

	// Отправка gauge
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

	// Отправка counter
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