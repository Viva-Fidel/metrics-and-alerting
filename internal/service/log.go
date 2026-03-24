package service

import (
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
)


type LogMetricsConfig struct {
	FilePath      string
	StoreInterval int64
	Restore       bool
}

type LogMetricsService struct {
	memRepo       *repository.MemRepository
	fileRepo      *repository.LogRepository
	storeInterval int64
	lastSaved     time.Time
}

func NewLogMetricsService(mem *repository.MemRepository, cfg LogMetricsConfig) *LogMetricsService {
	return &LogMetricsService{
		memRepo:       mem,
		fileRepo:      repository.NewLogRepository(cfg.FilePath),
		storeInterval: cfg.StoreInterval,
		lastSaved:     time.Now(),
	}
}

func (s *LogMetricsService) Save() error {
	gauges, counters := s.memRepo.GetAll()
	var metrics []payload.MetricsJSON

	for k, v := range gauges {
		val := v
		metrics = append(metrics, payload.MetricsJSON{
			ID:    k,
			MType: Gauge,
			Value: &val,
		})
	}

	for k, v := range counters {
		val := v
		metrics = append(metrics, payload.MetricsJSON{
			ID:    k,
			MType: Counter,
			Delta: &val,
		})
	}

	return s.fileRepo.Save(metrics)
}

func (s *LogMetricsService) Load() error {
	metrics, err := s.fileRepo.Load()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		switch m.MType {
		case Gauge:
			if m.Value != nil {
				s.memRepo.SetGauge(m.ID, *m.Value)
			}
		case Counter:
			if m.Delta != nil {
				s.memRepo.AddCounter(m.ID, *m.Delta)
			}
		}
	}
	return nil
}

func (s *LogMetricsService) SaveIfNeeded() error {
	if s.storeInterval == 0 {
		return s.Save()
	}
	if time.Since(s.lastSaved) >= time.Duration(s.storeInterval)*time.Second {
		err := s.Save()
		if err == nil {
			s.lastSaved = time.Now()
		}
		return err
	}
	return nil
}