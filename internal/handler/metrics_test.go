package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler_CreateMetricFromURL(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "positive gauge",
			url:  "/update/gauge/TestGauge/123.45",
			want: want{statusCode: 200},
		},
		{
			name: "positive counter",
			url:  "/update/counter/TestCounter/10",
			want: want{statusCode: 200},
		},
		{
			name: "unknown metric type",
			url:  "/update/unknown/Test/10",
			want: want{statusCode: 400},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			repo := repository.NewMemRepository()
			svc := service.NewMetricsService(repo)

			router := gin.New()
			handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
				MetricsService: svc,
			})

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}


func TestMetricsHandler_GetMetricFromURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemRepository()
	repo.SetGauge("TestGauge", 123.45)
	repo.AddCounter("TestCounter", 10)

	svc := service.NewMetricsService(repo)

	router := gin.New()
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: svc,
	})

	tests := []struct {
		name         string
		url          string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "existing gauge",
			url:          "/value/gauge/TestGauge",
			wantStatus:   200,
			wantContains: "123.45",
		},
		{
			name:         "existing counter",
			url:          "/value/counter/TestCounter",
			wantStatus:   200,
			wantContains: "10",
		},
		{
			name:         "non-existent metric",
			url:          "/value/gauge/NotExist",
			wantStatus:   404,
			wantContains: "metric not found",
		},
		{
			name:         "unknown metric type",
			url:          "/value/unknown/Test",
			wantStatus:   404,
			wantContains: "unknown metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, strings.ToLower(w.Body.String()), tt.wantContains)
		})
	}
}

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemRepository()
	repo.SetGauge("TestGauge", 123.45)
	repo.SetGauge("TestGauge2", 67.89)
	repo.AddCounter("TestCounter", 100)
	repo.AddCounter("TestCounter", 50)

	svc := service.NewMetricsService(repo)

	router := gin.New()
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: svc,
	})

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


func TestMetricsHandler_CreateMetricFromJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemRepository()
	svc := service.NewMetricsService(repo)

	router := gin.New()
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: svc,
	})

	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantResponse string
	}{
		{
			name:         "valid gauge",
			body:         `{"metrics_type":"gauge","metrics_name":"TestGauge","metrics_value":"123.45"}`,
			wantStatus:   http.StatusOK,
			wantResponse: `{"status":"OK"}`,
		},
		{
			name:         "valid counter",
			body:         `{"metrics_type":"counter","metrics_name":"TestCounter","metrics_value":"10"}`,
			wantStatus:   http.StatusOK,
			wantResponse: `{"status":"OK"}`,
		},
		{
			name:         "invalid JSON",
			body:         `{"metrics_type":}`,
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"error":"invalid JSON body"}`, 
		},
		{
			name:         "unknown metric type",
			body:         `{"metrics_type":"unknown","metrics_name":"Test","metrics_value":"10"}`,
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"error":"unknown metric type"}`, 
		},
		{
			name:         "invalid gauge value",
			body:         `{"metrics_type":"gauge","metrics_name":"TestGauge","metrics_value":"abc"}`,
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"error":"invalid gauge value"}`, 
		},
		{
			name:         "invalid counter value",
			body:         `{"metrics_type":"counter","metrics_name":"TestCounter","metrics_value":"abc"}`,
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"error":"invalid counter value"}`, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantResponse, w.Body.String())
		})
	}
}

func TestMetricsHandler_GetMetricFromJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemRepository()
	repo.SetGauge("TestGauge", 123.45)
	repo.AddCounter("TestCounter", 10)

	svc := service.NewMetricsService(repo)

	router := gin.New()
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: svc,
	})


	tests := []struct {
		name          string
		body          string
		wantStatus    int
		expectedJSON  string 
	}{
		{
			name:       "existing gauge",
			body:       `{"id":"TestGauge","type":"gauge"}`,
			wantStatus: http.StatusOK,
			expectedJSON: `{
				"id":"TestGauge",
				"type":"gauge",
				"value":123.45
			}`,
		},
		{
			name:       "existing counter",
			body:       `{"id":"TestCounter","type":"counter"}`,
			wantStatus: http.StatusOK,
			expectedJSON: `{
				"id":"TestCounter",
				"type":"counter",
				"delta":10
			}`,
		},
		{
			name:       "not found",
			body:       `{"id":"Unknown","type":"gauge"}`,
			wantStatus: http.StatusNotFound,
			expectedJSON: `{"error":"metric not found"}`,
		},
		{
			name:       "invalid json",
			body:       `{"id":}`,
			wantStatus: http.StatusBadRequest,
			expectedJSON: `{"error":"invalid JSON body"}`,
		},
		{
			name:       "unknown type",
			body:       `{"id":"Test","type":"unknown"}`,
			wantStatus: http.StatusNotFound,
			expectedJSON: `{"error":"unknown metric type"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/value/",
				bytes.NewBufferString(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}