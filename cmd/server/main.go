package main

import (
	"net/http"

	"github.com/makimaki04/go-metrics-agent.git/internal/handler"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
)

func main() {
	storage := repository.NewStorage()
	service := service.NewService(storage)
	handler := handler.NewHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, handler.HandleReq)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
