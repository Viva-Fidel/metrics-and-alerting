// Package payload содержит структуры данных для API метрик
package payload

// MetricsURL представляет метрику, полученную из URL-параметров
type MetricsURL struct {
	// Type — тип метрики (gauge или counter)
	Type string
	// Name — имя метрики
	Name string
	// Gauge — значение gauge-метрики
	Gauge *float64
	// Count — значение counter-метрики
	Count *int64
}

// MetricsJSON представляет метрику в JSON-формате API
type MetricsJSON struct {
	// ID — имя метрики
	ID string `json:"id"`
	// MType — тип метрики (gauge или counter)
	MType string `json:"type"`
	// Delta — значение counter-метрики
	Delta *int64 `json:"delta,omitempty"`
	// Value — значение gauge-метрики
	Value *float64 `json:"value,omitempty"`
}
