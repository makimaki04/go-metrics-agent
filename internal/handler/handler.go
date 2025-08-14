package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	gauges, err := h.service.GetAllGauges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	counters, err := h.service.GetAllCounters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
		m, err := h.service.GetCounter(metric.ID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, `{"error": "invalid metric"}`)
			return
		}
		value = fmt.Sprintf(`%v`, m)
	case models.Gauge:
		m, err := h.service.GetGauge(metric.ID)
		if err != nil {
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
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, `{"error": "invalid metric's value"}`)
			return
		}
		metric.Value = &value
	default:
		respondWithError(w, http.StatusBadRequest, `{"error": "unknown metric type"}`)
		return
	}

	h.service.UpdateMetric(metric)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, `{"error": "empty request body"}`)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, `{"error": "wrong body structure"}`)
		return
	}

	if err := h.service.UpdateMetric(metric); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf(`{"error": "%v"}`, err))
    	return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Println(metric)
}

func (h *Handler) PostMetrcInfo(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, `{"error": "empty request body"}`)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, `{"error": "wrong body structure"}`)
		return
	}

	switch metric.MType {
	case models.Counter:
		d, err := h.service.GetCounter(metric.ID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, `{"error": "invalid metric"}`)
			return
		}
		metric.Delta = &d
	case models.Gauge:
		v, err := h.service.GetGauge(metric.ID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, `{"error": "invalid metric"}`)
			return
		}
		metric.Value = &v
	}

	resp, err := json.MarshalIndent(metric, "", "	")
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, `{"error": "empty response body"}`)
		return
	}

	fmt.Println(metric)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) PingDatabase(w http.ResponseWriter, r *http.Request) {
	err := h.service.PingDB()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, `{"error}": "failed database connection"`)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}
