// Package models содержит доменные типы метрик.
package models

const (
	// Counter — тип метрики-счётчика.
	Counter = "counter"
	// Gauge — тип метрики с плавающей точкой.
	Gauge = "gauge"
)

// Metrics описывает метрику для сериализации в JSON.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
