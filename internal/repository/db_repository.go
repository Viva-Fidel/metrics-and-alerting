package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBRepository struct {
	db *sql.DB
}

var dbRetryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (p *DBRepository) SetGauge(name string, value float64) {
	_ = withDBRetry(func() error {
		_, err := p.db.Exec(
			`INSERT INTO metrics (id, type, value, delta, updated_at)
			 VALUES ($1, 'gauge', $2, NULL, NOW())
			 ON CONFLICT (id) DO UPDATE
			 SET type = 'gauge', value = EXCLUDED.value, delta = NULL, updated_at = NOW()`,
			name, value,
		)
		return err
	})
}

func (p *DBRepository) AddCounter(name string, value int64) {
	_ = withDBRetry(func() error {
		_, err := p.db.Exec(
			`INSERT INTO metrics (id, type, delta, value, updated_at)
			 VALUES ($1, 'counter', $2, NULL, NOW())
			 ON CONFLICT (id) DO UPDATE
			 SET type = 'counter', delta = metrics.delta + EXCLUDED.delta, value = NULL, updated_at = NOW()`,
			name, value,
		)
		return err
	})
}

func (p *DBRepository) ApplyBatch(metrics []payload.MetricsJSON) error {
	if len(metrics) == 0 {
		return nil
	}

	return withDBRetry(func() error {
		tx, err := p.db.Begin()
		if err != nil {
			return err
		}

		for i := range metrics {
			item := metrics[i]

			switch item.MType {
			case "gauge":
				_, err = tx.Exec(
					`INSERT INTO metrics (id, type, value, delta, updated_at)
					 VALUES ($1, 'gauge', $2, NULL, NOW())
					 ON CONFLICT (id) DO UPDATE
					 SET type = 'gauge', value = EXCLUDED.value, delta = NULL, updated_at = NOW()`,
					item.ID, *item.Value,
				)
			case "counter":
				_, err = tx.Exec(
					`INSERT INTO metrics (id, type, delta, value, updated_at)
					 VALUES ($1, 'counter', $2, NULL, NOW())
					 ON CONFLICT (id) DO UPDATE
					 SET type = 'counter', delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta, value = NULL, updated_at = NOW()`,
					item.ID, *item.Delta,
				)
			}
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
		return nil
	})
}

func (p *DBRepository) GetGauge(name string) (float64, bool) {
	var result float64
	if err := p.db.QueryRow(`SELECT value FROM metrics WHERE id = $1 AND type = 'gauge'`, name).Scan(&result); err != nil {
		return 0, false
	}
	return result, true
}

func (p *DBRepository) GetCounter(name string) (int64, bool) {
	var result int64
	if err := p.db.QueryRow(`SELECT delta FROM metrics WHERE id = $1 AND type = 'counter'`, name).Scan(&result); err != nil {
		return 0, false
	}
	return result, true
}

func (p *DBRepository) GetAll() (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	rows, err := p.db.Query(`SELECT id, type, delta, value FROM metrics`)
	if err != nil {
		return gauges, counters
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id    string
			mType string
			delta sql.NullInt64
			value sql.NullFloat64
		)
		if err := rows.Scan(&id, &mType, &delta, &value); err != nil {
			continue
		}
		if mType == "gauge" && value.Valid {
			gauges[id] = value.Float64
		}
		if mType == "counter" && delta.Valid {
			counters[id] = delta.Int64
		}
	}

	return gauges, counters
}

func withDBRetry(fn func() error) error {
	for attempt := 0; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if attempt >= len(dbRetryDelays) || !isRetriablePGError(err) {
			return err
		}
		time.Sleep(dbRetryDelays[attempt])
	}
}

func isRetriablePGError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	if pgErr.Code == pgerrcode.UniqueViolation {
		return false
	}

	return strings.HasPrefix(pgErr.Code, "08")
}
