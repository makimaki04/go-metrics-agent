package agent

import (
	"testing"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCollector_CollcetMetrics(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]float64
	}{
		{
			name: "Collector simple test",
			metrics: map[string]float64{
				"Alloc": float64(15.15),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewLocalStorage()
			for name, val := range tt.metrics {
				v := val
				storage.SetMetric(name, models.Metrics{
					ID:    name,
					MType: "gauge",
					Value: &v,
				})
				metric, ok := storage.GetMetric(name)
				assert.True(t, ok)
				assert.Equal(t, &v, metric.Value)
				assert.Equal(t, name, metric.ID)
			}
		})
	}
}
