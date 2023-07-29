package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/handlers"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
)

var counterStorage = storage.MemStorage[internal.Counter]{}
var gaugeStorage = storage.MemStorage[internal.Gauge]{}

func run(handler http.Handler) error {
	logger.Log.Infoln("Running server on ", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, handler)
}

func main() {
	Init()

	updateMetricHandler := handlers.UpdateMetricHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
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

	r.Handle("/update/{metricType}/{metricName}/{metricValue}", logger.RequestWithLogging(&updateMetricHandler))
	r.Handle("/update", logger.RequestWithLogging(&jsonUpdateMetricHandler))
	r.Handle("/value/{metricType}/{metricName}", logger.RequestWithLogging(&metricStateHandler))
	r.Handle("/value", logger.RequestWithLogging(&jsonMetricStateHandler))
	r.Handle("/json", logger.RequestWithLogging(&jsonStorageStateHandler))
	r.Handle("/", logger.RequestWithLogging(&storageStateHandler))

	err := run(r)
	if err != nil {
		panic(err)
	}
}
