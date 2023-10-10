package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/db"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type UpdateMetricHandler struct {
	GaugeStorage    *storage.MemStorage[internal.Gauge]
	CounterStorage  *storage.MemStorage[internal.Counter]
	FileStoragePath string
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
		}
		h.addGauge(key, internal.Gauge(value))
	case internal.CounterName:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Value should be int", http.StatusBadRequest)
		}
		h.addCounter(key, internal.Counter(value))
	default:
		http.Error(w, "Metric type should be \"gauge\" or \"counter\"", http.StatusBadRequest)
	}
	if h.FileStoragePath != "" {
		err := storage.SingletonOperator.SaveAllMetrics(h.FileStoragePath)
		if err != nil {
			logger.Log.Errorln(err)
		}
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
	sb := strings.Builder{}
	sb.WriteString(h.CounterStorage.String())
	sb.WriteString("\n")
	sb.WriteString(h.GaugeStorage.String())
	w.Header().Set("Content-Type", "text/html")
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
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type JSONUpdateMetricHandler struct {
	UpdateMetricHandler
}

func (h *JSONUpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		value, ok := h.GaugeStorage.Get(metric.ID)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		metric.Value = value
	case internal.CounterName:
		h.addCounter(metric.ID, *metric.Delta)
		delta, ok := h.CounterStorage.Get(metric.ID)
		if !ok {
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}
		metric.Delta = delta
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

type JSONStorageStateHandler struct {
	StorageStateHandler
}

func (h *JSONStorageStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	metrics := storage.SingletonOperator.GetAllMetrics()
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

type JSONMetricStateHandler struct {
	MetricStateHandler
}

func (h *JSONMetricStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

type DBPingHandler struct {
	DB *sql.DB
}

func (h *DBPingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	err := db.Ping(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
}
