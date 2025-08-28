package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

var cfg = config{
	Address:        ":8080",
	ReportInterval: 10,
	PollInterval:   2,
	Key:            "",
	RateLimit:      3,
}

func setConfig() {

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Server port")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Key value")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Rate limit value")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
