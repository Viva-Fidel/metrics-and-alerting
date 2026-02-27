package agent

import (
	"fmt"
	"net/http"
	"strconv"
)

func ReportMetrics(metrics *Metrics) {
	client := &http.Client{}

	// Отправка gauge
	for name, value := range metrics.Gauge {
		url := fmt.Sprintf(
			"http://localhost:8080/update/gauge/%s/%s",
			name,
			strconv.FormatFloat(value, 'f', 2, 64),
		)

		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")
		client.Do(req)
	}

	// Отправка counter
	for name, value := range metrics.Counter {
		url := fmt.Sprintf(
			"http://localhost:8080/update/counter/%s/%d",
			name,
			value,
		)

		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")
		client.Do(req)
	}
}