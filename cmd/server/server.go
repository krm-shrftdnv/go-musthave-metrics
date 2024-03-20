package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/compress/gzip"
	db2 "github.com/krm-shrftdnv/go-musthave-metrics/internal/db"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/handlers"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
)

func run(handler http.Handler) error {
	logger.Log.Infoln("Running server on ", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, handler)
}

func saveMetrics(ctx context.Context, storeInterval int64) {
	for range time.Tick(time.Duration(storeInterval) * time.Second) {
		err := storage.SingletonOperator.SaveAllMetrics(ctx)
		if err != nil {
			logger.Log.Errorln(err)
		}
	}
}

func main() {
	var counterStorage storage.Storage[internal.Counter]
	var gaugeStorage storage.Storage[internal.Gauge]
	var db *sql.DB
	ctx := context.TODO()

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
			db, err := db2.Init(db, cfg.DatabaseDsn)
			if err != nil {
				panic(err)
			}
			err = db2.CreateTable(ctx, db)
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
	operator, err := storage.NewOperator(ctx, gaugeStorage, counterStorage, cfg.Restore && cfg.FileStoragePath != "")
	if err != nil {
		panic(err)
	}
	storage.SingletonOperator = operator

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	updateMetricHandler := handlers.UpdateMetricHandler{
		GaugeStorage:   gaugeStorage,
		CounterStorage: counterStorage,
	}
	if cfg.StoreInterval == 0 {
		updateMetricHandler.FileStoragePath = cfg.FileStoragePath
	}
	storageStateHandler := handlers.StorageStateHandler{
		GaugeStorage:   gaugeStorage,
		CounterStorage: counterStorage,
	}
	metricStateHandler := handlers.MetricStateHandler{
		GaugeStorage:   gaugeStorage,
		CounterStorage: counterStorage,
	}
	jsonUpdateMetricHandler := handlers.JSONUpdateMetricHandler{
		UpdateMetricHandler: updateMetricHandler,
	}
	jsonMetricStateHandler := handlers.JSONMetricStateHandler{
		MetricStateHandler: metricStateHandler,
	}
	jsonStorageStateHandler := handlers.JSONStorageStateHandler{
		StorageStateHandler: storageStateHandler,
	}
	jsonUpdateMetricsHandler := handlers.JSONUpdateMetricsHandler{
		UpdateMetricHandler: updateMetricHandler,
	}
	dbPingHandler := handlers.DBPingHandler{
		DB: db,
	}

	r := chi.NewRouter()
	r.Use(
		middleware.StripSlashes,
		logger.RequestWithLogging,
		gzip.CompressRequestBody,
	)

	r.Route("/update", func(r chi.Router) {
		r.Handle("/", &jsonUpdateMetricHandler)
		r.Handle("/{metricType}/{metricName}/{metricValue}", &updateMetricHandler)
	})
	r.Route("/updates", func(r chi.Router) {
		r.Handle("/", &jsonUpdateMetricsHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Handle("/", &jsonMetricStateHandler)
		r.Handle("/{metricType}/{metricName}", &metricStateHandler)
	})
	r.Route("/", func(r chi.Router) {
		r.Handle("/json", &jsonStorageStateHandler)
		r.Handle("/", &storageStateHandler)
	})
	r.Route("/ping", func(r chi.Router) {
		r.Handle("/", &dbPingHandler)
	})

	go func() {
		err := run(r)
		if err != nil {
			panic(err)
		}
	}()

	if cfg.StoreInterval > 0 && cfg.FileStoragePath != "" {
		go func() {
			saveMetrics(ctx, cfg.StoreInterval)
		}()
	}

	<-gracefulShutdown
	logger.Log.Infoln("Graceful shutdown")
	err = storage.SingletonOperator.SaveAllMetrics(ctx)
	if err != nil {
		logger.Log.Errorln(err)
	}
}
