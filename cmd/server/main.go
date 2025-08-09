package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	"github.com/makimaki04/go-metrics-agent.git/internal/middleware"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
	"go.uber.org/zap"
)

func main() {
	setConfig()

	logger := zap.Must(zap.NewDevelopment())

	defer logger.Sync()

	handlersLogger := logger.With(
		zap.String("handler", "Handle Request"),
	)

	db, err := sql.Open("pgx", cfg.Database)
	if err != nil {
		panic("Database connection error:" + err.Error())
	}
	defer db.Close()

	storage := repository.NewStorage()
	service := service.NewService(storage)
	handler := handler.NewHandler(service, db)

	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Fatal("Couldn't create directory for storage file", zap.Error(err))
	}

	if cfg.Restore {
		loadMetricsFromFile(cfg.FilePath, service, logger)
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(handler.GetAllMetrics), handlersLogger))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(handler.PostMetrcInfo), handlersLogger))
			r.Route("/{MType}/{ID}", func(r chi.Router) {
				r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(handler.HandleReq), handlersLogger))
			})
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(handler.UpdateMetric), handlersLogger))
			r.Route("/{MType}/{ID}/{value}", func(r chi.Router) {
				r.Post("/", middleware.WithLogging(handler.HandleReq, handlersLogger))
			})
		})
		r.Route("/ping", func (r chi.Router)  {
			r.Get("/",  middleware.WithLogging(middleware.GzipMiddleware(handler.PingDatabase), handlersLogger))
		})

	})

	go func() {
		err := http.ListenAndServe(cfg.Address, r)
		if err != nil {
			panic(fmt.Errorf("server failed to start on %s: %w", cfg.Address, err))
		}
	}()

	if cfg.StoreInt == 0 {
		saveMetrcisToFile(cfg.FilePath, service, logger)
	} else {
		ticker := time.NewTicker(time.Duration(cfg.StoreInt) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			saveMetrcisToFile(cfg.FilePath, service, logger)
		}
	}
}

func loadMetricsFromFile(path string, service service.MetricsService, logger *zap.Logger) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Error("Failed to open metrics file for reading", zap.Error(err))
		return
	}
	defer file.Close()

	data, err := file.Stat()
	if err != nil {
		logger.Info("Couldn't read the file")
	}

	if data.Size() > 0 {
		var metrics struct {
			Counters map[string]int64   `json:"counters"`
			Gauges   map[string]float64 `json:"gauges"`
		}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&metrics); err != nil {
			logger.Info("Couldn't parse data")
		} else {
			for key, value := range metrics.Gauges {
				service.UpdateGauge(key, value)
			}

			for key, value := range metrics.Counters {
				service.UpdateCounter(key, value)
			}
			logger.Info("metrics successfully loaded from local storage located in ./data/save.json")
		}
	} else {
		logger.Info("the local file exists, but is empty")
	}
}

func saveMetrcisToFile(path string, service service.MetricsService, logger *zap.Logger) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error("Failed to open metrics file for writing", zap.Error(err))
		return
	}
	defer file.Close()

	allMetrics := struct {
		Counters map[string]int64   `json:"counters"`
		Gauges   map[string]float64 `json:"gauges"`
	}{
		Counters: service.GetAllCounters(),
		Gauges:   service.GetAllGauges(),
	}

	data, err := json.MarshalIndent(allMetrics, "", "	")
	if err != nil {
		logger.Error("Failed to marshal all metrics", zap.Error(err))
		return
	}

	if _, err := file.Write(data); err != nil {
		logger.Error("Failed to write metrics to file", zap.Error(err))
		return
	}

	logger.Info("metrics successfully added to the local storage located in ./data/save.json")
}
