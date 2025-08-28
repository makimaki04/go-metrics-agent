package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
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
	sender := agent.NewSender(client, url, storage, cfg.Key)

	collectTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	defer func() {
		collectTicker.Stop()
		sendTicker.Stop()
	}()


	metricsCh := make(chan models.Metrics, 100)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func ()  {
		for {
			select {
			case <- collectTicker.C:
					collector.CollectRuntimeMetrics()
			case <- stop:
				return
			}
		}
	}()

	go func ()  {
		for {
			select {
			case <- collectTicker.C:
				collector.CollectSysMetrics()
			case <- stop:
				return
			}
		}
	}()

	var wg sync.WaitGroup

	go func ()  {
		for {
			select {
			case <- sendTicker.C:
				metrics := storage.GetAll()
				for _, m := range metrics {
					metricsCh <- m
				}
			case <- stop:
				close(metricsCh)
				return
		}
		}
	}()

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go worker(*sender, metricsCh, &wg)
	}
	
	<-stop
	wg.Wait()
	fmt.Println("Agent was shutdown")
}

func worker(sender agent.Sender, ch chan models.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	batch := make([]models.Metrics, 0, 10)
	for m := range ch {
		batch = append(batch, m)
		if len(batch) == cap(batch) {
			err := sender.SendMetricsBatch(batch)
			if err != nil {
				log.Printf("error sending data: %v", err)
			}
			batch = batch[:0]
		}
	}
	
	if len(batch) > 0 {
		err := sender.SendMetricsBatch(batch)
		if err != nil {
			log.Printf("error sending data: %v", err)
		}
	}
}
