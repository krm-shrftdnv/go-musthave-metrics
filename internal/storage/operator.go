package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
	errs "github.com/pkg/errors"
)

var SingletonOperator *Operator

type Operator struct {
	GaugeStorage   Storage[internal.Gauge]
	CounterStorage Storage[internal.Counter]
}

func NewOperator(gs Storage[internal.Gauge], cs Storage[internal.Counter], restore bool) *Operator {
	if SingletonOperator == nil {
		SingletonOperator = &Operator{
			GaugeStorage:   gs,
			CounterStorage: cs,
		}
	}
	if restore {
		if err := SingletonOperator.LoadMetrics(); err != nil {
			logger.Log.Error(err)
		}
	}
	return SingletonOperator
}

func (o *Operator) GetAllMetrics() []serializer.Metrics {
	var metrics []serializer.Metrics
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

func (o *Operator) SaveAllMetrics() error {
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		return o.saveAllMetricsToFile()
	case *DBStorage[internal.Counter]:
		return o.saveAllMetricsToDB()
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
}

func (o *Operator) LoadMetrics() error {
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		return o.loadMetricsFromFile()
	case *DBStorage[internal.Counter]:
		return o.loadMetricsFromDB()
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
}

func (o *Operator) saveAllMetricsToFile() error {
	var filePath string
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		filePath = o.CounterStorage.(*FileStorage[internal.Counter]).FilePath
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	logger.Log.Infoln("Saving metrics to ", filePath)
	metrics := o.GetAllMetrics()
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errs.WithMessage(err, "failed to open file")
	}
	defer f.Close()
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return errs.WithMessagef(err, "failed to marshal metrics")
	}
	_, err = f.Write(metricsJSON)
	if err != nil {
		return errs.WithMessagef(err, "failed to write to file")
	}
	return nil
}

func (o *Operator) saveAllMetricsToDB() error {
	var db *sql.DB
	switch o.CounterStorage.(type) {
	case *DBStorage[internal.Counter]:
		db = o.CounterStorage.(*DBStorage[internal.Counter]).DB
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	logger.Log.Infoln("Saving metrics to DB")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	metrics := o.GetAllMetrics()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for _, m := range metrics {
		stmt, err := tx.PrepareContext(ctx, "SELECT id FROM metrics WHERE id = @id")
		if err != nil {
			return err
		}
		defer stmt.Close()
		row, err := stmt.QueryContext(ctx, sql.Named("id", m.ID))
		if err != nil {
			return err
		}
		var id string
		err = row.Scan(&id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if id != "" {
			stmt, err = tx.PrepareContext(ctx, "UPDATE metrics SET mtype = @mtype, delta = @delta, mvalue = @mvalue WHERE id = @id")
		} else {
			stmt, err = tx.PrepareContext(ctx, "INSERT INTO metrics (id, mtype, delta, mvalue) VALUES (@id, @mtype, @delta, @mvalue)")
		}
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.ExecContext(ctx, sql.Named("id", m.ID), sql.Named("mtype", m.MType), sql.Named("delta", m.Delta), sql.Named("mvalue", m.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) loadMetricsFromFile() error {
	var metrics []serializer.Metrics
	var filePath string
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		filePath = o.CounterStorage.(*FileStorage[internal.Counter]).FilePath
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	_, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file, err := os.Create(filePath)
			if err != nil {
				return errs.WithMessage(err, "failed to create file")
			}
			err = file.Chmod(0666)
			if err != nil {
				return errs.WithMessage(err, "failed to chmod file")
			}
			err = o.SaveAllMetrics()
			if err != nil {
				return err
			}

		} else {
			return errs.WithMessage(err, "failed to open file")
		}
	}
	metricsJSON, err := os.ReadFile(filePath)
	if err != nil {
		return errs.WithMessage(err, "failed to read file")
	}
	err = json.Unmarshal(metricsJSON, &metrics)
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

func (o *Operator) loadMetricsFromDB() error {
	var db *sql.DB
	switch o.CounterStorage.(type) {
	case *DBStorage[internal.Counter]:
		db = o.CounterStorage.(*DBStorage[internal.Counter]).DB
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	metrics := make([]serializer.Metrics, 0)
	rows, err := db.QueryContext(context.Background(), "SELECT id, mtype, delta, mvalue FROM metrics")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m serializer.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
		if err != nil {
			return err
		}
		metrics = append(metrics, m)
	}
	if rows.Err() != nil {
		return rows.Err()
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
