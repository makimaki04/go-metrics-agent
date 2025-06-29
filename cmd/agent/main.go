package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
)

type agentConfig struct {
	port string
	reportInterval time.Duration
	pollInterval time.Duration
}

func main() {
	var agentConfig agentConfig
	flag.StringVar(&agentConfig.port, "a", ":8080", "Server port")
	flag.DurationVar(&agentConfig.reportInterval, "r", 10 * time.Second, "Report interval (e.g. 10s, 30s)")
	flag.DurationVar(&agentConfig.pollInterval, "p", 2 * time.Second, "Poll interval (e.g. 2s, 10s)")
	flag.Parse()

	url := fmt.Sprintf(`http://%s`, agentConfig.port)
	storage := agent.NewLocalStorage()
	collector := agent.NewCollector(storage)
	client := resty.New()
	sender := agent.NewSender(client, url, storage)

	collectTicker := time.NewTicker(agentConfig.pollInterval)
	sendTicker := time.NewTicker(agentConfig.reportInterval)

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
