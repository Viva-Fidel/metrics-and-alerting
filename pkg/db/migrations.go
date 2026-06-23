package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var migrationUpPattern = regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)

type migration struct {
	name    string
	path    string
	version int
}

// RunMigrations применяет все новые миграции к базе данных
func RunMigrations(db *sql.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}

	dir, err := detectMigrationsDir()
	if err != nil {
		return err
	}

	files, err := loadUpMigrations(dir)
	if err != nil {
		return err
	}

	for _, m := range files {
		var alreadyApplied bool
		if err := db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			m.version,
		).Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("failed to check migration %d: %w", m.version, err)
		}
		if alreadyApplied {
			continue
		}

		sqlBytes, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("failed to read migration %q: %w", m.path, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start tx for migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations(version, name) VALUES ($1, $2)`,
			m.version,
			m.name,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to save migration version %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

// ensureSchemaMigrationsTable создаёт таблицу schema_migrations, если её нет
func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}
	return nil
}

// detectMigrationsDir автоматически определяет директорию с миграциями
func detectMigrationsDir() (string, error) {
	candidates := []string{"migrations", "./migrations", "../migrations", "../../migrations"}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}

// loadUpMigrations загружает и сортирует .up.sql миграции из указанной директории
func loadUpMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir %q: %w", dir, err)
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		match := migrationUpPattern.FindStringSubmatch(name)
		if len(match) != 2 {
			continue
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("bad migration version in %q: %w", name, err)
		}
		migrations = append(migrations, migration{
			version: version,
			name:    name,
			path:    filepath.Join(dir, name),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}
