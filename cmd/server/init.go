package main

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
)

func Init() {
	parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	counterStorage.Init()
	gaugeStorage.Init()
	var fileStoragePath string
	if cfg.Restore {
		fileStoragePath = cfg.FileStoragePath
	}
	storage.SingletonOperator = storage.NewOperator(&gaugeStorage, &counterStorage, fileStoragePath)
}
