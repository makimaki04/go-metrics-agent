package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_SetGauge(t *testing.T) {
	storage := NewStorage()

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
			assert.True(t, exist, "Gauge value should exist")
			assert.Equal(t, tt.input.value, value)
		})
	}
}

func TestMemStorage_SetCounter(t *testing.T) {
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
			storage := NewStorage()

			var counterName string
			for _, c := range tt.input {
				storage.SetCounter(c.name, c.value)
				counterName = c.name
			}

			value, exist := storage.GetCounter(counterName)
			assert.True(t, exist, "counter should exist")
			assert.Equal(t, tt.want, value)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
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
			storage := NewStorage()
			storage.SetGauge(tt.input.name, tt.input.value)
			value, ok := storage.GetGauge(tt.key)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
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
				name:  "requests",
				value: 10,
			},
			key:       "requests",
			wantValue: 10,
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
			storage := NewStorage()
			storage.SetCounter(tt.input.name, tt.input.value)
			value, ok := storage.GetCounter(tt.key)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
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
			storage := NewStorage()
			storage.SetGauge(tt.mock.name, tt.mock.value)
			gauges := storage.GetAllGauges()
			assert.Equal(t, tt.want, gauges)
		})
	}
}

func TestMemStorage_GetAllCounters(t *testing.T) {
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
			storage := NewStorage()
			storage.SetCounter(tt.mock.name, tt.mock.value)
			gauges := storage.GetAllCounters()
			assert.Equal(t, tt.want, gauges)
		})
	}
}
