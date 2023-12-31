package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/compress/gzip"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/handlers"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var counterStorage = storage.MemStorage[internal.Counter]{}
var gaugeStorage = storage.MemStorage[internal.Gauge]{}

func run(handler http.Handler) error {
	logger.Log.Infoln("Running server on ", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, handler)
}

func saveMetrics(storeInterval int64) {
	for range time.Tick(time.Duration(storeInterval) * time.Second) {
		logger.Log.Infoln("Saving metrics to ", cfg.FileStoragePath)
		err := storage.SingletonOperator.SaveAllMetrics(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Errorln(err)
		}
	}
}

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	Init()
	updateMetricHandler := handlers.UpdateMetricHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	if cfg.StoreInterval == 0 {
		updateMetricHandler.FileStoragePath = cfg.FileStoragePath
	}
	storageStateHandler := handlers.StorageStateHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	metricStateHandler := handlers.MetricStateHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
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
	r.Route("/value", func(r chi.Router) {
		r.Handle("/", &jsonMetricStateHandler)
		r.Handle("/{metricType}/{metricName}", &metricStateHandler)
	})
	r.Route("/", func(r chi.Router) {
		r.Handle("/json", &jsonStorageStateHandler)
		r.Handle("/", &storageStateHandler)
	})

	go func() {
		err := run(r)
		if err != nil {
			panic(err)
		}
	}()

	if cfg.StoreInterval > 0 {
		go func() {
			saveMetrics(cfg.StoreInterval)
		}()
	}

	<-gracefulShutdown
	logger.Log.Infoln("Graceful shutdown")
	err := storage.SingletonOperator.SaveAllMetrics(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Errorln(err)
	}
}
