package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	Address        string `env:"ADDRESS"`
}

var cfg = config{
	Address:        ":8080",
}

func setConfig() {
	flag.StringVar(&cfg.Address, "a", ":8080", "Server port")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
