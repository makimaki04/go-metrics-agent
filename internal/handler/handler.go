package handler

import (
	"net/http"
	"strconv"
	"strings"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
)

type Handler struct {
	service service.MetricsService
}

func NewHandler(service service.MetricsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleReq(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, `{"error": "method not allowed"}`)
		return
	}
	h.PostMetric(w, r)

}

func (h *Handler) PostMetric(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(urlParts) != 4 {
		respondWithError(w, http.StatusNotFound, `{"error": "metric ID is required"}`)
		return
	}

	metric := models.Metrics{
		ID: urlParts[2],
		MType: urlParts[1],
	}
	metricValue := urlParts[3]

	if metric.ID == "" {
		respondWithError(w, http.StatusNotFound, `{"error": "metric ID is required"}`)
		return
	}

	switch metric.MType {
	case models.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, `{"error": "invalid delta value"}`)
			return
		}
		metric.Delta = &delta

		h.service.UpdateCounter(metric.ID, *metric.Delta)
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, `{"error": "invalid metric's value"}`)
			return
		}
		metric.Value = &value

		h.service.UpdateGauge(metric.ID, *metric.Value)
	default:
		respondWithError(w, http.StatusBadRequest, `{"error": "unknown metric type"}`)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}
