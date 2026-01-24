package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

type BindFlags func(fs *flag.FlagSet)
type ApplyVisitedFlag func(name string)

func Load(
	cfg any,
	bind BindFlags,
	apply ApplyVisitedFlag,
) error {
	fs := flag.NewFlagSet("cfg", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "config", "", "config path")

	bind(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	path := configPath
	if path == "" {
		path = os.Getenv("CONFIG")
	}

	if path != "" {
		file, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("could't read config file: %v", err)
		}
		if err := json.Unmarshal(file, cfg); err != nil {
			return err
		}
	}

	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			return
		}

		apply(f.Name)
	})

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("could't parse config: %v", err)
	}

	return nil
}
