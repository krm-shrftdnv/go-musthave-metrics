package handlers

import (
	"github.com/go-chi/chi/v5"
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
	metricType := chi.URLParam(r, "metricType")
	key := chi.URLParam(r, "metricName")
	value := chi.URLParam(r, "metricValue")
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
	oldValue, ok := h.CounterStorage.Get(key)
	if !ok {
		h.CounterStorage.Set(key, value)
	} else {
		h.CounterStorage.Set(key, oldValue+value)
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
	sb := strings.Builder{}
	sb.WriteString(h.CounterStorage.String())
	sb.WriteString("\n")
	sb.WriteString(h.GaugeStorage.String())
	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type MetricStateHandler struct {
	GaugeStorage   *storage.MemStorage[internal.Gauge]
	CounterStorage *storage.MemStorage[internal.Counter]
}

func (h *MetricStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	metricType := chi.URLParam(r, "metricType")
	key := chi.URLParam(r, "metricName")
	var value string
	switch internal.MetricTypeName(metricType) {
	case internal.GaugeName:
		element, ok := h.GaugeStorage.Get(key)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		value = element.String()
	case internal.CounterName:
		element, ok := h.CounterStorage.Get(key)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		value = element.String()
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
		return
	}
	_, err := w.Write([]byte(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
