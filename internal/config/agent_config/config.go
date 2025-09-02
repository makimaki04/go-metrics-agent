package agentconfig

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func SetConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", ":8080", "Server port")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key value")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Rate limit value")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}
