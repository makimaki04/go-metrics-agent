package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/observer"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"go.uber.org/zap"
)

type MetricsService interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	UpdateGauge(name string, value float64) error
	GetGauge(name string) (float64, bool)
	GetAllGauges() (map[string]float64, error)
	UpdateCounter(name string, value int64) error
	GetCounter(name string) (int64, bool)
	GetAllCounters() (map[string]int64, error)
	SetLocalStorage(storage repository.Repository)
	UpdateMetricBatch(ctx context.Context, metrics []models.Metrics) error
	PingDB() error
	RegisterObserver(o observer.Observer)
}

type Service struct {
	storage  repository.Repository
	logger   *zap.Logger
	observers []observer.Observer
}

func NewService(storage repository.Repository, logger *zap.Logger) MetricsService {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("metric %q: Delta is nil", metric.ID)
		}
		if err := s.UpdateCounter(metric.ID, *metric.Delta); err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}

		s.sendMetricEvent(ctx, metric.ID)
		return nil
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("metric %q: Value is nil", metric.ID)
		}
		if err := s.UpdateGauge(metric.ID, *metric.Value); err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}

		s.sendMetricEvent(ctx, metric.ID)
	}

	return fmt.Errorf("unknown metric type: %q", metric.MType)
}

func (s *Service) sendMetricEvent(ctx context.Context, id string) {
	var ids []string

	ids = append(ids, id)

	event := observer.AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Metrics: ids,
		IPAddress: idFromContext(ctx),
	}

	for _, o := range s.observers {
		o.Notify(ctx, event)
	}
}

func idFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(observer.ReqIDKey).(string)
    if !ok {
        return "unknown"
    }
    return ip
}


func (s *Service) UpdateGauge(name string, value float64) error {
	return withRetry(func() error {
		return s.storage.SetGauge(name, value)
	}, s.logger)
}

func (s *Service) GetGauge(name string) (float64, bool) {
	value, err := retryValue(func() (float64, error) {
		v, ok := s.storage.GetGauge(name)
		if !ok {
			return 0, fmt.Errorf("metric %q not found", name)
		}
		return v, nil
	}, s.logger)
	return value, err == nil
}

func (s *Service) GetAllGauges() (map[string]float64, error) {
	return retryValue(func() (map[string]float64, error) {
		return s.storage.GetAllGauges()
	}, s.logger)
}

func (s *Service) UpdateCounter(name string, value int64) error {
	return withRetry(func() error {
		return s.storage.SetCounter(name, value)
	}, s.logger)
}

func (s *Service) GetCounter(name string) (int64, bool) {
	value, err := retryValue(func() (int64, error) {
		v, ok := s.storage.GetCounter(name)
		if !ok {
			return 0, fmt.Errorf("counter %q not found", name)
		}
		return v, nil
	}, s.logger)
	return value, err == nil
}

func (s *Service) GetAllCounters() (map[string]int64, error) {
	return retryValue(func() (map[string]int64, error) {
		return s.storage.GetAllCounters()
	}, s.logger)
}

func (s *Service) SetLocalStorage(storage repository.Repository) {
	s.storage = storage
}

func (s *Service) UpdateMetricBatch(ctx context.Context, metrics []models.Metrics) error {
	err := withRetry(func() error {
		return s.storage.SetMetricBatch(metrics)
	}, s.logger)

	if err != nil {
		return err
	}

	ids := make([]string, 0, len(metrics))
	for _, m := range metrics {
		ids = append(ids, m.ID)
	}

	s.sendMetricBatchEvent(ctx, ids)

	return nil
}

func (s *Service) sendMetricBatchEvent(ctx context.Context, ids []string) {
	event := observer.AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Metrics: ids,
		IPAddress: idFromContext(ctx),
	}

	for _, o := range s.observers {
		o.Notify(ctx, event)
	}
}

func (s *Service) PingDB() error {
	return s.storage.Ping()
}

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func withRetry(fn func() error, logger *zap.Logger) error {
	var lastErr error

	for i, interval := range retryIntervals {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if isTemporary(err) {
			logger.Sugar().Infof("Temporary error, retrying in %s: %v\n", interval, err)
		} else {
			return err
		}

		if i < len(retryIntervals)-1 {
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("operation failed after retries: %w", lastErr)
}

func isTemporary(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		return true
	}

	return false
}

func retryValue[T any](fn func() (T, error), logger *zap.Logger) (T, error) {
	var result T

	err := withRetry(func() error {
		var err error
		result, err = fn()
		return err
	}, logger)

	return result, err
}

func (s *Service) RegisterObserver(o observer.Observer) {
	s.observers = append(s.observers, o)
}