package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
)

type agentConfig struct {
	port string
	reportInterval int
	pollInterval int
}

func main() {
	var agentConfig agentConfig
	flag.StringVar(&agentConfig.port, "a", "localhost:8080", "Server port")
	flag.IntVar(&agentConfig.reportInterval, "r", 10, "Report interval in seconds")
	flag.IntVar(&agentConfig.pollInterval, "p", 2, "Poll interval in seconds")
	flag.Parse()

	url := fmt.Sprintf(`http://%s`, agentConfig.port)
	storage := agent.NewLocalStorage()
	collector := agent.NewCollector(storage)
	client := resty.New()
	sender := agent.NewSender(client, url, storage)

	collectTicker := time.NewTicker(time.Duration(agentConfig.pollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(agentConfig.reportInterval) * time.Second)

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
