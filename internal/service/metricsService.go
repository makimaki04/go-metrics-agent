package service

import (
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
)

type MetricsService interface {
	UpdateMetric(metric models.Metrics)
	UpdateGauge(name string, value float64)
	GetGauge(name string) (float64, bool)
	GetAllGauges() map[string]float64
	UpdateCounter(name string, value int64)
	GetCounter(name string) (int64, bool)
	GetAllCounters() map[string]int64
}

type Service struct {
	storage repository.Repository
}

func NewService(storage repository.Repository) MetricsService {
	return &Service{storage: storage}
}

func (s *Service) UpdateMetric(metric models.Metrics) {
	switch metric.MType {
	case models.Counter:
		s.UpdateCounter(metric.ID, *metric.Delta)
	case models.Gauge:
		s.UpdateGauge(metric.ID, *metric.Value)
	}
}

func (s *Service) UpdateGauge(name string, value float64) {
	s.storage.SetGauge(name, value)
}

func (s *Service) GetGauge(name string) (float64, bool) {
	return s.storage.GetGauge(name)
}

func (s *Service) GetAllGauges() map[string]float64 {
	return s.storage.GetAllGauges()
}

func (s *Service) UpdateCounter(name string, value int64) {
	s.storage.SetCounter(name, value)
}

func (s *Service) GetCounter(name string) (int64, bool) {
	return s.storage.GetCounter(name)
}

func (s *Service) GetAllCounters() map[string]int64 {
	return s.storage.GetAllCounters()
}