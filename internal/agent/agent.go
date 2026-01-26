package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	agentconfig "github.com/makimaki04/go-metrics-agent.git/internal/config/agent_config"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

//Agent - struct for the agent
type Agent struct {
	cfg       agentconfig.Config
	storage   *LocalStorage
	collector *Collector
	sender    *Sender

	collectTicker *time.Ticker
	sendTicker    *time.Ticker
	metricsCh     chan models.Metrics
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

//NewAgent - method for creating a new agent
//create a new agent
func NewAgent(cfg agentconfig.Config) *Agent {
	url := fmt.Sprintf(`http://%s`, cfg.Address)
	storage := NewLocalStorage()
	collector := NewCollector(storage)
	client := resty.New()
	sender, err := NewSender(client, url, storage, cfg.Key, cfg.CryptoKey)
	if err != nil {
		fmt.Printf("load public key error: %v, continuing without encryption", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		cfg:           cfg,
		storage:       storage,
		collector:     collector,
		sender:        sender,
		collectTicker: time.NewTicker(time.Duration(cfg.PollInterval) * time.Second),
		sendTicker:    time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second),
		metricsCh:     make(chan models.Metrics, cfg.RateLimit),
		ctx:           ctx,
		cancel:        cancel,
	}
}

//Run - method for running the agent
//run the agent in separate goroutines
func (a *Agent) Run() {
	go a.runRuntimeCollector()
	go a.runSysCollector()
	go a.runSender()

	for i := 0; i < a.cfg.RateLimit; i++ {
		a.wg.Add(1)
		go a.worker()
	}

	<-a.ctx.Done()
	a.wg.Wait()
	fmt.Println("Agent was shutdown")
}

//Stop - method for stopping the agent
//stop the agent
func (a *Agent) Stop() {
	a.cancel()
	a.collectTicker.Stop()
	a.sendTicker.Stop()
}

//runRuntimeCollector - method for running the runtime collector
//run the runtime metrics collector
func (a *Agent) runRuntimeCollector() {
	for {
		select {
		case <-a.collectTicker.C:
			a.collector.CollectRuntimeMetrics()
		case <-a.ctx.Done():
			return
		}
	}
}

//runSysCollector - method for running the system collector
//run the system metrics collector
func (a *Agent) runSysCollector() {
	for {
		select {
		case <-a.collectTicker.C:
			a.collector.CollectSysMetrics()
		case <-a.ctx.Done():
			return
		}
	}
}

//runSender - method for running the sender
//run the sender
func (a *Agent) runSender() {
	defer close(a.metricsCh)
	for {
		select {
		case <-a.sendTicker.C:
			metrics := a.storage.GetAll()
			for _, m := range metrics {
				a.metricsCh <- m
			}
			a.collector.ResetPollCount()
		case <-a.ctx.Done():
			return
		}
	}
}

//worker - method for running the worker
//run the worker that collects metrics and sends them to the server
func (a *Agent) worker() {
	defer a.wg.Done()

	batch := make([]models.Metrics, 0, 10)
	for m := range a.metricsCh {
		batch = append(batch, m)
		if len(batch) == cap(batch) {
			err := a.sender.SendMetricsBatch(batch)
			if err != nil {
				log.Printf("error sending data: %v", err)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		err := a.sender.SendMetricsBatch(batch)
		if err != nil {
			log.Printf("error sending data: %v", err)
		}
	}
}
