package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DATABASE_DSN")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := pingWithRetry(db, 10, time.Second); err != nil {
		return nil, err
	}

	return db, nil
}

func pingWithRetry(db *sql.DB, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("database ping failed after %d attempts: %w", attempts, lastErr)
}