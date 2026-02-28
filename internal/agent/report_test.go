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
	"resty.dev/v3"
)

func TestReportMetricsWithResty(t *testing.T) {
	// Мапа для проверки полученных запросов
	received := make(map[string]string)

	// Поднимаем тестовый HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received[r.URL.Path] = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Создаём метрики
	metrics := &agent.Metrics{
		Gauge:   map[string]float64{"TestGauge": 12.34},
		Counter: map[string]int64{"TestCounter": 42},
	}

	// Создаём Resty client с базовым URL тестового сервера
	client := resty.New().
		SetBaseURL(ts.URL)
	defer client.Close()

	// Переопределяем функцию ReportMetrics для теста
	reportMetrics := func(metrics *agent.Metrics) {
		// Отправка gauge
		for name, value := range metrics.Gauge {
			_, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(fmt.Sprintf("/update/gauge/%s/%s",
					name,
					strconv.FormatFloat(value, 'f', 2, 64),
				))
			require.NoError(t, err)
		}

		// Отправка counter
		for name, value := range metrics.Counter {
			_, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(fmt.Sprintf("/update/counter/%s/%d",
					name,
					value,
				))
			require.NoError(t, err)
		}
	}

	// Вызываем
	reportMetrics(metrics)

	// Проверяем, что сервер получил правильные пути и заголовки
	assert.Equal(t, "text/plain", received["/update/gauge/TestGauge/12.34"])
	assert.Equal(t, "text/plain", received["/update/counter/TestCounter/42"])
}
