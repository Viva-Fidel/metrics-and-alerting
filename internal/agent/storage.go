package agent

import "sync"

type Metrics struct {
	mu      sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

// SetGauge устанавливает значение метрики типа gauge
func (m *Metrics) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gauge[name] = value
}

// IncCounter увеличивает значение метрики типа counter
func (m *Metrics) IncCounter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Counter[name]++
}

// Snapshot создает снимок текущего состояния метрик
func (m *Metrics) Snapshot() (map[string]float64, map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gauge := make(map[string]float64, len(m.Gauge))
	for k, v := range m.Gauge {
		gauge[k] = v
	}

	counter := make(map[string]int64, len(m.Counter))
	for k, v := range m.Counter {
		counter[k] = v
	}

	return gauge, counter
}
