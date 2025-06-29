package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
)

func main() {
	storage := repository.NewStorage()
	service := service.NewService(storage)
	handler := handler.NewHandler(service)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllMetrics)
		r.Route("/value/{MType}/{ID}", func(r chi.Router) {
			r.Get("/", handler.HandleReq)
		})
		r.Route("/update/{MType}/{ID}/{value}", func(r chi.Router) {
			r.Post("/", handler.HandleReq)
		})
	})
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}