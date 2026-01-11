package agentconfig

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

func SetConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", ":8080", "Server port")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key value")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Rate limit value")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "crypto-key file path")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("could't parse config: %v", err)
	}

	return cfg
}
