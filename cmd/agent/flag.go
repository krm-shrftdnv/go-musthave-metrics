package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
)

var cfg internal.Config

func parseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.StringVar(&cfg.HashKey, "k", "", "hash key")
	flag.Parse()

	if err := godotenv.Load(".env", ".env.local"); err != nil {
		log.Println("No .env file found")
	}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
}
