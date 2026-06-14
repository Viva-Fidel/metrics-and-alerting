package repository

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
)

func benchLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func benchMetrics(gaugeCount, counterCount int) (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64, gaugeCount)
	for i := range gaugeCount {
		gauges[fmt.Sprintf("gauge_%d", i)] = float64(i) * 1.23
	}

	counters := make(map[string]int64, counterCount)
	for i := range counterCount {
		counters[fmt.Sprintf("counter_%d", i)] = int64(i)
	}

	return gauges, counters
}

func newBenchRepo(b *testing.B, filePath string, gaugeCount, counterCount int) *MemRepository {
	b.Helper()

	repo := NewMemRepository(benchLogger(), filePath, 0, false)
	gauges, counters := benchMetrics(gaugeCount, counterCount)

	repo.mu.Lock()
	repo.gauges = gauges
	repo.counters = counters
	repo.mu.Unlock()

	return repo
}

func benchBatch(size int) []payload.MetricsJSON {
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

func BenchmarkWriteMetricsJSON(b *testing.B) {
	sizes := []int{30, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size*2), func(b *testing.B) {
			gauges, counters := benchMetrics(size, size)
			var buf bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := writeMetricsJSON(&buf, gauges, counters); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMemRepository_SaveToFile(b *testing.B) {
	sizes := []int{30, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size*2), func(b *testing.B) {
			dir := b.TempDir()
			repo := newBenchRepo(b, filepath.Join(dir, "metrics.json"), size, size)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := repo.SaveToFile(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMemRepository_LoadFromFile(b *testing.B) {
	sizes := []int{30, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size*2), func(b *testing.B) {
			dir := b.TempDir()
			filePath := filepath.Join(dir, "metrics.json")
			repo := newBenchRepo(b, filePath, size, size)
			if err := repo.SaveToFile(); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				target := NewMemRepository(benchLogger(), filePath, 0, false)
				if err := target.loadFromFile(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMemRepository_SetGauge(b *testing.B) {
	repo := NewMemRepository(benchLogger(), "", 0, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		repo.SetGauge(fmt.Sprintf("gauge_%d", i), float64(i))
	}
}

func BenchmarkMemRepository_ApplyBatch(b *testing.B) {
	sizes := []int{30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("batch=%d", size), func(b *testing.B) {
			repo := NewMemRepository(benchLogger(), "", 0, false)
			batch := benchBatch(size)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := repo.ApplyBatch(batch); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMemRepository_GetAll(b *testing.B) {
	sizes := []int{30, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size*2), func(b *testing.B) {
			repo := newBenchRepo(b, "", size, size)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = repo.GetAll()
			}
		})
	}
}
