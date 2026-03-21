package domain

import "errors"

var (
	// Ошибки по метрикам
	ErrMetricNotFound      = errors.New("metric not found") // Метрика не найдена
	ErrInvalidGaugeValue   = errors.New("invalid gauge value") // Неправильное значение gauge
	ErrInvalidCounterValue = errors.New("invalid counter value") // Неправильное значение counter
	ErrUnknownMetricType   = errors.New("unknown metric type") // Неправильный тип метрики
	ErrInvalidJSONBody = errors.New("invalid JSON body") // Неправлильные значения в JSON
)