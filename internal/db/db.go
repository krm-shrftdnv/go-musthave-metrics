package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/logger"
)

const maxAttempts = 3

func Init(db *sql.DB, databaseDsn string) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	db, err := sql.Open("pgx", databaseDsn)
	i := 0
	var pgErr *pgconn.PgError
	for err != nil && errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException && i < maxAttempts {
		logger.Log.Warnf("error connecting to db: %v. waiting %d seconds\n", err, 2*i+1)
		time.Sleep(time.Duration(2*i+1) * time.Second)
		logger.Log.Infof("retrying: attempt %d\n", i+1)
		db, err = sql.Open("pgx", databaseDsn)
		i++
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Ping(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("database connection not specified")
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err := db.PingContext(ctx)
	i := 0
	var pgErr *pgconn.PgError
	for err != nil && errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException && i < maxAttempts {
		logger.Log.Warnf("error connecting to db: %v. waiting %d seconds\n", err, 2*i+1)
		time.Sleep(time.Duration(2*i+1) * time.Second)
		logger.Log.Infof("retrying: attempt %d\n", i+1)
		err = db.PingContext(ctx)
		i++
	}
	if err != nil {
		return err
	}
	return nil
}

func CreateTable(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS metrics (
			id VARCHAR PRIMARY KEY,
			mtype VARCHAR NOT NULL,
			delta INT DEFAULT NULL,
			mvalue FLOAT DEFAULT NULL
		)
	`)
	if err != nil {
		return err
	}
	return nil
}
