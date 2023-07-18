package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"log"
)

var cfg internal.Config

func parseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
}
