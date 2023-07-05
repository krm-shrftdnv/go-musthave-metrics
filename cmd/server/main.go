package main

import (
	"encoding/json"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"net/http"
	"strconv"
	"strings"
)

type Metric string

const (
	Gauge   Metric = "gauge"
	Counter Metric = "counter"
)

var counterStorage = internal.MemStorage[int64]{}
var gaugeStorage = internal.MemStorage[float64]{}

func main() {
	counterStorage.Init()
	gaugeStorage.Init()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateMetric)
	mux.HandleFunc("/state/gauge", func(w http.ResponseWriter, r *http.Request) {
		gaugeJSON, err := json.Marshal(gaugeStorage.GetAll())
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(gaugeJSON)
	})
	mux.HandleFunc("/state/counter", func(w http.ResponseWriter, r *http.Request) {
		counterJSON, err := json.Marshal(counterStorage.GetAll())
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(counterJSON)
	})
	mux.HandleFunc(`/`, http.NotFound)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func updateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) != 4 {
		http.Error(w, "Wrong path", http.StatusNotFound)
		return
	}
	metricType := pathParts[1]
	key := pathParts[2]
	value := pathParts[3]
	switch Metric(metricType) {
	case Gauge:
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "Value should be float", http.StatusBadRequest)
			return
		}
		addGauge(key, value)
	case Counter:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Value should be int", http.StatusBadRequest)
			return
		}
		addCounter(key, value)
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
		return
	}
}

func addGauge(key string, value float64) {
	gaugeStorage.Set(key, value)
}

func addCounter(key string, value int64) {
	newValue := counterStorage.Get(key) + value
	counterStorage.Set(key, newValue)
}
