package repository

import (
	"fmt"
	"sync"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

//MemStorage - struct for the memory storage
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

//SetGauge - method for setting a gauge
//set the value of the gauge
//if error, return error
//if success, return nil
func (m *MemStorage) SetGauge(name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
	return nil
}

//GetGauge - method for getting a gauge
//get the value of the gauge
//if error, return error
//if success, return the value of the gauge
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	return value, ok
}

//GetAllGauges - method for getting all gauges
//get all the gauges
//if error, return error
//if success, return the value of the gauges
func (m *MemStorage) GetAllGauges() (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]float64)
	for k, v := range m.gauges {
		copy[k] = v
	}
	return copy, nil
}

//SetCounter - method for setting a counter
//set the value of the counter
//if error, return error
//if success, return nil
func (m *MemStorage) SetCounter(name string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += value
	return nil
}

//GetCounter - method for getting a counter
//get the value of the counter
//if error, return error
//if success, return the value of the counter
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	return value, ok
}

//GetAllCounters - method for getting all counters
//get all the counters
//if error, return error
//if success, return the value of the counters
func (m *MemStorage) GetAllCounters() (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]int64)
	for k, v := range m.counters {
		copy[k] = v
	}
	return copy, nil
}

//SetMetricBatch - method for setting a batch of metrics
//set the value of the metrics
//if error, return error
//if success, return nil
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

//Ping - method for pinging the database
func (m *MemStorage) Ping() error {
	return nil
}
