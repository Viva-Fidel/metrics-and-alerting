// Package payload содержит структуры данных для API метрик
package payload

// MetricsURL представляет метрику, полученную из URL-параметров
type MetricsURL struct {
	Gauge *float64
	Count *int64
	Type  string
	Name  string
}

// MetricsJSON представляет метрику в JSON-формате API
type MetricsJSON struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}
