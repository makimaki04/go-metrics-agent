package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	"github.com/makimaki04/go-metrics-agent.git/internal/middleware"
	"github.com/makimaki04/go-metrics-agent.git/internal/migrations"
	"github.com/makimaki04/go-metrics-agent.git/internal/observer"
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

	var storage repository.Repository
	var mService service.MetricsService

	switch {
	case cfg.DSN != "":
		db, storage := initDBStorage(logger)
		defer db.Close()
		mService = service.NewService(storage, logger)
		logger.Info("Database storage initialized")
	case cfg.FilePath != "":
		storage = repository.NewStorage()
		mService = service.NewService(storage, logger)
		initFileStorage(mService, logger)
		logger.Info("Local storage initialized")
	default:
		storage = repository.NewStorage()
		mService = service.NewService(storage, logger)
		logger.Info("In-memory storage initialized")
	}

	InitObservers(mService, logger)

	handler := handler.NewHandler(mService, cfg.KEY)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(handler.GetAllMetrics), handlersLogger))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(handler.PostMetricInfo), handlersLogger))
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
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(handler.PingDatabase), handlersLogger))
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(handler.UpdateMetricBatch), handlersLogger))
		})

	})

	pprofServer := &http.Server{
		Addr: cfg.PprofServer,
	}

	APIServer := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	signalctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Warnf("pprof server failed to start on %s: %w", cfg.PprofServer, err)
		}
	}()

	go func() {
		if err := APIServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server failed to start on %s: %w", cfg.Address, err))
		}
	}()

	<-signalctx.Done()
	shutDownCtx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	APIServer.Shutdown(shutDownCtx)
	pprofServer.Shutdown(shutDownCtx)
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
			logger.Error("Couldn't parse data")
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

func saveMetricsToFile(path string, service service.MetricsService, logger *zap.Logger) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error("Failed to open metrics file for writing", zap.Error(err))
		return
	}
	defer file.Close()

	gauges, err := service.GetAllGauges()
	if err != nil {
		logger.Error("Failed to get gauges", zap.Error(err))
		gauges = make(map[string]float64)
	}

	counters, err := service.GetAllCounters()
	if err != nil {
		logger.Error("Failed to get counters", zap.Error(err))
		counters = make(map[string]int64)
	}

	allMetrics := struct {
		Counters map[string]int64   `json:"counters"`
		Gauges   map[string]float64 `json:"gauges"`
	}{
		Counters: counters,
		Gauges:   gauges,
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

func initDBStorage(logger *zap.Logger) (*sql.DB, repository.Repository) {
	if err := migrations.RunMigration(cfg.DSN); err != nil {
		logger.Fatal("Error when starting migrations: %v", zap.Error(err))
	}
	logger.Info("Migration successfully started")

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		logger.Fatal("Database connection error:" + err.Error())
	}

	return db, repository.NewDBStorage(db, logger)
}

func initFileStorage(service service.MetricsService, logger *zap.Logger) {
	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Fatal("Couldn't create directory for storage file", zap.Error(err))
	}

	if cfg.Restore {
		loadMetricsFromFile(cfg.FilePath, service, logger)
	}

	if cfg.StoreInt == 0 {
		saveMetricsToFile(cfg.FilePath, service, logger)
	} else {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInt) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				saveMetricsToFile(cfg.FilePath, service, logger)
			}
		}()
	}
}

func InitObservers(service service.MetricsService, logger *zap.Logger) {
	if cfg.AuditFile != "" {
		fObs := &observer.FileObserver{
			FilePath: cfg.AuditFile,
			Logger:   logger,
		}

		service.RegisterObserver(fObs)
		logger.Info("file observer successfully registered in service")
	}

	if cfg.AuditURL != "" {
		httpObs := &observer.HTTPObserver{
			URL:    cfg.AuditURL,
			Logger: logger,
		}

		service.RegisterObserver(httpObs)
		logger.Info("http observer successfully registered in service")
	}
}
