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

// MetricsService - interface for the metrics service
// UpdateMetric - method for updating a metric
// UpdateGauge - method for updating a gauge
// GetGauge - method for getting a gauge
// GetAllGauges - method for getting all gauges
// UpdateCounter - method for updating a counter
// GetCounter - method for getting a counter
// GetAllCounters - method for getting all counters
// SetLocalStorage - method for setting the local storage
// UpdateMetricBatch - method for updating a batch of metrics
// PingDB - method for pinging the database
// RegisterObserver - method for registering an observer
type MetricsService interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	UpdateGauge(name string, value float64) error
	GetGauge(ctx context.Context,name string) (float64, bool)
	GetAllGauges() (map[string]float64, error)
	UpdateCounter(name string, value int64) error
	GetCounter(name string) (int64, bool)
	GetAllCounters() (map[string]int64, error)
	SetLocalStorage(storage repository.Repository)
	UpdateMetricBatch(ctx context.Context, metrics []models.Metrics) error
	PingDB() error
	RegisterObserver(o observer.Observer)
}

// Service - struct for the metrics service
// generate:reset
type Service struct {
	storage   repository.Repository
	logger    *zap.Logger
	observers []observer.Observer
}

// NewService - method for creating a new metrics service
// create a new metrics service
func NewService(storage repository.Repository, logger *zap.Logger) MetricsService {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

// UpdateMetric - method for updating a metric
// update the value of the metric
// if error, return error
// if success, return nil
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

// sendMetricEvent - method for sending a metric event
func (s *Service) sendMetricEvent(ctx context.Context, id string) {
	var ids []string

	ids = append(ids, id)

	event := observer.AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Metrics:   ids,
		IPAddress: idFromContext(ctx),
	}

	for _, o := range s.observers {
		o.Notify(ctx, event)
	}
}

// idFromContext - method for getting the id from the context
func idFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(observer.ReqIDKey).(string)
	if !ok {
		return "unknown"
	}
	return ip
}

// UpdateGauge - method for updating a gauge
// update the value of the gauge
// if error, return error
// if success, return nil
func (s *Service) UpdateGauge(name string, value float64) error {
	return withRetry(func() error {
		return s.storage.SetGauge(name, value)
	}, s.logger)
}

// GetGauge - method for getting a gauge
// get the value of the gauge
// if error, return error
// if success, return the value of the gauge
func (s *Service) GetGauge(ctx context.Context, name string) (float64, bool) {
	value, err := retryValue(func() (float64, error) {
		v, ok := s.storage.GetGauge(ctx, name)
		if !ok {
			return 0, fmt.Errorf("metric %q not found", name)
		}
		return v, nil
	}, s.logger)
	return value, err == nil
}

// GetAllGauges - method for getting all gauges
// get all the gauges
// if error, return error
// if success, return the value of the gauges
func (s *Service) GetAllGauges() (map[string]float64, error) {
	return retryValue(func() (map[string]float64, error) {
		return s.storage.GetAllGauges()
	}, s.logger)
}

// UpdateCounter - method for updating a counter
// update the value of the counter
// if error, return error
// if success, return nil
func (s *Service) UpdateCounter(name string, value int64) error {
	return withRetry(func() error {
		return s.storage.SetCounter(name, value)
	}, s.logger)
}

// GetCounter - method for getting a counter
// get the value of the counter
// if error, return error
// if success, return the value of the counter
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

// GetAllCounters - method for getting all counters
// get all the counters
// if error, return error
// if success, return the value of the counters
func (s *Service) GetAllCounters() (map[string]int64, error) {
	return retryValue(func() (map[string]int64, error) {
		return s.storage.GetAllCounters()
	}, s.logger)
}

// SetLocalStorage - method for setting the local storage
// set the local storage
// if error, return error
// if success, return nil
func (s *Service) SetLocalStorage(storage repository.Repository) {
	s.storage = storage
}

// UpdateMetricBatch - method for updating a batch of metrics
// update the value of the metrics
// if error, return error
// if success, return nil
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

// sendMetricBatchEvent - method for sending a metric batch event
func (s *Service) sendMetricBatchEvent(ctx context.Context, ids []string) {
	event := observer.AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Metrics:   ids,
		IPAddress: idFromContext(ctx),
	}

	for _, o := range s.observers {
		o.Notify(ctx, event)
	}
}

// PingDB - method for pinging the database
// checks the connection to the database
// returns error if connection fails, nil otherwise
func (s *Service) PingDB() error {
	return s.storage.Ping()
}

// retryIntervals - intervals for retrying failed operations
var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// withRetry - retries a function with exponential backoff
// retries the function up to 3 times with intervals of 1s, 3s, 5s
// fn - function to retry
// logger - logger for logging retry attempts
// returns error if all retries fail, nil on success
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

// isTemporary - checks if an error is temporary and can be retried
// returns true for network timeout errors and PostgreSQL connection exceptions
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

// retryValue - retries a function that returns a value with exponential backoff
// generic function that retries operations returning a value
// fn - function to retry that returns a value and error
// logger - logger for logging retry attempts
// returns the value and error (nil on success)
func retryValue[T any](fn func() (T, error), logger *zap.Logger) (T, error) {
	var result T

	err := withRetry(func() error {
		var err error
		result, err = fn()
		return err
	}, logger)

	return result, err
}

// RegisterObserver - method for registering an observer
// adds an observer to the list of observers that will be notified of metric events
// o - observer to register
func (s *Service) RegisterObserver(o observer.Observer) {
	s.observers = append(s.observers, o)
}
