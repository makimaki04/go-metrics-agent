package agentconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	Address        string `json:"address" env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `json:"poll_interval" env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `json:"crypto_key" env:"CRYPTO_KEY"`
	Config         string `env:"CONFIG"`
}

func SetConfig() Config {
	var cfg Config

	flagSet := flag.NewFlagSet("cfg", flag.ContinueOnError)
	flagSet.StringVar(&cfg.Config, "config", "", "config path")
	flagSet.Parse(os.Args[1:])

	cfgPath := os.Getenv("CONFIG")
	if cfgPath != "" {
		cfg.Config = cfgPath
	}

	if cfg.Config != "" {
		file, err := os.ReadFile(cfg.Config)
		if err != nil {
			log.Fatal("could't read config file", err)
		}
		err = json.Unmarshal(file, &cfg)
		if err != nil {
			log.Fatal("could't unmarshal config file", err)
		}
	}

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
