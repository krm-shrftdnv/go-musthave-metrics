package storage

import (
	"encoding/json"
	"errors"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
	errs "github.com/pkg/errors"
	"os"
	"sort"
)

var SingletonOperator *Operator

type Operator struct {
	GaugeStorage   *MemStorage[internal.Gauge]
	CounterStorage *MemStorage[internal.Counter]
}

func NewOperator(gs *MemStorage[internal.Gauge], cs *MemStorage[internal.Counter], fname string) *Operator {
	if SingletonOperator == nil {
		SingletonOperator = &Operator{
			GaugeStorage:   gs,
			CounterStorage: cs,
		}
	}
	if fname != "" {
		if err := SingletonOperator.LoadMetrics(fname); err != nil {
			logger.Log.Error(err)
		}
	}
	return SingletonOperator
}

func (o *Operator) GetAllMetrics() []serializer.Metrics {
	metrics := []serializer.Metrics{}
	counterStorage := o.CounterStorage.GetAll()
	keys := make([]string, 0, len(counterStorage))
	for k := range counterStorage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		c, _ := o.CounterStorage.Get(key)
		metrics = append(metrics, serializer.Metrics{
			ID:    key,
			MType: string(c.GetTypeName()),
			Delta: c,
		})
	}
	gaugeStorage := o.GaugeStorage.GetAll()
	keys = make([]string, 0, len(gaugeStorage))
	for k := range gaugeStorage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g, _ := o.GaugeStorage.Get(key)
		metrics = append(metrics, serializer.Metrics{
			ID:    key,
			MType: string(g.GetTypeName()),
			Value: g,
		})
	}
	return metrics
}

func (o *Operator) SaveAllMetrics(fname string) error {
	metrics := o.GetAllMetrics()
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errs.WithMessage(err, "failed to open file")
	}
	defer f.Close()
	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		return errs.WithMessagef(err, "failed to marshal metrics")
	}
	_, err = f.Write(metricsJson)
	if err != nil {
		return errs.WithMessagef(err, "failed to write to file")
	}
	return nil
}

func (o *Operator) LoadMetrics(fname string) error {
	var metrics []serializer.Metrics
	_, err := os.Open(fname)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file, err := os.Create(fname)
			if err != nil {
				return errs.WithMessage(err, "failed to create file")
			}
			err = file.Chmod(0666)
			if err != nil {
				return errs.WithMessage(err, "failed to chmod file")
			}
			err = o.SaveAllMetrics(fname)
			if err != nil {
				return err
			}

		} else {
			return errs.WithMessage(err, "failed to open file")
		}
	}
	metricsJson, err := os.ReadFile(fname)
	if err != nil {
		return errs.WithMessage(err, "failed to read file")
	}
	err = json.Unmarshal(metricsJson, &metrics)
	if err != nil {
		return errs.WithMessage(err, "failed to unmarshal metrics")
	}
	for _, m := range metrics {
		switch m.MType {
		case string(internal.GaugeName):
			o.GaugeStorage.Set(m.ID, *m.Value)
		case string(internal.CounterName):
			o.CounterStorage.Set(m.ID, *m.Delta)
		}
	}
	return nil
}
