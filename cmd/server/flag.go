package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"

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
	flag.Parse()

	if err := godotenv.Load(".env", ".env.local"); err != nil {
		log.Println("No .env file found")
	}
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		cfg.StoreInterval, _ = strconv.ParseInt(envStoreInterval, 10, 64)
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		cfg.Restore, _ = strconv.ParseBool(envRestore)
	}
	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		cfg.DatabaseDsn = envDatabaseDsn
	}
}
