package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
)

func newExampleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := repository.NewMemRepository(logger, "", 0, false)
	svc := service.NewMetricsService(repo)
	router := gin.New()
	handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
		MetricsService: svc,
		AuditPublisher: audit.NewPublisher(logger, "", ""),
	})
	return router
}

func Example() {
	router := newExampleRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleMetricsHandler_CreateMetricFromURL() {
	router := newExampleRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleMetricsHandler_GetMetricFromURL() {
	router := newExampleRouter()

	createReq := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/42.5", nil)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// 42.5
}

func ExampleMetricsHandler_CreateMetricFromJSON() {
	router := newExampleRouter()

	body := `{"id":"Alloc","type":"gauge","value":42.5}`
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// {"status":"OK"}
}

func ExampleMetricsHandler_GetMetricFromJSON() {
	router := newExampleRouter()

	createBody := `{"id":"PollCount","type":"counter","delta":5}`
	createReq := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	getBody := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBufferString(getBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// {"delta":5,"id":"PollCount","type":"counter"}
}

func ExampleMetricsHandler_CreateMetricsFromJSONBatch() {
	router := newExampleRouter()

	body := `[
		{"id":"GaugeMetric","type":"gauge","value":100.5},
		{"id":"CounterMetric","type":"counter","delta":7}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), `"id":"GaugeMetric"`))
	// Output:
	// 200
	// true
}

func ExampleMetricsHandler_GetAllMetrics() {
	router := newExampleRouter()

	createReq := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/42.5", nil)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))
	fmt.Println(strings.Contains(w.Body.String(), "Alloc: 42.5"))
	// Output:
	// 200
	// text/html; charset=utf-8
	// true
}
