package agent

import models "github.com/makimaki04/go-metrics-agent.git/internal/model"

type LocalStorage struct {
	metrics map[string]models.Metrics
}

type Storage interface {
	SetMetric(name string, metric models.Metrics)
	GetAll() map[string]models.Metrics
}

func NewLocalStorage() Storage {
	return &LocalStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (l *LocalStorage) SetMetric(name string, metric models.Metrics) {
	l.metrics[name] = models.Metrics(metric)
}

func (l *LocalStorage) GetAll() map[string]models.Metrics {
	copy := make(map[string]models.Metrics)
	for k, v := range l.metrics {
		copy[k] = v
	}
	return copy
} 