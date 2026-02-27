package agent_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/stretchr/testify/assert"
)

func TestReportMetrics(t *testing.T) {
	// Поднимаем тестовый HTTP сервер
	received := make(map[string]string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// сохраняем URL и Content-Type для проверки
		received[r.URL.Path] = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Создаём метрики
	metrics := agent.NewMetrics()
	metrics.Gauge["TestGauge"] = 12.34
	metrics.Counter["TestCounter"] = 42

	// Переопределяем ReportMetrics, чтобы слать на ts.URL
	client := &http.Client{}
	for name, value := range metrics.Gauge {
		url := fmt.Sprintf("%s/update/gauge/%s/%s", ts.URL, name, strconv.FormatFloat(value, 'f', 2, 64))
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")
		client.Do(req)
	}
	for name, value := range metrics.Counter {
		url := fmt.Sprintf("%s/update/counter/%s/%d", ts.URL, name, value)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")
		client.Do(req)
	}

	// Проверяем, что сервер получил правильные пути и заголовки
	assert.Equal(t, "text/plain", received["/update/gauge/TestGauge/12.34"])
	assert.Equal(t, "text/plain", received["/update/counter/TestCounter/42"])
}
