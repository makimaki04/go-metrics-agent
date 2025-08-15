package repository

import (
	"database/sql"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	SetGauge(name string, value float64) error
	SetCounter(name string, value int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() (map[string]float64, error)
	GetAllCounters() (map[string]int64, error)
	SetMetricBatch(metrics []models.Metrics) error
	Ping() error
}

func NewStorage() Repository {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func NewDBStorage(db *sql.DB, logger *zap.Logger) Repository {
	return &DBStorage{
		db:     db,
		logger: logger,
	}
}
