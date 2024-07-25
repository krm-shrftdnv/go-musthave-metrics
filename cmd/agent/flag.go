package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
)

var cfg internal.Config

func parseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.Parse()

	if err := godotenv.Load(".env", ".env.local"); err != nil {
		log.Println("No .env file found")
	}
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		cfg.PollInterval, _ = strconv.ParseInt(envPollInterval, 10, 64)
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		cfg.ReportInterval, _ = strconv.ParseInt(envReportInterval, 10, 64)
	}
}
