package service

import (
	"fmt"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
)

type MetricsService interface {
	UpdateMetric(metric models.Metrics) error
	UpdateGauge(name string, value float64) error
	GetGauge(name string) (float64, bool)
	GetAllGauges() (map[string]float64, error)
	UpdateCounter(name string, value int64) error
	GetCounter(name string) (int64, bool)
	GetAllCounters() (map[string]int64, error)
	SetLocalStorage(storage repository.Repository)
	PingDB() error
}

type Service struct {
	storage repository.Repository
}

func NewService(storage repository.Repository) MetricsService {
	return &Service{
		storage: storage,
	}
}

func (s *Service) UpdateMetric(metric models.Metrics) error {
	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("metric %q: Delta is nil", metric.ID)
		}
		return s.UpdateCounter(metric.ID, *metric.Delta)
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("metric %q: Value is nil", metric.ID)
		}
		return s.UpdateGauge(metric.ID, *metric.Value)
	}

	return nil
}

func (s *Service) UpdateGauge(name string, value float64) error {
	return s.storage.SetGauge(name, value)
}

func (s *Service) GetGauge(name string) (float64, bool) {
	return s.storage.GetGauge(name)
}

func (s *Service) GetAllGauges() (map[string]float64, error) {
	return s.storage.GetAllGauges()
}

func (s *Service) UpdateCounter(name string, value int64) error {
	return s.storage.SetCounter(name, value)
}

func (s *Service) GetCounter(name string) (int64, bool) {
	return s.storage.GetCounter(name)
}

func (s *Service) GetAllCounters() (map[string]int64, error) {
	return s.storage.GetAllCounters()
}

func (s *Service) SetLocalStorage(storage repository.Repository) {
	s.storage = storage
}

func (s *Service) PingDB() error {
	return s.storage.Ping()
}
