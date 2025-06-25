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
	sender := agent.NewSender(client, url, storage)

	collectTicker := time.NewTicker(2 * time.Second)
	sendTicker := time.NewTicker(10 * time.Second)

	defer func() {
		collectTicker.Stop()
		sendTicker.Stop()
	}()

	for {
		select {
		case <- collectTicker.C:
			collector.CollcetMetrics()
		case <- sendTicker.C:
			sender.SendMetrics()
		}
	}
}
