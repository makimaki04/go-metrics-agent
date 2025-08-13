package repository

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu sync.RWMutex
}

func (m *MemStorage) SetGauge(name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
	return nil
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	if !ok {
		return 0, fmt.Errorf("gauge %q not found", name)
	}
	return value, nil
}

func (m *MemStorage) GetAllGauges() (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]float64)
	for k, v := range m.gauges {
		copy[k] = v
	}
	return copy, nil
}

func (m *MemStorage) SetCounter(name string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += value
	return nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	if !ok {
		return 0, fmt.Errorf("counter %q not found", name)
	}
	return value, nil
}

func (m *MemStorage) GetAllCounters() (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]int64)
	for k, v := range m.counters {
		copy[k] = v
	}
	return copy, nil
}

func (m *MemStorage) Ping() error {
	return nil
}