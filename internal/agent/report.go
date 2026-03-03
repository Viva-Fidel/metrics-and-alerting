package agent

import (
	"fmt"
	"strconv"

	"resty.dev/v3"
)

func ReportMetrics(metrics *Metrics) {
	client := resty.New()
	defer client.Close()

	// Отправка gauge
	for name, value := range metrics.Gauge {
		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(fmt.Sprintf(
				"http://localhost:8080/update/gauge/%s/%s",
				name,
				strconv.FormatFloat(value, 'f', 2, 64),
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
				"http://localhost:8080/update/counter/%s/%d",
				name,
				value,
			))
		if err != nil {
			continue
		}
	}
}