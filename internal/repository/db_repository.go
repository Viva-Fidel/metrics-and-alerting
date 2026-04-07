package repository

import (
	"database/sql"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (p *DBRepository) SetGauge(name string, value float64) {
	_, _ = p.db.Exec(
		`INSERT INTO metrics (id, type, value, delta, updated_at)
		 VALUES ($1, 'gauge', $2, NULL, NOW())
		 ON CONFLICT (id) DO UPDATE
		 SET type = 'gauge', value = EXCLUDED.value, delta = NULL, updated_at = NOW()`,
		name, value,
	)
}

func (p *DBRepository) AddCounter(name string, value int64) {
	_, _ = p.db.Exec(
		`INSERT INTO metrics (id, type, delta, value, updated_at)
		 VALUES ($1, 'counter', $2, NULL, NOW())
		 ON CONFLICT (id) DO UPDATE
		 SET type = 'counter', delta = metrics.delta + EXCLUDED.delta, value = NULL, updated_at = NOW()`,
		name, value,
	)
}

func (p *DBRepository) ApplyBatch(metrics []payload.MetricsJSON) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

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
			return err
		}
	}

	err = tx.Commit()
	return err
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
