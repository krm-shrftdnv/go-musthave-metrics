package main

import (
	"fmt"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	serverHost     = "http://localhost:8080"
)

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
var pollCount = internal.Metric[internal.Counter]{
	Name:  "PollCount",
	Value: 0,
}
var gaugeMetrics = map[string]*internal.Metric[internal.Gauge]{}

func main() {
	for _, metricName := range metricNames {
		gaugeMetrics[metricName] = &internal.Metric[internal.Gauge]{
			Name: metricName,
		}
	}
	for {
		poll()
		sendMetrics()
	}
}

func poll() {
	runtime.ReadMemStats(&m)
	pollCount.Value++
	for _, metricName := range metricNames {
		updateMetric(metricName)
	}
	time.Sleep(pollInterval * time.Second)
}

func updateMetric(name string) {
	metric, ok := gaugeMetrics[name]
	randNum := internal.Gauge(rand.Float64()) * 1000
	if !ok {
		return
	}
	switch name {
	case "Alloc":
		metric.Value = internal.Gauge(m.Alloc) / (1024 * 1024)
	case "BuckHashSys":
		metric.Value = internal.Gauge(m.BuckHashSys) / (1024 * 1024)
	case "Frees":
		metric.Value = internal.Gauge(m.Frees) / (1024 * 1024)
	case "GCCPUFraction":
		metric.Value = internal.Gauge(m.GCCPUFraction) / (1024 * 1024)
	case "GCSys":
		metric.Value = internal.Gauge(m.GCSys) / (1024 * 1024)
	case "HeapAlloc":
		metric.Value = internal.Gauge(m.HeapAlloc) / (1024 * 1024)
	case "HeapIdle":
		metric.Value = internal.Gauge(m.HeapIdle) / (1024 * 1024)
	case "HeapInuse":
		metric.Value = internal.Gauge(m.HeapInuse) / (1024 * 1024)
	case "HeapObjects":
		metric.Value = internal.Gauge(m.HeapObjects) / (1024 * 1024)
	case "HeapReleased":
		metric.Value = internal.Gauge(m.HeapReleased) / (1024 * 1024)
	case "HeapSys":
		metric.Value = internal.Gauge(m.HeapSys) / (1024 * 1024)
	case "LastGC":
		metric.Value = internal.Gauge(m.LastGC) / (1024 * 1024)
	case "Lookups":
		metric.Value = internal.Gauge(m.Lookups) / (1024 * 1024)
	case "MCacheInuse":
		metric.Value = internal.Gauge(m.MCacheInuse) / (1024 * 1024)
	case "MCacheSys":
		metric.Value = internal.Gauge(m.MCacheSys) / (1024 * 1024)
	case "MSpanInuse":
		metric.Value = internal.Gauge(m.MSpanInuse) / (1024 * 1024)
	case "MSpanSys":
		metric.Value = internal.Gauge(m.MSpanSys) / (1024 * 1024)
	case "Mallocs":
		metric.Value = internal.Gauge(m.Mallocs) / (1024 * 1024)
	case "NextGC":
		metric.Value = internal.Gauge(m.NextGC) / (1024 * 1024)
	case "NumForcedGC":
		metric.Value = internal.Gauge(m.NumForcedGC) / (1024 * 1024)
	case "NumGC":
		metric.Value = internal.Gauge(m.NumGC) / (1024 * 1024)
	case "OtherSys":
		metric.Value = internal.Gauge(m.OtherSys) / (1024 * 1024)
	case "PauseTotalNs":
		metric.Value = internal.Gauge(m.PauseTotalNs) / (1024 * 1024)
	case "StackInuse":
		metric.Value = internal.Gauge(m.StackInuse) / (1024 * 1024)
	case "StackSys":
		metric.Value = internal.Gauge(m.StackSys) / (1024 * 1024)
	case "Sys":
		metric.Value = internal.Gauge(m.Sys) / (1024 * 1024)
	case "TotalAlloc":
		metric.Value = internal.Gauge(m.TotalAlloc) / (1024 * 1024)
	case "RandomValue":
		metric.Value = randNum
	}
}

func sendMetrics() {
	client := http.Client{}
	for _, metricName := range metricNames {
		metric, ok := gaugeMetrics[metricName]
		if !ok {
			return
		}
		if metricName == "RandomValue" {
			_ = 1
		}

		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/update/%s/%s/%v", serverHost, metric.Value.GetTypeName(), metric.Name, metric.Value),
			nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Set("Content-Type", "text/plain")
		_, err = client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/update/%s/pollCount/%v", serverHost, pollCount.Value.GetTypeName(), pollCount.Value),
		nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	_, err = client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(reportInterval * time.Second)
}
