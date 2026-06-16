package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"testing"

	models "github.com/Viva-Fidel/metrics-and-alerting/internal/model"
)

func benchMetricsBatch(b *testing.B, size int) []models.Metrics {
	b.Helper()

	batch := make([]models.Metrics, size)
	for i := range size {
		if i%2 == 0 {
			value := float64(i) * 1.23
			batch[i] = models.Metrics{
				ID:    fmt.Sprintf("gauge_%d", i),
				MType: metricTypeGauge,
				Value: &value,
			}
			continue
		}

		delta := int64(i)
		batch[i] = models.Metrics{
			ID:    fmt.Sprintf("counter_%d", i),
			MType: metricTypeCounter,
			Delta: &delta,
		}
	}

	return batch
}

func BenchmarkSendBatchMetrics_Encode(b *testing.B) {
	sizes := []int{30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("batch=%d", size), func(b *testing.B) {
			batch := benchMetricsBatch(b, size)
			var compressed bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				b.StopTimer()
				compressed.Reset()
				zipWriter := gzip.NewWriter(&compressed)
				b.StartTimer()
				if err := json.NewEncoder(zipWriter).Encode(batch); err != nil {
					b.Fatal(err)
				}
				if err := zipWriter.Close(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
