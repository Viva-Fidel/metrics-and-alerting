package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type storedMetric struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MemRepository struct {
	mu sync.RWMutex

	gauges   map[string]float64
	counters map[string]int64

	filePath      string
	storeInterval time.Duration

	saveMu sync.Mutex

	ticker *time.Ticker
	stopCh chan struct{}
}

func NewMemRepository() *MemRepository {
	return &MemRepository{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemRepository) SetGauge(name string, value float64) {
	m.mu.Lock()
	m.gauges[name] = value
	m.mu.Unlock()

	if m.storeInterval == 0 {
		_ = m.SaveToFile()
	}
}

func (m *MemRepository) AddCounter(name string, value int64) {
	m.mu.Lock()
	m.counters[name] += value
	m.mu.Unlock()

	if m.storeInterval == 0 {
		_ = m.SaveToFile()
	}
}

func (m *MemRepository) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.gauges[name]
	return val, ok
}

func (m *MemRepository) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.counters[name]
	return val, ok
}

func (m *MemRepository) GetAll() (map[string]float64, map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gaugesCopy := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}

	countersCopy := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}

func (m *MemRepository) ConfigureStorage(filePath string, storeIntervalSeconds int64) {
	m.filePath = filePath
	if storeIntervalSeconds <= 0 {
		m.storeInterval = 0
		return
	}

	m.storeInterval = time.Duration(storeIntervalSeconds) * time.Second
}

func (m *MemRepository) StartSaver() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.filePath == "" || m.storeInterval <= 0 {
		return
	}
	if m.ticker != nil {
		return
	}

	m.stopCh = make(chan struct{})
	m.ticker = time.NewTicker(m.storeInterval)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				_ = m.SaveToFile()
			case <-m.stopCh:
				m.ticker.Stop()
				return
			}
		}
	}()
}

func (m *MemRepository) LoadFromFile() error {
	if m.filePath == "" {
		return nil
	}

	raw, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var items []storedMetric
	if err := json.Unmarshal(raw, &items); err != nil {
		return err
	}

	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	for _, item := range items {
		switch item.Type {
		case "gauge":
			if item.Value != nil {
				gauges[item.ID] = *item.Value
			}
		case "counter":
			if item.Delta != nil {
				counters[item.ID] = *item.Delta
			}
		}
	}

	m.mu.Lock()
	m.gauges = gauges
	m.counters = counters
	m.mu.Unlock()
	return nil
}

func (m *MemRepository) SaveToFile() error {
	if m.filePath == "" {
		return nil
	}

	m.saveMu.Lock()
	defer m.saveMu.Unlock()

	m.mu.RLock()
	gaugesCopy := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}
	countersCopy := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		countersCopy[k] = v
	}
	m.mu.RUnlock()

	items := make([]storedMetric, 0, len(gaugesCopy)+len(countersCopy))
	for name, val := range gaugesCopy {
		v := val
		items = append(items, storedMetric{
			ID:    name,
			Type:  "gauge",
			Value: &v,
		})
	}
	for name, val := range countersCopy {
		v := val
		items = append(items, storedMetric{
			ID:    name,
			Type:  "counter",
			Delta: &v,
		})
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.filePath)
	base := filepath.Base(m.filePath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, m.filePath); err != nil {
		return err
	}
	return nil
}