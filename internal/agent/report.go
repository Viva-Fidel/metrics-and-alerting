package agent

import (
	"fmt"
	"strconv"

	"resty.dev/v3"
)

func ReportMetrics(metrics *Metrics, addr string) {
	client := resty.New()
	defer client.Close()

	baseURL := fmt.Sprintf("http://%s", addr)

	// Отправка gauge
	for name, value := range metrics.Gauge {
		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(fmt.Sprintf(
				"%s/update/gauge/%s/%s",
				baseURL,
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
			SetHeader("Content-Type", "text/plain").
			Post(fmt.Sprintf(
				"%s/update/counter/%s/%d",
				baseURL,
				name,
				value,
			))
		if err != nil {
			continue
		}
	}
}