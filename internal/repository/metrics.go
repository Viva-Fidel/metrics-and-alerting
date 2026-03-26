package repository

import (
	"encoding/json"
	"os"
	"time"
)

type MemRepository struct {
	gauges   map[string]float64
	counters map[string]int64

	filePath      string
	storeInterval time.Duration
	lastSave      time.Time
}

func NewMemRepository() *MemRepository {
	return &MemRepository{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemRepository) SetGauge(name string, value float64) {
	m.gauges[name] = value
	_ = m.SaveIfNeeded()
}

func (m *MemRepository) AddCounter(name string, value int64) {
	m.counters[name] += value
	_ = m.SaveIfNeeded()
}

func (m *MemRepository) GetGauge(name string) (float64, bool) {
	val, ok := m.gauges[name]
	return val, ok
}

func (m *MemRepository) GetCounter(name string) (int64, bool) {
	val, ok := m.counters[name]
	return val, ok
}

func (m *MemRepository) GetAll() (map[string]float64, map[string]int64) {
	return m.gauges, m.counters
}

func (m *MemRepository) ConfigureStorage(filePath string, storeIntervalSeconds int64) {
	m.filePath = filePath
	if storeIntervalSeconds <= 0 {
		m.storeInterval = 0
		m.lastSave = time.Time{}
		return
	}

	m.storeInterval = time.Duration(storeIntervalSeconds) * time.Second
	m.lastSave = time.Now()
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

	var dump struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	if err := json.Unmarshal(raw, &dump); err != nil {
		return err
	}

	if dump.Gauges == nil {
		dump.Gauges = make(map[string]float64)
	}
	if dump.Counters == nil {
		dump.Counters = make(map[string]int64)
	}

	m.gauges = dump.Gauges
	m.counters = dump.Counters
	return nil
}

func (m *MemRepository) SaveToFile() error {
	if m.filePath == "" {
		return nil
	}

	dump := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   m.gauges,
		Counters: m.counters,
	}

	raw, err := json.Marshal(dump)
	if err != nil {
		return err
	}

	if err := os.WriteFile(m.filePath, raw, 0o644); err != nil {
		return err
	}

	m.lastSave = time.Now()
	return nil
}

func (m *MemRepository) SaveIfNeeded() error {
	if m.filePath == "" {
		return nil
	}

	if m.storeInterval == 0 {
		return m.SaveToFile()
	}

	if time.Since(m.lastSave) >= m.storeInterval {
		return m.SaveToFile()
	}

	return nil
}