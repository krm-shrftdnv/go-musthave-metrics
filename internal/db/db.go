package db

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

func Init(db *sql.DB, databaseDsn string) *sql.DB {
	if db != nil {
		return db
	}
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		panic(err)
	}
	return db
}

func Ping(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func CreateTable(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
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
