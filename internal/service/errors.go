package service

import "errors"

var (
	ErrMetricNotFound      = errors.New("metric not found")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrUnknownMetricType   = errors.New("unknown metric type")
)
