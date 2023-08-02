package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"log"
)

var cfg internal.Config

func parseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Int64Var(&cfg.StoreInterval, "i", 30, "store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "restore from file")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
}
