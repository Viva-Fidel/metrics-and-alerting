package agent_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportMetrics(t *testing.T) {

	// Поднимаем тестовый HTTP сервер
	received := make(map[string]string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received[r.URL.Path] = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Создаём метрики
	metrics := agent.NewMetrics()
	metrics.Gauge["TestGauge"] = 12.34
	metrics.Counter["TestCounter"] = 42

	client := &http.Client{}

	// Отправляем gauge
	for name, value := range metrics.Gauge {
		url := fmt.Sprintf("%s/update/gauge/%s/%s",
			ts.URL,
			name,
			strconv.FormatFloat(value, 'f', 2, 64),
		)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Отправляем counter
	for name, value := range metrics.Counter {
		url := fmt.Sprintf("%s/update/counter/%s/%d",
			ts.URL,
			name,
			value,
		)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close() 

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Проверяем, что сервер получил правильные пути и заголовки
	assert.Equal(t, "text/plain", received["/update/gauge/TestGauge/12.34"])
	assert.Equal(t, "text/plain", received["/update/counter/TestCounter/42"])
}
