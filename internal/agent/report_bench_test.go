package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"testing"

	models "github.com/Viva-Fidel/metrics-and-alerting/internal/model"
)

func benchMetricsBatch(size int) []models.Metrics {
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

func BenchmarkReportMetrics_BuildBatch(b *testing.B) {
	sizes := []int{30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size), func(b *testing.B) {
			metrics := &Metrics{
				Gauge:   make(map[string]float64, size/2),
				Counter: make(map[string]int64, size/2),
			}
			for i := range size / 2 {
				metrics.Gauge[fmt.Sprintf("gauge_%d", i)] = float64(i) * 1.23
				metrics.Counter[fmt.Sprintf("counter_%d", i)] = int64(i)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				gauge, counter := metrics.Snapshot()
				batch := make([]models.Metrics, 0, len(gauge)+len(counter))

				for name, value := range gauge {
					v := value
					batch = append(batch, models.Metrics{
						ID:    name,
						MType: metricTypeGauge,
						Value: &v,
					})
				}

				for name, value := range counter {
					v := value
					batch = append(batch, models.Metrics{
						ID:    name,
						MType: metricTypeCounter,
						Delta: &v,
					})
				}
				_ = batch
			}
		})
	}
}

func BenchmarkSendBatchMetrics_Encode(b *testing.B) {
	sizes := []int{30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("batch=%d", size), func(b *testing.B) {
			batch := benchMetricsBatch(size)
			var compressed bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				compressed.Reset()
				zipWriter := gzip.NewWriter(&compressed)
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
