package service

import (
	"errors"
	"fmt"
	"time"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	UpdateMetricBatch(metrics []models.Metrics) error
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
	return withRetry(func() error {
		return s.storage.SetGauge(name, value)
	})
}

func (s *Service) GetGauge(name string) (float64, bool) {
	value, err := retryValue(func() (float64, error) {
		v, ok := s.storage.GetGauge(name)
		if !ok {
			return 0, fmt.Errorf("metric %q not found", name)
		}
		return v, nil
	})
	return value, err == nil
}

func (s *Service) GetAllGauges() (map[string]float64, error) {
	return retryValue(func() (map[string]float64, error) {
		return s.storage.GetAllGauges()
	})
}

func (s *Service) UpdateCounter(name string, value int64) error {
	return withRetry(func() error {
		return s.storage.SetCounter(name, value)
	})
}

func (s *Service) GetCounter(name string) (int64, bool) {
	value, err := retryValue(func() (int64, error) {
		v, ok := s.storage.GetCounter(name)
		if !ok {
			return 0, fmt.Errorf("counter %q not found", name)
		}
		return v, nil
	})
	return value, err == nil
}

func (s *Service) GetAllCounters() (map[string]int64, error) {
	return retryValue(func() (map[string]int64, error) {
		return s.storage.GetAllCounters()
	})
}

func (s *Service) SetLocalStorage(storage repository.Repository) {
	s.storage = storage
}

func (s *Service) UpdateMetricBatch(metrics []models.Metrics) error {
	return withRetry(func() error {
		return s.storage.SetMetricBatch(metrics)
	})
}

func (s *Service) PingDB() error {
	return s.storage.Ping()
}

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func withRetry(fn func() error) error {
	var lastErr error

	for i, interval := range retryIntervals {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsConnectionException(pgErr.Code) {
				fmt.Printf("Connection exception, retrying in %s: %v\n", interval, err)
			} else {
				return err
			}
		}

		if i < len(retryIntervals)-1 {
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("operation failed after retries: %w", lastErr)
}

func retryValue[T any](fn func() (T, error)) (T, error) {
	var result T

	err := withRetry( func() error {
		var err error
		result, err = fn()
		return err
	})

	return result, err
}