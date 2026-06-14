package audit

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
)

func benchLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func BenchmarkPublisher_Notify(b *testing.B) {
	sizes := []int{1, 30, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("metrics=%d", size), func(b *testing.B) {
			dir := b.TempDir()
			publisher := NewPublisher(
				benchLogger(),
				filepath.Join(dir, "audit.json"),
				"",
			)
			defer publisher.Close()

			metrics := make([]string, size)
			for i := range size {
				metrics[i] = fmt.Sprintf("metric_%d", i)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				publisher.Notify(metrics, "127.0.0.1")
			}
		})
	}
}

func BenchmarkFileObserver_Update(b *testing.B) {
	dir := b.TempDir()
	observer := NewFileObserver(benchLogger(), filepath.Join(dir, "audit.json"))
	defer observer.Close()

	event := Event{
		TS:        1_700_000_000,
		Metrics:   []string{"Alloc", "HeapAlloc", "Sys"},
		IPAddress: "127.0.0.1",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		observer.Update(event)
	}
}
