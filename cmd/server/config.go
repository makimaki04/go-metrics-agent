package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	Address  string `env:"ADDRESS"`
	StoreInt int    `env:"STORE_INTERVAL"`
	FilePath string `env:"FILE_STORAGE_PATH"`
	Restore  bool   `env:"RESTORE"`
	DSN      string `env:"DATABASE_DSN"`
}

var cfg config

func setConfig() {
	flag.StringVar(&cfg.Address, "a", ":8080", "Server port")
	flag.IntVar(&cfg.StoreInt, "i", 300, "collect data to store interval in secodns")
	flag.StringVar(&cfg.FilePath, "f", "", "storage file path")
	flag.BoolVar(&cfg.Restore, "r", false, "should load data from local file when starting the server")
	flag.StringVar(&cfg.DSN, "d", "", "databse connection string")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}

// host=localhost port=5432 user=metrics_user password=password dbname=metrics_db sslmode=disable

//../../data/save.json


//"postgres://metrics_user:password@localhost:5432/metrics_db?sslmode=disable"