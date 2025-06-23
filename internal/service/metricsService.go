package service

import "github.com/makimaki04/go-metrics-agent.git/internal/repository"

type MetricsService interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type Service struct {
	storage repository.Repository
}

func NewService(storage repository.Repository) MetricsService {
	return &Service{storage: storage}
}

func (s *Service) UpdateGauge(name string, value float64) {
	s.storage.SetGauge(name, value)
}

func (s *Service) UpdateCounter(name string, value int64) {
	s.storage.SetCounter(name, value)
}
