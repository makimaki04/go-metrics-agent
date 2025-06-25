package agent

import (
	"testing"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestLocalStorage_SetMetric(t *testing.T) {
	storage := NewLocalStorage()
	v := 75.25
	var d int64 = 15
	type args struct {
		name   string
		metric models.Metrics
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set gauge simple test",
			args: args{
				name: "CPU",
				metric: models.Metrics{
					ID:    "CPU",
					MType: "gauge",
					Value: &v,
				},
			},
		},
		{
			name: "Set counter simple test",
			args: args{
				name: "requests",
				metric: models.Metrics{
					ID:    "requests",
					MType: "counter",
					Delta: &d,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.SetMetric(tt.args.name, tt.args.metric)
			metric, exist := storage.GetMetric(tt.args.name)
			assert.True(t, exist, "metric should exist")
			assert.Equal(t, tt.args.metric.Value, metric.Value)
		})
	}
}

func TestLocalStorage_GetMetric(t *testing.T) {
	v := 75.25
	var d int64 = 15
	type args struct {
		name   string
		metric models.Metrics
	}
	tests := []struct {
		name   string
		args   args
		want   models.Metrics
		wantOk bool
	}{
		{
			name: "Get gauge simple test",
			args: args{
				name: "CPU",
				metric: models.Metrics{
					ID:    "CPU",
					MType: "gauge",
					Value: &v,
				},
			},
			want: models.Metrics{
				ID:    "CPU",
				MType: "gauge",
				Value: &v,
			},
			wantOk: true,
		},
		{
			name: "Get counter simple test",
			args: args{
				name: "requests",
				metric: models.Metrics{
					ID:    "requests",
					MType: "counter",
					Delta: &d,
				},
			},
			want: models.Metrics{
				ID:    "requests",
				MType: "counter",
				Delta: &d,
			},
			wantOk: true,
		},
		{
			name: "Get counter simple test",
			args: args{
				name: "name",
				metric: models.Metrics{
					ID: "name",
				},
			},
			want:   models.Metrics{},
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewLocalStorage()
			if tt.wantOk == false {
				metric, ok := storage.GetMetric(tt.args.name)
				assert.Equal(t, tt.wantOk, ok)
				assert.Equal(t, tt.want, metric)
				return
			}

			storage.SetMetric(tt.args.name, tt.args.metric)
			metric, ok := storage.GetMetric(tt.args.name)
			assert.Equal(t, tt.want, metric)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestLocalStorage_GetAll(t *testing.T) {
	v := 75.25
	var d int64 = 15
	type args struct {
		name   string
		metric models.Metrics
	}
	tests := []struct {
		name   string
		args   []args
		want   map[string]models.Metrics
	}{
		{
			name: "Get gauge simple test",
			args: []args{
				{
					name: "CPU",
					metric: models.Metrics{
						ID:    "CPU",
						MType: "gauge",
						Value: &v,
					},
				},
				{
					name: "requests",
					metric: models.Metrics{
						ID:    "requests",
						MType: "counter",
						Delta: &d,
					},
				},
			},
			want: map[string]models.Metrics{
				"CPU" : {
					ID:    "CPU",
					MType: "gauge",
					Value: &v,
				},
				"requests" : {
					ID:    "requests",
					MType: "counter",
					Delta: &d,
				},
			}, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewLocalStorage()
			for _, m := range tt.args {
				storage.SetMetric(m.name, m.metric)
			}

			metrics := storage.GetAll()
			assert.Equal(t, tt.want, metrics)
		})
	}
}
