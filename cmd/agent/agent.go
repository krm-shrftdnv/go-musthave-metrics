package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/compress/gzip"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
)

type SafeMetricsMap struct {
	mx sync.RWMutex
	m  map[string]*internal.Metric[internal.Gauge]
}

func (s *SafeMetricsMap) Get(key string) (*internal.Metric[internal.Gauge], bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.m[key]
	return val, ok
}

func (s *SafeMetricsMap) Set(key string, value internal.Metric[internal.Gauge]) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[key] = &value
}

func (s *SafeMetricsMap) Update(key string, newValue internal.Gauge) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[key].Value = newValue
}

func (s *SafeMetricsMap) GetAll() map[string]*internal.Metric[internal.Gauge] {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.m
}

type SafeMetric struct {
	mx sync.RWMutex
	m  *internal.Metric[internal.Counter]
}

func (s *SafeMetric) Get() *internal.Metric[internal.Counter] {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.m
}

func (s *SafeMetric) Set(value internal.Metric[internal.Counter]) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m = &value
}

func (s *SafeMetric) Inc() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m.Value++
}

var m runtime.MemStats
var metricNames = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"RandomValue",
}
var pollCount = SafeMetric{
	m: &internal.Metric[internal.Counter]{
		Name:  "PollCount",
		Value: 0,
	},
}
var gaugeMetrics = &SafeMetricsMap{
	m: make(map[string]*internal.Metric[internal.Gauge]),
}
var client *resty.Client

func main() {
	parseFlags()

	for _, metricName := range metricNames {
		gaugeMetrics.Set(metricName, internal.Metric[internal.Gauge]{Name: metricName})
	}
	client = resty.New()
	go func() {
		poll()
	}()
	go func() {
		sendMetrics()
	}()
	select {}
}

func poll() {
	for range time.Tick(time.Duration(cfg.PollInterval) * time.Second) {
		runtime.ReadMemStats(&m)
		pollCount.Inc()
		for _, metricName := range metricNames {
			updateMetric(metricName)
		}
	}
}

func updateMetric(name string) {
	_, ok := gaugeMetrics.Get(name)
	randNum := internal.Gauge(rand.Float64()) * 1000
	if !ok {
		return
	}
	switch name {
	case "Alloc":
		gaugeMetrics.Update(name, internal.Gauge(m.Alloc)/(1024*1024))
	case "BuckHashSys":
		gaugeMetrics.Update(name, internal.Gauge(m.BuckHashSys)/(1024*1024))
	case "Frees":
		gaugeMetrics.Update(name, internal.Gauge(m.Frees)/(1024*1024))
	case "GCCPUFraction":
		gaugeMetrics.Update(name, internal.Gauge(m.GCCPUFraction)/(1024*1024))
	case "GCSys":
		gaugeMetrics.Update(name, internal.Gauge(m.GCSys)/(1024*1024))
	case "HeapAlloc":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapAlloc)/(1024*1024))
	case "HeapIdle":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapIdle)/(1024*1024))
	case "HeapInuse":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapInuse)/(1024*1024))
	case "HeapObjects":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapObjects)/(1024*1024))
	case "HeapReleased":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapReleased)/(1024*1024))
	case "HeapSys":
		gaugeMetrics.Update(name, internal.Gauge(m.HeapSys)/(1024*1024))
	case "LastGC":
		gaugeMetrics.Update(name, internal.Gauge(m.LastGC)/(1024*1024))
	case "Lookups":
		gaugeMetrics.Update(name, internal.Gauge(m.Lookups)/(1024*1024))
	case "MCacheInuse":
		gaugeMetrics.Update(name, internal.Gauge(m.MCacheInuse)/(1024*1024))
	case "MCacheSys":
		gaugeMetrics.Update(name, internal.Gauge(m.MCacheSys)/(1024*1024))
	case "MSpanInuse":
		gaugeMetrics.Update(name, internal.Gauge(m.MSpanInuse)/(1024*1024))
	case "MSpanSys":
		gaugeMetrics.Update(name, internal.Gauge(m.MSpanSys)/(1024*1024))
	case "Mallocs":
		gaugeMetrics.Update(name, internal.Gauge(m.Mallocs)/(1024*1024))
	case "NextGC":
		gaugeMetrics.Update(name, internal.Gauge(m.NextGC)/(1024*1024))
	case "NumForcedGC":
		gaugeMetrics.Update(name, internal.Gauge(m.NumForcedGC)/(1024*1024))
	case "NumGC":
		gaugeMetrics.Update(name, internal.Gauge(m.NumGC)/(1024*1024))
	case "OtherSys":
		gaugeMetrics.Update(name, internal.Gauge(m.OtherSys)/(1024*1024))
	case "PauseTotalNs":
		gaugeMetrics.Update(name, internal.Gauge(m.PauseTotalNs)/(1024*1024))
	case "StackInuse":
		gaugeMetrics.Update(name, internal.Gauge(m.StackInuse)/(1024*1024))
	case "StackSys":
		gaugeMetrics.Update(name, internal.Gauge(m.StackSys)/(1024*1024))
	case "Sys":
		gaugeMetrics.Update(name, internal.Gauge(m.Sys)/(1024*1024))
	case "TotalAlloc":
		gaugeMetrics.Update(name, internal.Gauge(m.TotalAlloc)/(1024*1024))
	case "RandomValue":
		gaugeMetrics.Update(name, randNum)
	}
}

func sendMetrics() {
	const maxAttempts = 3
	for range time.Tick(time.Duration(cfg.ReportInterval) * time.Second) {
		var metrics []serializer.Metrics
		gaugeMetricsMap := gaugeMetrics.GetAll()
		gaugeMetrics.mx.RLock()
		for k, v := range gaugeMetricsMap {
			metrics = append(metrics, serializer.Metrics{
				ID:    k,
				MType: string(v.Value.GetTypeName()),
				Value: &v.Value,
			})
		}
		pollCountMetric := pollCount.Get()
		metrics = append(metrics, serializer.Metrics{
			ID:    pollCountMetric.Name,
			MType: string(pollCountMetric.Value.GetTypeName()),
			Delta: &pollCountMetric.Value,
		})
		gaugeMetrics.mx.RUnlock()

		if len(metrics) == 0 {
			continue
		}

		req := gzip.CompressedRequest{
			Request: client.R(),
		}
		pollCount.mx.RLock()
		gaugeMetrics.mx.RLock()
		body, err := json.Marshal(metrics)
		pollCount.mx.RUnlock()
		gaugeMetrics.mx.RUnlock()
		if err != nil {
			log.Printf("error marshalling metrics: %v\n", err)
		}
		_, err = req.
			SetBody(body).
			SetHeader("Content-Type", "application/json; charset=utf-8").
			Post(fmt.Sprintf("http://%s/updates/", cfg.ServerAddress))
		i := 0
		for err != nil && i < maxAttempts {
			log.Printf("error sending metrics: %v. waiting %d seconds\n", err, 2*i+1)
			time.Sleep(time.Duration(2*i+1) * time.Second)
			log.Printf("retrying: attempt %d\n", i+1)
			_, err = req.
				SetBody(body).
				SetHeader("Content-Type", "application/json; charset=utf-8").
				Post(fmt.Sprintf("http://%s/updates/", cfg.ServerAddress))
			i++
		}
	}
}
