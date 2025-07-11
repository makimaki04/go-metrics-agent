package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
)

func main() {
	setConfig()
	if strings.HasPrefix(cfg.Address, ":") {
		cfg.Address = "localhost" + cfg.Address
	}
	url := fmt.Sprintf(`http://%s`, cfg.Address)
	storage := agent.NewLocalStorage()
	collector := agent.NewCollector(storage)
	client := resty.New()
	sender := agent.NewSender(client, url, storage)

	collectTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	defer func() {
		collectTicker.Stop()
		sendTicker.Stop()
	}()

	for {
		select {
		case <-collectTicker.C:
			collector.CollectMetrics()
		case <-sendTicker.C:
			err := sender.SendMetrics()
			if err != nil {
				log.Printf("error sending data: %v", err)
			} else {
				collector.ResetPollCount()
			}
		}
	}
}
