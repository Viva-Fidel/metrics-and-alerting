package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDb(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DATABASE_DSN")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Проверим соединение сразу
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}