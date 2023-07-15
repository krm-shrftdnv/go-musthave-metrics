package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/handlers"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
)

var counterStorage = storage.MemStorage[internal.Counter]{}
var gaugeStorage = storage.MemStorage[internal.Gauge]{}

func run(handler http.Handler) error {
	fmt.Println("Running server on ", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, handler)
}

func main() {
	parseFlags()

	counterStorage.Init()
	gaugeStorage.Init()

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

	r := chi.NewRouter()

	r.Handle("/update/{metricType}/{metricName}/{metricValue}", &updateMetricHandler)
	r.Handle("/value/{metricType}/{metricName}", &metricStateHandler)
	r.Handle("/", &storageStateHandler)

	err := run(r)
	if err != nil {
		panic(err)
	}
}
