package handler

import (
	"net/http"
	"strconv"
	"github.com/makimaki04/go-metrics-agent.git/internal/model"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "method not allowed"}`))
		return
	}
	PostMetricsHandler(w, r)

}

func PostMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metric := models.Metrics{
		ID:    r.FormValue("id"),
		MType: r.FormValue("type"),
	}

	if metric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "metric id is required"}`))
		return
	}

	switch metric.MType {
	case models.Counter:
		deltaStr := r.FormValue("delta")
		if deltaStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "delta is required for counter"}`))
		}

		delta, err := strconv.ParseInt(deltaStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid delra value"}`))
		}
		metric.Delta = &delta

	case models.Gauge:

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "unknown metric type"}`))
		return
	}

	w.WriteHeader(http.StatusOK)

}
