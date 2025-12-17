package agent

import (
	"sync"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

type LocalStorage struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (l *LocalStorage) SetMetric(name string, metric models.Metrics) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.metrics[name] = models.Metrics(metric)
}

func (l *LocalStorage) GetMetric(name string) (models.Metrics, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	metric, ok := l.metrics[name]
	return metric, ok
}

func (l *LocalStorage) GetAll() map[string]models.Metrics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	copy := make(map[string]models.Metrics)
	for k, v := range l.metrics {
		copy[k] = v
	}
	return copy
}
