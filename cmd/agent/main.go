package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/makimaki04/go-metrics-agent.git/internal/agent"
	agentconfig "github.com/makimaki04/go-metrics-agent.git/internal/config/agent_config"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

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
	
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	agent.Run()
}
