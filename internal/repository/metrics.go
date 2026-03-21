package repository



type MemRepository struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemRepository() *MemRepository {
	return &MemRepository{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemRepository) SetGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemRepository) AddCounter(name string, value int64) {
	m.counters[name] += value
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