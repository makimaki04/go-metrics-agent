package repository

import "database/sql"

type Repository interface {
	SetGauge(name string, value float64) error
	SetCounter(name string, value int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() (map[string]float64, error) 
	GetAllCounters() (map[string]int64, error)
	Ping() error
}

func NewStorage() Repository {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func NewDBStorage(db *sql.DB) Repository {
	return &DBStorage{
		db: db,
	}
}