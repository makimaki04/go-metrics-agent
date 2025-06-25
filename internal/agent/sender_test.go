package agent

import (
	"testing"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

func TestSender_SendMetrics(t *testing.T) {
	type args struct {
		metrics map[string]models.Metrics
	}
	tests := []struct {
		name string
		s    Sender
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.SendMetrics()
		})
	}
}
