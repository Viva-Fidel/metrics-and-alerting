// Package agent реализует агент сбора и отправки метрик на сервер
package agent

import (
	"maps"
	"sync"
)

// Metrics хранит собранные gauge- и counter-метрики агента
type Metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
	mu      sync.RWMutex
}

// NewMetrics создаёт пустое хранилище метрик агента
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

	gauge := maps.Clone(m.Gauge)
	counter := maps.Clone(m.Counter)

	return gauge, counter
}
