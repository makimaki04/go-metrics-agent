package agentconfig

import (
	"flag"

	loader "github.com/makimaki04/go-metrics-agent.git/internal/config"
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

func SetConfig() (Config, error) {
	cfg := Config{
		Address:        ":8080",
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "",
		RateLimit:      3,
		CryptoKey:      "",
		Config:         "",
	}

	var address string
	var repInt int
	var pollInt int
	var key string
	var rateLim int
	var cryptoKey string

	bind := func(fs *flag.FlagSet) {
		fs.StringVar(&address, "a", ":8080", "Server port")
		fs.IntVar(&repInt, "r", 10, "Report interval in seconds")
		fs.IntVar(&pollInt, "p", 2, "Poll interval in seconds")
		fs.StringVar(&key, "k", "", "Key value")
		fs.IntVar(&rateLim, "l", 3, "Rate limit value")
		fs.StringVar(&cryptoKey, "crypto-key", "", "crypto-key file path")
	}

	apply := func(name string) {
		switch name {
		case "a":
			cfg.Address = address
		case "r":
			cfg.ReportInterval = repInt
		case "p":
			cfg.PollInterval = pollInt
		case "k":
			cfg.Key = key
		case "l":
			cfg.RateLimit = rateLim
		case "crypto-key":
			cfg.CryptoKey = cryptoKey
		}
	}

	if err := loader.Load(&cfg, bind, apply); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
