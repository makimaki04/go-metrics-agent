package main

import (
	"flag"
	"fmt"

	"github.com/makimaki04/go-metrics-agent.git/internal/config"
)

type Config struct {
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

func setConfig() (Config, error) {
	cfg := Config{
		Address:     ":8080",
		StoreInt:    300,
		FilePath:    "",
		Restore:     false,
		DSN:         "",
		KEY:         "",
		AuditFile:   "",
		AuditURL:    "",
		PprofServer: ":6060",
		CryptoKey:   "",
	}

	var address string
	var storeInt int
	var filePath string
	var restore bool
	var dsn string
	var key string
	var auditFile string
	var auditURL string
	var pprof string
	var cryptoKey string

	bind := func(fs *flag.FlagSet) {
		fs.StringVar(&address, "a", ":8080", "Server port")
		fs.IntVar(&storeInt, "i", 300, "collect data to store interval in secodns")
		fs.StringVar(&filePath, "f", "", "storage file path")
		fs.BoolVar(&restore, "r", false, "should load data from local file when starting the server")
		fs.StringVar(&dsn, "d", "", "databse connection string")
		fs.StringVar(&key, "k", "", "key value")
		fs.StringVar(&auditFile, "audit-file", "", "audit file address")
		fs.StringVar(&auditURL, "audit-url", "", "audit url")
		fs.StringVar(&pprof, "p", ":6060", "pprof server port")
		fs.StringVar(&cryptoKey, "crypto-key", "", "crypto-key file path")
	}

	apply := func(name string) {
		switch name {
		case "a":
			cfg.Address = address
		case "i":
			cfg.StoreInt = storeInt
		case "f":
			cfg.FilePath = filePath
		case "r":
			cfg.Restore = restore
		case "d":
			cfg.DSN = dsn
		case "k":
			cfg.KEY = key
		case "audit-file":
			cfg.AuditFile = auditFile
		case "audit-url":
			cfg.AuditURL = auditURL
		case "p":
			cfg.PprofServer = pprof
		case "crypto-key":
			cfg.CryptoKey = cryptoKey
		}
	}

	if err := config.Load(&cfg, bind, apply); err != nil {
		return Config{}, fmt.Errorf("could't load config: %v", err)
	}

	return cfg, nil
}

// host=localhost port=5432 user=metrics_user password=password dbname=metrics_db sslmode=disable

//../../data/save.json

//"postgres://metrics_user:password@localhost:5432/metrics_db?sslmode=disable"

//../../data/audit_file.json
