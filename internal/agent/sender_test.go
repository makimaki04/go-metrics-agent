package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct{}

func (m *mockStorage) GetAll() map[string]models.Metrics {
	val := 123.45
	delta := int64(10)

	return map[string]models.Metrics{
		"test_gauge": {
			ID:    "test_gauge",
			MType: "gauge",
			Value: &val,
		},
		"test_counter": {
			ID:    "test_counter",
			MType: "counter",
			Delta: &delta,
		},
	}
}

func TestSender_SendMetrics(t *testing.T) {
	var receivedRequests []struct {
		Method  string
		Path    string
		Headers http.Header
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests = append(receivedRequests, struct {
			Method  string
			Path    string
			Headers http.Header
		}{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
		})
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
	}))

	defer testServer.Close()

	storage := &mockStorage{}
	sender := NewSender(resty.New(), testServer.URL, storage, "")

	sender.SendMetrics()

	for _, req := range receivedRequests {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.True(t, strings.HasPrefix(req.Path, "/update/"))
		assert.Equal(t, "text/plain", req.Headers.Get("Content-Type"))
	}
}
