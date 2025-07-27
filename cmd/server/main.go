package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
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
	storage := repository.NewStorage()
	service := service.NewService(storage)
	handler := handler.NewHandler(service)

	file, error := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_RDWR, 0666)
	if error != nil {
		logger.Error("Couldn't create or open the local storage file")
	}

	defer file.Close()

	if cfg.Restore {
		data, err := file.Stat()
		if err != nil {
			logger.Info("Couldn't read the file")
		}
		
		if data.Size() > 0 {
			var  metrics struct {
				Counters map[string]int64    `json:"counters"`
				Gauges   map[string]float64  `json:"gauges"`
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

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(handler.GetAllMetrics), handlersLogger))
		r.Route("/value", func(r chi.Router) {
			r.Post("/",  middleware.WithLogging(middleware.GzipMiddleware(handler.PostMetrcInfo), handlersLogger))
			r.Route("/{MType}/{ID}", func(r chi.Router) {
				r.Get("/",  middleware.WithLogging(middleware.GzipMiddleware(handler.HandleReq), handlersLogger))
			})
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/",  middleware.WithLogging(middleware.GzipMiddleware(handler.UpdateMetric), handlersLogger))
			r.Route("/{MType}/{ID}/{value}", func(r chi.Router) {
				r.Post("/",  middleware.WithLogging(handler.HandleReq, handlersLogger))
			})
		})
		
	})

	go func() {
		err := http.ListenAndServe(cfg.Address, r)
		if err != nil {
			panic(err)
		}
	}()

	ticker := time.NewTicker(time.Duration(cfg.StoreInt) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		allMetrics := struct {
			Counters map[string]int64    `json:"counters"`
			Gauges   map[string]float64  `json:"gauges"`
		}{
			Counters: service.GetAllCounters(),
			Gauges:   service.GetAllGauges(),
		}

		data, err := json.MarshalIndent(allMetrics, "", "	")
		if err != nil {
			logger.Error("Failed to marshal all metrics", zap.Error(err))
			continue
		}

		if err := file.Truncate(0); err != nil {
			logger.Error("Failed to truncate file", zap.Error(err))
			continue
		}
		
		if _, err := file.Seek(0, 0); err != nil {
			logger.Error("Failed to seek to beginning of file", zap.Error(err))
			continue
		}

		if _, err := file.Write(data); err != nil {
			logger.Error("Failed to write metrics to file", zap.Error(err))
			continue
		}

		logger.Info("metrics successfully added to the local storage located in ./data/save.json")
	}
}