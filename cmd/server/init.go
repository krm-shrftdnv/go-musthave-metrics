package main

import "github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"

func Init() {
	parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	counterStorage.Init()
	gaugeStorage.Init()
}
