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

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		resp.Body.Close() 
	}

	// Отправка counter
	for name, value := range metrics.Counter {
		url := fmt.Sprintf(
			"http://localhost:8080/update/counter/%s/%d",
			name,
			value,
		)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		resp.Body.Close() 
	}
}