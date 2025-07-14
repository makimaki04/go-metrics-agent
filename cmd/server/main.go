package main

import (
	"net/http"

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
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", middleware.WithLogging(handler.GetAllMetrics, handlersLogger))
		r.Route("/value/{MType}/{ID}", func(r chi.Router) {
			r.Get("/",  middleware.WithLogging(handler.HandleReq, handlersLogger))
		})
		r.Route("/update/{MType}/{ID}/{value}", func(r chi.Router) {
			r.Post("/",  middleware.WithLogging(handler.HandleReq, handlersLogger))
		})
	})
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}
}