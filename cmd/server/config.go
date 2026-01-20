package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type config struct {
	Address     string `json:"address" env:"ADDRESS"`
	StoreInt    int    `json:"store_interval" env:"STORE_INTERVAL"`
	FilePath    string `env:"FILE_STORAGE_PATH"`
	Restore     bool   `json:"restore" env:"RESTORE"`
	DSN         string `json:"database_dsn" env:"DATABASE_DSN"`
	KEY         string `env:"KEY"`
	AuditFile   string `json:"store_file" env:"AUDIT_FILE"`
	AuditURL    string `env:"AUDIT_URL"`
	PprofServer string `env:"PPROF_SERVER"`
	CryptoKey   string `json:"crypto_key" env:"CRYPTO_KEY"`
	Config      string `env:"CONFIG"`
}

var cfg config

func setConfig() {
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
	flag.IntVar(&cfg.StoreInt, "i", 300, "collect data to store interval in secodns")
	flag.StringVar(&cfg.FilePath, "f", "", "storage file path")
	flag.BoolVar(&cfg.Restore, "r", false, "should load data from local file when starting the server")
	flag.StringVar(&cfg.DSN, "d", "", "databse connection string")
	flag.StringVar(&cfg.KEY, "k", "", "key value")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "audit file address")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "audit url")
	flag.StringVar(&cfg.PprofServer, "p", ":6060", "pprof server port")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "crypto-key file path")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("could't parse config: %v", err)
	}
}

// host=localhost port=5432 user=metrics_user password=password dbname=metrics_db sslmode=disable

//../../data/save.json

//"postgres://metrics_user:password@localhost:5432/metrics_db?sslmode=disable"

//../../data/audit_file.json
