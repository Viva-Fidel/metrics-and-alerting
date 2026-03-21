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


func (s *MetricsService) SetMetric(metricType, name, value string) error {
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

func (s *MetricsService) GetMetric(metricType, name string) (*payload.Metric, error) {
	switch metricType {
	case Gauge:
		val, ok := s.MetricsRepository.GetGauge(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metric{
			ID:    name,
			MType: Gauge,
			Value: &val,
		}, nil

	case Counter:
		val, ok := s.MetricsRepository.GetCounter(name)
		if !ok {
			return nil, domain.ErrMetricNotFound
		}
		return &payload.Metric{
			ID:    name,
			MType: Counter,
			Delta: &val,
		}, nil

	default:
		return nil, domain.ErrUnknownMetricType
	}
}