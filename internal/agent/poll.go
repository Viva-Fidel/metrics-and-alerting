package agent

import (
	"math/rand/v2"
	"runtime"
)

func PollMetrics(metrics *Metrics) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics.Gauge["Alloc"] = float64(m.Alloc)
	metrics.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	metrics.Gauge["Frees"] = float64(m.Frees)
	metrics.Gauge["GCCPUFraction"] = m.GCCPUFraction
	metrics.Gauge["GCSys"] = float64(m.GCSys)
	metrics.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	metrics.Gauge["HeapIdle"] = float64(m.HeapIdle)
	metrics.Gauge["HeapInuse"] = float64(m.HeapInuse)
	metrics.Gauge["HeapObjects"] = float64(m.HeapObjects)
	metrics.Gauge["HeapReleased"] = float64(m.HeapReleased)
	metrics.Gauge["HeapSys"] = float64(m.HeapSys)
	metrics.Gauge["LastGC"] = float64(m.LastGC)
	metrics.Gauge["Lookups"] = float64(m.Lookups)
	metrics.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	metrics.Gauge["MCacheSys"] = float64(m.MCacheSys)
	metrics.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	metrics.Gauge["MSpanSys"] = float64(m.MSpanSys)
	metrics.Gauge["Mallocs"] = float64(m.Mallocs)
	metrics.Gauge["NextGC"] = float64(m.NextGC)
	metrics.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	metrics.Gauge["NumGC"] = float64(m.NumGC)
	metrics.Gauge["OtherSys"] = float64(m.OtherSys)
	metrics.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	metrics.Gauge["StackInuse"] = float64(m.StackInuse)
	metrics.Gauge["StackSys"] = float64(m.StackSys)
	metrics.Gauge["Sys"] = float64(m.Sys)
	metrics.Gauge["TotalAlloc"] = float64(m.TotalAlloc)

	// Дополнительные метрики
	metrics.Counter["PollCount"]++
	metrics.Gauge["RandomValue"] = rand.Float64()
}