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
	Database string `env:"DATABASE_DSN"`
}

var cfg = config{
	Address:  ":8080",
	StoreInt: 300,
	FilePath: "./data/save.json",
	Restore:  false,
	Database: "host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable",
}

func setConfig() {
	flag.StringVar(&cfg.Address, "a", ":8080", "Server port")
	flag.IntVar(&cfg.StoreInt, "i", 300, "collect data to store interval in secodns")
	flag.StringVar(&cfg.FilePath, "f", "../../data/save.json", "storage file path")
	flag.BoolVar(&cfg.Restore, "r", false, "should load data from local file when starting the server")
	flag.StringVar(&cfg.Database, "d", "host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable", "databse connection string")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
