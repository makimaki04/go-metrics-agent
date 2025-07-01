package repository

import "sync"

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu sync.RWMutex
}

type Repository interface {
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

func NewStorage() Repository {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	return value, ok
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]float64)
	for k, v := range m.gauges {
		copy[k] = v
	}
	return copy
}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += value
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	return value, ok
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]int64)
	for k, v := range m.counters {
		copy[k] = v
	}
	return copy
}