package handlers

import (
	"encoding/json"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type UpdateMetricHandler struct {
	GaugeStorage   *storage.MemStorage[internal.Gauge]
	CounterStorage *storage.MemStorage[internal.Counter]
}

func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	switch internal.MetricTypeName(metricType) {
	case internal.GaugeName:
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "Value should be float", http.StatusBadRequest)
			return
		}
		h.addGauge(key, internal.Gauge(value))
	case internal.CounterName:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Value should be int", http.StatusBadRequest)
			return
		}
		h.addCounter(key, internal.Counter(value))
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
		return
	}
}

func (h *UpdateMetricHandler) addGauge(key string, value internal.Gauge) {
	h.GaugeStorage.Set(key, value)
}

func (h *UpdateMetricHandler) addCounter(key string, value internal.Counter) {
	newValue := h.CounterStorage.Get(key) + value
	h.CounterStorage.Set(key, newValue)
}

type StorageStateHandler[T storage.Element] struct {
	Storage *storage.MemStorage[T]
}

func (h *StorageStateHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	storageJSON, err := json.Marshal(h.Storage.GetAll())
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(storageJSON)
}
