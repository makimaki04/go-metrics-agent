package service

import (
	"context"
	"database/sql"
	"time"

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
	SetLocalStorage(storage repository.Repository)
	SetDB(db *sql.DB)
	PingDB() error
}

type Service struct {
	storage repository.Repository
	db *sql.DB
}

func NewService() MetricsService {
	return &Service{}
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

func (s *Service) SetLocalStorage(storage repository.Repository) {
	s.storage = storage
}

func (s *Service) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil 
}

func (s *Service) SetDB(db *sql.DB) {
	s.db = db
}