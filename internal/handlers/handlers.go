package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
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
	var metric serializer.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch internal.MetricTypeName(metric.MType) {
	case internal.GaugeName:
		h.addGauge(metric.ID, *metric.Value)
	case internal.CounterName:
		h.addCounter(metric.ID, *metric.Delta)
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UpdateMetricHandler) addGauge(key string, value internal.Gauge) {
	h.GaugeStorage.Set(key, value)
}

func (h *UpdateMetricHandler) addCounter(key string, value internal.Counter) {
	oldValue, ok := h.CounterStorage.Get(key)
	if !ok {
		h.CounterStorage.Set(key, value)
	} else {
		h.CounterStorage.Set(key, *oldValue+value)
	}
}

type StorageStateHandler struct {
	GaugeStorage   *storage.MemStorage[internal.Gauge]
	CounterStorage *storage.MemStorage[internal.Counter]
}

func (h *StorageStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	metrics := []serializer.Metrics{}
	for id, c := range h.CounterStorage.GetAll() {
		metrics = append(metrics, serializer.Metrics{
			ID:    id,
			MType: string(c.GetTypeName()),
			Delta: c,
		})
	}
	gaugeStorage := h.GaugeStorage.GetAll()
	for id, g := range gaugeStorage {
		metrics = append(metrics, serializer.Metrics{
			ID:    id,
			MType: string(g.GetTypeName()),
			Value: g,
		})
	}
	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type MetricStateHandler struct {
	GaugeStorage   *storage.MemStorage[internal.Gauge]
	CounterStorage *storage.MemStorage[internal.Counter]
}

func (h *MetricStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	var metric serializer.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch internal.MetricTypeName(metric.MType) {
	case internal.GaugeName:
		element, ok := h.GaugeStorage.Get(metric.ID)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		metric.Value = element
	case internal.CounterName:
		element, ok := h.CounterStorage.Get(metric.ID)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		metric.Delta = element
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
