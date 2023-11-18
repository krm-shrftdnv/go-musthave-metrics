package storage

import "database/sql"

type DBStorage[T Element] struct {
	*MemStorage[T]
	DB *sql.DB
}
