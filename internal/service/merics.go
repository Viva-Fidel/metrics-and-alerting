package service

import (
	"errors"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsRepository interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)
	ApplyBatch(metrics []payload.MetricsJSON) error

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	GetAll() (map[string]float64, map[string]int64)
}

type MetricsService struct {
	MetricsRepository MetricsRepository
}

// NewMetricsService создаёт новый сервис метрик на основе переданного репозитория.
func NewMetricsService(metricsRepository MetricsRepository) *MetricsService {
	return &MetricsService{MetricsRepository: metricsRepository}
}

// SetMetricJSON устанавливает значение метрики на основе структуры payload.MetricsJSON.
func (s *MetricsService) SetMetricJSON(m *payload.MetricsJSON) error {
	switch m.MType {
	case Gauge:
		if m.Value == nil {
			return errors.New("invalid gauge value")
		}
		s.MetricsRepository.SetGauge(m.ID, *m.Value)
	case Counter:
		if m.Delta == nil {
			return errors.New("invalid counter value")
		}
		s.MetricsRepository.AddCounter(m.ID, *m.Delta)
	default:
		return errors.New("unknown metric type")
	}
	return nil
}

// SetMetricsJSONBatch выполняет пакетное обновление метрик из слайса payload.MetricsJSON.
func (s *MetricsService) SetMetricsJSONBatch(metrics []payload.MetricsJSON) error {
	if len(metrics) == 0 {
		return nil
	}
	for i := range metrics {
		switch metrics[i].MType {
		case Gauge:
			if metrics[i].Value == nil {
				return errors.New("invalid gauge value")
			}
		case Counter:
			if metrics[i].Delta == nil {
				return errors.New("invalid counter value")
			}
		default:
			return errors.New("unknown metric type")
		}
	}
	return s.MetricsRepository.ApplyBatch(metrics)
}

// SetMetricURL устанавливает значение метрики, переданное в виде строковых параметров (например, из URL).
func (s *MetricsService) SetMetricURL(metricType, name, value string) error {
	switch metricType {
	case Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.New("invalid gauge value")
		}
		s.MetricsRepository.SetGauge(name, v)
	case Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.New("invalid counter value")
		}
		s.MetricsRepository.AddCounter(name, v)
	default:
		return errors.New("unknown metric type")
	}
	return nil
}

// GetMetricURL возвращает метрику (структуру payload.MetricsURL) по типу и имени (например, для ответа на запрос по URL).
func (s *MetricsService) GetMetricURL(metricType, name string) (*payload.MetricsURL, error) {
	switch metricType {
	case Gauge:
		val, ok := s.MetricsRepository.GetGauge(name)
		if !ok {
			return nil, errors.New("metric not found")
		}
		return &payload.MetricsURL{
			Type:  Gauge,
			Name:  name,
			Gauge: &val,
		}, nil
	case Counter:
		val, ok := s.MetricsRepository.GetCounter(name)
		if !ok {
			return nil, errors.New("metric not found")
		}
		return &payload.MetricsURL{
			Type:  Counter,
			Name:  name,
			Count: &val,
		}, nil
	default:
		return nil, errors.New("unknown metric type")
	}
}

// GetMetricJSON возвращает метрику (структуру payload.MetricsJSON) по типу и имени (для API с JSON).
func (s *MetricsService) GetMetricJSON(metricType, name string) (*payload.MetricsJSON, error) {
	switch metricType {
	case Gauge:
		val, ok := s.MetricsRepository.GetGauge(name)
		if !ok {
			return nil, errors.New("metric not found")
		}
		return &payload.MetricsJSON{
			ID:    name,
			MType: Gauge,
			Value: &val,
		}, nil
	case Counter:
		val, ok := s.MetricsRepository.GetCounter(name)
		if !ok {
			return nil, errors.New("metric not found")
		}
		return &payload.MetricsJSON{
			ID:    name,
			MType: Counter,
			Delta: &val,
		}, nil
	default:
		return nil, errors.New("unknown metric type")
	}
}
