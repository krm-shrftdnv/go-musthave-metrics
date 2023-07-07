package main

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/handlers"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
)

var counterStorage = storage.MemStorage[internal.Counter]{}
var gaugeStorage = storage.MemStorage[internal.Gauge]{}

func main() {
	counterStorage.Init()
	gaugeStorage.Init()

	updateMetricHandler := handlers.UpdateMetricHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	gaugeStorageStateHandler := handlers.StorageStateHandler[internal.Gauge]{
		Storage: &gaugeStorage,
	}
	counterStorageStateHandler := handlers.StorageStateHandler[internal.Counter]{
		Storage: &counterStorage,
	}

	mux := http.NewServeMux()
	mux.Handle("/update/", &updateMetricHandler)
	mux.Handle("/state/gauge", &gaugeStorageStateHandler)
	mux.Handle("/state/counter", &counterStorageStateHandler)
	mux.HandleFunc(`/`, http.NotFound)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
