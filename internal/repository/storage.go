package repository

import (
	"fmt"
	"sync"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func (m *MemStorage) SetGauge(name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
	return nil
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	return value, ok
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

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	return value, ok
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

func (m *MemStorage) SetMetricBatch(metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return fmt.Errorf("gauge %s has no value", metric.ID)
			}
			err := m.SetGauge(metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		case "counter":
			if metric.Delta == nil {
				return fmt.Errorf("counter %s has no delta", metric.ID)
			}
			err := m.SetCounter(metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MemStorage) Ping() error {
	return nil
}
