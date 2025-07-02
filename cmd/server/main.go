package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
)

type serverConfig struct {
	port string
}

func main() {
	var serverConfig serverConfig
	flag.StringVar(&serverConfig.port, "a", ":8080", "Server port")
	flag.Parse()

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
	err := http.ListenAndServe(serverConfig.port, r)
	if err != nil {
		panic(err)
	}
}