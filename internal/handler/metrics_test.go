package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_Create(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "positive gauge",
			url:  "/update/gauge/TestGauge/123.45",
			want: want{
				statusCode:  200,
				contentType: "text/plain; charset=utf-8",
				response:    "",
			},
		},
		{
			name: "positive counter",
			url:  "/update/counter/TestCounter/10",
			want: want{
				statusCode:  200,
				contentType: "text/plain; charset=utf-8",
				response:    "",
			},
		},
		{
			name: "unknown metric type",
			url:  "/update/unknown/Test/10",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "Unknown metric type\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			memStore := storage.NewMemStorage()
			router := http.NewServeMux()

			// ВАЖНО: регистрируем маршрут с path-параметрами
			router.HandleFunc(
				"/update/{metrics_type}/{metrics_name}/{metrics_value}",
				(&handler.MetricsHandler{Storage: memStore}).Create(),
			)

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			result.Body.Close()

			assert.Equal(t, tt.want.response, string(bodyBytes))
		})
	}
}