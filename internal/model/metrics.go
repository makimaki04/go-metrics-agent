package models

// Constants for the metric types
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics - struct for metrics
// ID - id of the metric
// MType - type of the metric
// Delta - delta of the metric
// Value - value of the metric
// Hash - hash of the metric
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
