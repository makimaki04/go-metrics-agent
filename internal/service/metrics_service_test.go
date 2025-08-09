package service

import (
	"testing"

	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestService_UpdateCounter(t *testing.T) {
	type counter struct {
		name  string
		value int64
	}
	tests := []struct {
		name  string
		input []counter
		want  int64
	}{
		{
			name: "Single counter update",
			input: []counter{
				{name: "requests", value: 5},
			},
			want: 5,
		},
		{
			name: "Multiple increments",
			input: []counter{
				{name: "requests", value: 2},
				{name: "requests", value: 3},
				{name: "requests", value: 10},
			},
			want: 15,
		},
		{
			name: "Negative increment",
			input: []counter{
				{name: "errors", value: -1},
			},
			want: -1,
		},
		{
			name: "Zero increment",
			input: []counter{
				{name: "zero_count", value: 0},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := NewService(storage)

			var counterName string
			for _, c := range tt.input {
				service.UpdateCounter(c.name, c.value)
				counterName = c.name
			}

			value, exist := storage.GetCounter(counterName)
			assert.True(t, exist, "counter should exist")
			assert.Equal(t, tt.want, value)
		})
	}
}

func TestService_UpdateGauge(t *testing.T) {
	storage := repository.NewStorage()

	type gauge struct {
		name  string
		value float64
	}

	tests := []struct {
		name  string
		input gauge
	}{
		{
			name:  "Set CPU usage gauge",
			input: gauge{name: "cpu_usage", value: 87.3},
		},
		{
			name:  "Set disk usage gauge",
			input: gauge{name: "disk_usage", value: 99.9},
		},
		{
			name:  "Set zero gauge",
			input: gauge{name: "zero", value: 0.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.SetGauge(tt.input.name, tt.input.value)
			value, exist := storage.GetGauge(tt.input.name)
			assert.True(t, exist, "gauge should exist")
			assert.Equal(t, tt.input.value, value)
		})
	}
}

func TestService_GetGauge(t *testing.T) {
	type input struct {
		name  string
		value float64
	}
	tests := []struct {
		name      string
		input     input
		key       string
		wantValue float64
		wantOk    bool
	}{
		{
			name: "Gauge exists",
			input: input{
				name:  "CPU",
				value: 75.5,
			},
			key:       "CPU",
			wantValue: 75.5,
			wantOk:    true,
		},
		{
			name:   "Gauge does not exist",
			key:    "UNKNOWN",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := NewService(storage)
			service.UpdateGauge(tt.input.name, tt.input.value)

			value, ok := service.GetGauge(tt.key)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestService_GetCounter(t *testing.T) {
	type input struct {
		name  string
		value int64
	}
	tests := []struct {
		name      string
		input     input
		key       string
		wantValue int64
		wantOk    bool
	}{
		{
			name: "Counter exists",
			input: input{
				name:  "PollCount",
				value: 5,
			},
			key:       "PollCount",
			wantValue: 5,
			wantOk:    true,
		},
		{
			name:   "Counter does not exist",
			key:    "UNKNOWN",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := NewService(storage)
			service.UpdateCounter(tt.input.name, tt.input.value)

			value, ok := service.GetCounter(tt.key)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestService_GetAllCounters(t *testing.T) {
	type mock struct {
		name  string
		value int64
	}
	tests := []struct {
		name string
		mock mock
		want map[string]int64
	}{
		{
			name: "Simple test",
			mock: mock{
				name:  "PollCount",
				value: 5,
			},
			want: map[string]int64{
				"PollCount": 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := NewService(storage)
			service.UpdateCounter(tt.mock.name, tt.mock.value)
			gauges := service.GetAllCounters()
			assert.Equal(t, tt.want, gauges)
		})
	}
}

func TestService_GetAllGauges(t *testing.T) {
		type mock struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		mock mock
		want map[string]float64
	}{
		{
			name: "Simple test",
			mock: mock{
				name:  "CMD",
				value: 12.34,
			},
			want: map[string]float64{
				"CMD": 12.34,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := NewService(storage)
			service.UpdateGauge(tt.mock.name, tt.mock.value)
			gauges := service.GetAllGauges()
			assert.Equal(t, tt.want, gauges)
		})
	}
}
