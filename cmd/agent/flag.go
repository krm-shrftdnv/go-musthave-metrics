package main

import (
	"flag"
)

var (
	serverAddress                string
	pollInterval, reportInterval int64
)

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "http://localhost:8080", "address and port to run server")
	flag.Int64Var(&pollInterval, "p", 2, "poll interval")
	flag.Int64Var(&reportInterval, "r", 10, "report interval")
	flag.Parse()
}
