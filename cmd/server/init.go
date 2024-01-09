package main

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	db2 "github.com/krm-shrftdnv/go-musthave-metrics/internal/db"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
)

func init() {
	parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	counterMemStorage := &storage.MemStorage[internal.Counter]{}
	gaugeMemStorage := &storage.MemStorage[internal.Gauge]{}
	counterMemStorage.Init()
	gaugeMemStorage.Init()
	switch {
	case cfg.DatabaseDsn != "":
		{
			db = db2.Init(db, cfg.DatabaseDsn)
			err := db2.CreateTable(db)
			if err != nil {
				panic(err)
			}
			counterStorage = &storage.DBStorage[internal.Counter]{
				MemStorage: counterMemStorage,
				DB:         db,
			}
			gaugeStorage = &storage.DBStorage[internal.Gauge]{
				MemStorage: gaugeMemStorage,
				DB:         db,
			}
			break
		}
	case cfg.Restore && cfg.FileStoragePath != "":
		{
			counterStorage = &storage.FileStorage[internal.Counter]{
				MemStorage: counterMemStorage,
				FilePath:   cfg.FileStoragePath,
			}
			gaugeStorage = &storage.FileStorage[internal.Gauge]{
				MemStorage: gaugeMemStorage,
				FilePath:   cfg.FileStoragePath,
			}
			break
		}
	default:
		{
			counterStorage = &storage.MemStorage[internal.Counter]{}
			gaugeStorage = &storage.MemStorage[internal.Gauge]{}
		}
	}
	storage.SingletonOperator = storage.NewOperator(gaugeStorage, counterStorage, cfg.Restore)
}
