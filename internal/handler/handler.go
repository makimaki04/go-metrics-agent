package handler

import (
	"html/template"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/go-chi/chi/v5"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
)

type Handler struct {
	service service.MetricsService
}

func NewHandler(service service.MetricsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	gauges := h.service.GetAllGauges()
	counters := h.service.GetAllCounters()
	const marking = `
					<!DOCTYPE html>
					<html>
						<head>
						<title>Metrics List</title>
						<style>
							body {
								width: 520px;
								margin: 0 auto;
							}
							ul {
								list-style: none;
								padding: 0;
								background: white;
								border-radius: 5px;
								box-shadow: 0 2px 5px rgba(0,0,0,0.1);
								padding: 15px;
							}
							li {
								padding: 8px 15px;
								margin: 5px 0;
								background: #f8f9fa;
								border-radius: 3px;
								display: flex;
								justify-content: space-between;
							}	
							.metric-name {
								font-weight: bold;
								color: #2c3e50;
							}
							.metric-value {
								font-family: monospace;
								color: #e74c3c;
							}
						</style>
						</head>
						<body>
							<h1>Metrics List</h1>
							<h2>Counters</h2>
							<ul>
								{{range $key, $value := .Counters}}
								<li style="list-style-type:none">
									<span class="metric-name">{{ $key }}</span>
									<span class="metric-value">{{ $value }}</span>
								</li>
								{{end}}
							</ul>
							<h2>Gauges</h2>
							<ul>
								{{range $key, $value := .Gauges}}
								<li style="list-style-type:none">
									<span class="metric-name">{{ $key }}</span>
									<span class="metric-value">{{ $value }}</span>
								</li>
								{{end}}
							</ul>
						</body>
					</html>`
	tmpl, err := template.New("metrics").Parse(marking)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templateData := struct{
		Counters map[string]int64
		Gauges map[string]float64
	}{
		Counters: counters,
		Gauges: gauges,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8") //обязательно ли прописывать заголовки? Насколько я знаю, Execute автоматически выставляет text/html; charset=utf-8
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) HandleReq(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetMetric(w, r)
	case http.MethodPost:
		h.PostMetric(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, `{"error": "method not allowed"}`)
		return
	}
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metric := models.Metrics{
		ID:    chi.URLParam(r, ("ID")),
		MType: chi.URLParam(r, "MType"),
	}
	var value string
	switch metric.MType {
	case models.Counter:
		m, ok := h.service.GetCounter(metric.ID)
		if !ok {
			respondWithError(w, http.StatusNotFound, `{"error": "invalid metric"}`)
			return
		}
		value = fmt.Sprintf(`%v`, m)
	case models.Gauge:
		m, ok := h.service.GetGauge(metric.ID)
		if !ok {
			respondWithError(w, http.StatusNotFound, `{"error": "invalid metric"}`)
			return
		}
		value = fmt.Sprintf(`%v`, m)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *Handler) PostMetric(w http.ResponseWriter, r *http.Request) {
	metric := models.Metrics{
		ID:    chi.URLParam(r, "ID"),
		MType: chi.URLParam(r, "MType"),
	}
	metricValue := chi.URLParam(r, "value")

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
	log.Printf("Handle %s %s = %s", metric.MType, metric.ID, metricValue)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}
