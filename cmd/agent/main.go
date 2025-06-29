package main

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
)

func main() {
	storage := agent.NewLocalStorage()
	collector := agent.NewCollector(storage)
	client := resty.New()
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
		case <-collectTicker.C:
			collector.CollectMetrics()
		case <-sendTicker.C:
			sender.SendMetrics()
			collector.ResetPollCount()
		}
	}
}
