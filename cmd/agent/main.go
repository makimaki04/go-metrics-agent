package main

import (
	"net/http"
	"time"

	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
)

func main() {
	storage := agent.NewLocalStorage()
	collector := agent.NewCollector(storage)
	client := &http.Client{}
	url := "http://localhost:8080"
	sender := agent.NewSender(client, url)

	go func () {
		for {
			collector.CollcetMetrics()
			time.Sleep(2 * time.Second)
		}
	}()
	
	go func() {
		for {
			metrics := storage.GetAll()
			sender.SendMetrics(metrics)
			time.Sleep(10 * time.Second)
		}
	}()
	
	select {}
	
}
