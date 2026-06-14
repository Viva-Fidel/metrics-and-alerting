// Package models содержит доменные типы метрик
package models

const (
	// Counter — тип метрики-счётчика
	Counter = "counter"
	// Gauge — тип метрики с плавающей точкой
	Gauge = "gauge"
)

// Metrics описывает метрику для сериализации в JSON
type Metrics struct {
	// ID — имя метрики
	ID string `json:"id"`
	// MType — тип метрики (gauge или counter)
	MType string `json:"type"`
	// Delta — значение counter-метрики
	Delta *int64 `json:"delta,omitempty"`
	// Value — значение gauge-метрики
	Value *float64 `json:"value,omitempty"`
}
