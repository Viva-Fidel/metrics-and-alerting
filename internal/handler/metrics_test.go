package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
				response:    "Unknown metric type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			memStore := repository.NewMemStorage()
			handler := handler.MetricsHandler{MemStorage: memStore}

			router.POST("/update/:metrics_type/:metrics_name/:metrics_value", handler.Create())

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.response, strings.TrimSpace(w.Body.String()))
		})
	}
}


func TestMetricsHandler_GetMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memStore := repository.NewMemStorage()
	memStore.SetGauge("TestGauge", 123.45)
	memStore.AddCounter("TestCounter", 10)

	h := handler.MetricsHandler{MemStorage: memStore}

	router := gin.New()
	router.GET("/value/:metrics_type/:metrics_name", h.GetMetric())

	tests := []struct {
		name         string
		url          string
		wantStatus   int
		wantResponse string
	}{
		{
			name:         "existing gauge",
			url:          "/value/gauge/TestGauge",
			wantStatus:   200,
			wantResponse: "123.45",
		},
		{
			name:         "existing counter",
			url:          "/value/counter/TestCounter",
			wantStatus:   200,
			wantResponse: "10",
		},
		{
			name:         "non-existent metric",
			url:          "/value/gauge/NotExist",
			wantStatus:   404,
			wantResponse: "Metric not found",
		},
		{
			name:         "unknown metric type",
			url:          "/value/unknown/Test",
			wantStatus:   404,
			wantResponse: "Unknown metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.wantResponse, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memStore := repository.NewMemStorage()
	memStore.SetGauge("TestGauge", 123.45)
	memStore.SetGauge("TestGauge2", 67.89)
	memStore.AddCounter("TestCounter", 100)
	memStore.AddCounter("TestCounter", 50)

	h := handler.MetricsHandler{MemStorage: memStore}

	router := gin.New()
	router.GET("/", h.GetAllMetrics())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "TestGauge: 123.45")
	assert.Contains(t, body, "TestGauge2: 67.89")
	assert.Contains(t, body, "TestCounter: 150")
}