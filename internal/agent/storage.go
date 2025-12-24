package agent

import (
	"sync"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

//LocalStorage - struct for the local storage
//metrics - map of metrics
//mu - mutex for the metrics
type LocalStorage struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

//NewLocalStorage - method for creating a new local storage
//create a new local storage
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		metrics: make(map[string]models.Metrics),
	}
}

//SetMetric - method for setting a metric
//set the metric in the storage
func (l *LocalStorage) SetMetric(name string, metric models.Metrics) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.metrics[name] = models.Metrics(metric)
}

//GetMetric - method for getting a metric
//get the metric from the storage
//if metric not found, return false
//if metric found, return the metric and true
func (l *LocalStorage) GetMetric(name string) (models.Metrics, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	metric, ok := l.metrics[name]
	return metric, ok
}

//GetAll - method for getting all metrics from the storage
//get all the metrics from the storage
//return the copy of the metrics map
func (l *LocalStorage) GetAll() map[string]models.Metrics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	copy := make(map[string]models.Metrics)
	for k, v := range l.metrics {
		copy[k] = v
	}
	return copy
}
