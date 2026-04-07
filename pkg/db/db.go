package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewDB(dsn string) (*sql.DB, error) {
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

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to init postgres migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to init migration source: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		sourceErr, dbErr := migrator.Close()
		if sourceErr != nil || dbErr != nil {
			// no-op: close errors are non-critical here
		}
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}