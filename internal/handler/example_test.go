package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
	"go.uber.org/zap"
)

func ExampleHandler_UpdateMetric_counter() {
	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/update/{MType}/{ID}/{value}", handler.HandleReq)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/CMD/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	//check response status
	//should be 200
	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}

func ExampleHandler_UpdateMetric_gauge() {
	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/update/{MType}/{ID}/{value}", handler.HandleReq)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/metric/10.5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	//check response status
	//should be 200
	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}

func ExampleHandler_GetMetric() {
	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/update/{MType}/{ID}/{value}", handler.HandleReq)
	r.Get("/value/{MType}/{ID}", handler.HandleReq)

	post := httptest.NewRequest(http.MethodPost, "/update/gauge/metric/10.5", nil)
	postW := httptest.NewRecorder()
	r.ServeHTTP(postW, post)
	postW.Result().Body.Close()

	getReq := httptest.NewRequest(http.MethodGet, "/value/gauge/metric", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	res := getW.Result()
	defer res.Body.Close()

	//check response status
	//should be 200
	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}

func ExampleHandler_UpdateMetric() {
	v := int64(5)
	metric := models.Metrics{
		ID:    "Metric",
		MType: "counter",
		Delta: &v,
	}
	jsonBody, err := json.Marshal(metric)
	if err != nil {
		log.Printf("failed to marshal metric: %v", err)
	}
	rBody := bytes.NewReader(jsonBody)

	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/update", handler.UpdateMetric)

	updR := httptest.NewRequest(http.MethodPost, "/update", rBody)
	updR.Header.Set("Content-Type", "application/json")

	updW := httptest.NewRecorder()
	r.ServeHTTP(updW, updR)

	res := updW.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}

func ExampleHandler_PostMetricInfo() {
	v := float64(15.5)
	metric := models.Metrics{
		ID:    "CMD",
		MType: "gauge",
		Value: &v,
	}
	jsonBody, err := json.Marshal(metric)
	if err != nil {
		log.Printf("failed to marshal metric: %v", err)
	}
	rBody := bytes.NewReader(jsonBody)

	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/value", handler.PostMetricInfo)

	req := httptest.NewRequest(http.MethodPost, "/value", rBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}

func ExampleHandler_GetAllMetrics() {
	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	service.UpdateCounter("PollCount", 1)
	service.UpdateGauge("Metric", 10.5)

	r := chi.NewRouter()
	r.Get("/", handler.GetAllMetrics)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read response body: %v", err)
	}
	
	fmt.Printf("response body: %s", string(body))
}

func ExampleHandler_UpdateMetricBatch() {
	d := int64(5)
	v := float64(1.23)
	batch := []models.Metrics{
		{
			ID: "metric1",
			MType: "counter",
			Delta: &d,
		},
		{
			ID: "metric2",
			MType: "gauge",
			Value: &v,
		},
	}

	jsonB, err := json.Marshal(batch)
	if err != nil {
		log.Printf("failed to marshal metric: %v", err)
	}

	rb := bytes.NewReader(jsonB)

	repo := repository.NewStorage()
	service := service.NewService(repo, zap.NewExample())
	handler := handler.NewHandler(service, "test")

	r := chi.NewRouter()
	r.Post("/updates", handler.UpdateMetricBatch)

	req := httptest.NewRequest(http.MethodPost, "/updates", rb)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("expected status 200, got %d", res.StatusCode)
	}
}