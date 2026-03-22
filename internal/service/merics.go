package service

import (
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/domain"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsRepository interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	GetAll() (map[string]float64, map[string]int64)
}

type MetricsService struct {
	MetricsRepository *repository.MemRepository
}

func NewMetricsService(metricsRepository *repository.MemRepository) *MetricsService {
	return &MetricsService{MetricsRepository: metricsRepository}
}


func (s *MetricsService) SetMetricJson(m *payload.Metrics) error {
	switch m.MType {
	case Gauge:
		if m.Value == nil {
			return domain.ErrInvalidGaugeValue
		}
		s.MetricsRepository.SetGauge(m.ID, *m.Value)

	case Counter:
		if m.Delta == nil {
			return domain.ErrInvalidCounterValue
		}
		s.MetricsRepository.AddCounter(m.ID, *m.Delta)

	default:
		return domain.ErrUnknownMetricType
	}

	return nil
}

func (s *MetricsService) SetMetricUrl(metricType, name, value string) error {
	switch metricType {
	case Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return domain.ErrInvalidGaugeValue
		}
		s.MetricsRepository.SetGauge(name, v)

	case Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return domain.ErrInvalidCounterValue
		}
		s.MetricsRepository.AddCounter(name, v)

	default:
		return domain.ErrUnknownMetricType
	}
	return nil
}

func (s *MetricsService) GetMetricUrl(metricType, name string) (*payload.Metric, error) {
	switch metricType {
	case Gauge:
		val, ok := s.MetricsRepository.GetGauge(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metric{
			Type:  Gauge,
			Name:  name,
			Gauge: &val,
		}, nil

	case Counter:
		val, ok := s.MetricsRepository.GetCounter(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metric{
			Type:  Counter,
			Name:  name,
			Count: &val,
		}, nil

	default:
		return nil, domain.ErrUnknownMetricType
	}
}

func (s *MetricsService) GetMetricJson(metricType, name string) (*payload.Metrics, error) {
	switch metricType {
	case Gauge:
		val, ok := s.MetricsRepository.GetGauge(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metrics{
			ID:    name,
			MType: Gauge,
			Value: &val,
		}, nil

	case Counter:
		val, ok := s.MetricsRepository.GetCounter(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metrics{
			ID:    name,
			MType: Counter,
			Delta: &val,
		}, nil

	default:
		return nil, domain.ErrUnknownMetricType
	}
}