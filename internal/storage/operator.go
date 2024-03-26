package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func NewOperator(ctx context.Context, gs Storage[internal.Gauge], cs Storage[internal.Counter], restore bool) (*Operator, error) {
	if SingletonOperator == nil {
		SingletonOperator = &Operator{
			GaugeStorage:   gs,
			CounterStorage: cs,
		}
	}
	if restore {
		if err := SingletonOperator.LoadMetrics(ctx); err != nil {
			return nil, err
		}
	}
	return SingletonOperator, nil
}

func (o *Operator) GetAllMetrics() []serializer.Metrics {
	metrics := make([]serializer.Metrics, 0)
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

func (o *Operator) SaveAllMetrics(ctx context.Context) error {
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		return o.saveAllMetricsToFile()
	case *DBStorage[internal.Counter]:
		return o.saveAllMetricsToDB(ctx)
	default:
		{
			logger.Log.Infoln("metrics will not be saved")
			return nil
		}
	}
}

func (o *Operator) LoadMetrics(ctx context.Context) error {
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		return o.loadMetricsFromFile()
	case *DBStorage[internal.Counter]:
		return o.loadMetricsFromDB(ctx)
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
	f, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return errs.WithMessagef(err, "failed to marshal metrics")
	}
	_, err = f.Write(metricsJSON)
	if err != nil {
		return errs.WithMessagef(err, "failed to write to file")
	}
	return f.Close()
}

func (o *Operator) saveAllMetricsToDB(ctx context.Context) error {
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
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	for _, m := range metrics {
		stmt, err := tx.PrepareContext(ctx, "SELECT id FROM metrics WHERE id = $1")
		if err != nil {
			return err
		}
		defer stmt.Close()
		rows, err := stmt.QueryContext(ctx, m.ID)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		var id string
		for rows.Next() {
			err = rows.Scan(&id)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		if id != "" {
			stmt, err = tx.PrepareContext(ctx, "UPDATE metrics SET mtype = $1, delta = $2, mvalue = $3 WHERE id = $4")
		} else {
			stmt, err = tx.PrepareContext(ctx, "INSERT INTO metrics (mtype, delta, mvalue, id) VALUES ($1, $2, $3, $4)")
		}
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.ExecContext(ctx, m.MType, m.Delta, m.Value, m.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) loadMetricsFromFile() error {
	var metrics []serializer.Metrics
	var absPath string
	switch o.CounterStorage.(type) {
	case *FileStorage[internal.Counter]:
		absPath = o.CounterStorage.(*FileStorage[internal.Counter]).FilePath
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	f, err := openFile(absPath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()
	metricsJSON, err := os.ReadFile(absPath)
	if err != nil {
		return errs.WithMessage(err, "failed to read file")
	}
	if (metricsJSON != nil) && (len(metricsJSON) != 0) {
		err = json.Unmarshal(metricsJSON, &metrics)
		if err != nil {
			return errs.WithMessage(err, "failed to unmarshal metrics")
		}
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

func (o *Operator) loadMetricsFromDB(ctx context.Context) error {
	var db *sql.DB
	switch o.CounterStorage.(type) {
	case *DBStorage[internal.Counter]:
		db = o.CounterStorage.(*DBStorage[internal.Counter]).DB
	default:
		return errs.WithMessage(errors.New("unsupported storage"), "unsupported storage")
	}
	metrics := make([]serializer.Metrics, 0)
	rows, err := db.QueryContext(ctx, "SELECT id, mtype, delta, mvalue FROM metrics")
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

func openFile(absPath string) (*os.File, error) {
	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(filepath.Dir(absPath), 0777)
			if err != nil {
				return nil, errs.WithMessage(err, "failed to create directory")
			}
			f, err = os.Create(absPath)
			if err != nil {
				return nil, errs.WithMessage(err, "failed to create file")
			}
		} else {
			return nil, errs.WithMessage(err, "failed to open file")
		}
	}
	return f, nil
}
