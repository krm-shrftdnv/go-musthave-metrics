package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
)

var cfg internal.Config

func parseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Int64Var(&cfg.StoreInterval, "i", 10, "store interval")
	absPath, err := filepath.Abs("../../tmp/metrics-db.json")
	if err != nil {
		absPath = "/tmp/metrics-db.json"
	}
	flag.StringVar(&cfg.FileStoragePath, "f", absPath, "file storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "restore from file")
	flag.StringVar(&cfg.DatabaseDsn, "d", "", "database dsn")
	flag.StringVar(&cfg.HashKey, "k", "", "hash key")
	flag.Parse()

	if err := godotenv.Load(".env", ".env.local"); err != nil {
		log.Println("No .env file found")
	}
	err = env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
}
