package agent

import (
	"math/rand/v2"
	"runtime"
)

// PollMetrics собирает стандартные метрики runtime
func PollMetrics(metrics *Metrics) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics.SetGauge("Alloc", float64(m.Alloc))
	metrics.SetGauge("BuckHashSys", float64(m.BuckHashSys))
	metrics.SetGauge("Frees", float64(m.Frees))
	metrics.SetGauge("GCCPUFraction", m.GCCPUFraction)
	metrics.SetGauge("GCSys", float64(m.GCSys))
	metrics.SetGauge("HeapAlloc", float64(m.HeapAlloc))
	metrics.SetGauge("HeapIdle", float64(m.HeapIdle))
	metrics.SetGauge("HeapInuse", float64(m.HeapInuse))
	metrics.SetGauge("HeapObjects", float64(m.HeapObjects))
	metrics.SetGauge("HeapReleased", float64(m.HeapReleased))
	metrics.SetGauge("HeapSys", float64(m.HeapSys))
	metrics.SetGauge("LastGC", float64(m.LastGC))
	metrics.SetGauge("Lookups", float64(m.Lookups))
	metrics.SetGauge("MCacheInuse", float64(m.MCacheInuse))
	metrics.SetGauge("MCacheSys", float64(m.MCacheSys))
	metrics.SetGauge("MSpanInuse", float64(m.MSpanInuse))
	metrics.SetGauge("MSpanSys", float64(m.MSpanSys))
	metrics.SetGauge("Mallocs", float64(m.Mallocs))
	metrics.SetGauge("NextGC", float64(m.NextGC))
	metrics.SetGauge("NumForcedGC", float64(m.NumForcedGC))
	metrics.SetGauge("NumGC", float64(m.NumGC))
	metrics.SetGauge("OtherSys", float64(m.OtherSys))
	metrics.SetGauge("PauseTotalNs", float64(m.PauseTotalNs))
	metrics.SetGauge("StackInuse", float64(m.StackInuse))
	metrics.SetGauge("StackSys", float64(m.StackSys))
	metrics.SetGauge("Sys", float64(m.Sys))
	metrics.SetGauge("TotalAlloc", float64(m.TotalAlloc))

	// Дополнительные метрики
	metrics.IncCounter("PollCount")
	metrics.SetGauge("RandomValue", rand.Float64())
}
