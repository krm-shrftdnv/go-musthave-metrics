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
