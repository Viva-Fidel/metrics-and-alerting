package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/handler"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
)

func benchBatchPayload(size int) []payload.MetricsJSON {
	batch := make([]payload.MetricsJSON, size)
	for i := range size {
		if i%2 == 0 {
			value := float64(i) * 1.23
			batch[i] = payload.MetricsJSON{
				ID:    fmt.Sprintf("gauge_%d", i),
				MType: "gauge",
				Value: &value,
			}
			continue
		}

		delta := int64(i)
		batch[i] = payload.MetricsJSON{
			ID:    fmt.Sprintf("counter_%d", i),
			MType: "counter",
			Delta: &delta,
		}
	}

	return batch
}

func BenchmarkCreateMetricsFromJSONBatch(b *testing.B) {
	sizes := []int{30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("batch=%d", size), func(b *testing.B) {
			gin.SetMode(gin.TestMode)

			logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
			repo := repository.NewMemRepository(logger, "", 0, false)
			svc := service.NewMetricsService(repo)

			router := gin.New()
			handler.NewMetricsHandler(router, handler.MetricsHandlerDeps{
				MetricsService: svc,
				AuditPublisher: audit.NewPublisher(logger, "", ""),
			})

			body, err := json.Marshal(benchBatchPayload(size))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("unexpected status: %d", w.Code)
				}
			}
		})
	}
}
