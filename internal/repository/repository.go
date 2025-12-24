package repository

import (
	"context"
	"database/sql"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"go.uber.org/zap"
)

// Repository - interface for the repository
// SetGauge - method for setting a gauge
// SetCounter - method for setting a counter
// GetGauge - method for getting a gauge
// GetCounter - method for getting a counter
// GetAllGauges - method for getting all gauges
// GetAllCounters - method for getting all counters
// SetMetricBatch - method for setting a batch of metrics
// Ping - method for pinging the database
type Repository interface {
	SetGauge(name string, value float64) error
	SetCounter(name string, value int64) error
	GetGauge(ctx context.Context,name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() (map[string]float64, error)
	GetAllCounters() (map[string]int64, error)
	SetMetricBatch(metrics []models.Metrics) error
	Ping() error
}

// NewStorage - creates a new in-memory storage implementation
// returns a Repository interface implementation using MemStorage
// the storage is thread-safe and stores metrics in memory
func NewStorage() Repository {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// NewDBStorage - creates a new database storage implementation
// db - database connection to use for storage
// logger - logger instance for logging operations
// returns a Repository interface implementation using DBStorage
// the storage persists metrics in PostgreSQL database
func NewDBStorage(db *sql.DB, logger *zap.Logger) Repository {
	return &DBStorage{
		db:     db,
		logger: logger,
	}
}
