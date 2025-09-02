package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
	agentconfig "github.com/makimaki04/go-metrics-agent.git/internal/config/agent_config"
)

func main() {
	cfg := agentconfig.SetConfig()
	if strings.HasPrefix(cfg.Address, ":") {
		cfg.Address = "localhost" + cfg.Address
	}

	agent := agent.NewAgent(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		agent.Stop()
	}()
	
	agent.Run()
}
