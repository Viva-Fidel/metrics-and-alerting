package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Options struct {
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
}

// NewDB инициализирует подключение к базе данных с повторными попытками пинга.
func NewDB(dsn string, opts Options) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DATABASE_DSN")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(opts.MaxOpenConns)
	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	db.SetConnMaxIdleTime(opts.ConnMaxIdleTime)

	if err := pingWithRetry(db, 10, time.Second); err != nil {
		return nil, err
	}

	return db, nil
}

// pingWithRetry делает попытки подключения к БД несколько раз с задержкой.
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