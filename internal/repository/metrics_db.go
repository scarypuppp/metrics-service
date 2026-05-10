package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/scarypuppp/metrics-service/internal/model"
)

type DBMetricsStorage struct {
	Metrics map[string]models.Metrics
	dbObj   *sql.DB
}

func NewDBMetricsStorage(dbObj *sql.DB) *DBMetricsStorage {
	return &DBMetricsStorage{
		Metrics: make(map[string]models.Metrics),
		dbObj:   dbObj,
	}
}

func (s *DBMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	rows, err := s.dbObj.QueryContext(ctx, `
	SELECT id, mtype, delta, value, hash
	FROM metrics
	ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]models.Metrics, 0)
	for rows.Next() {
		m, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return metrics, nil
}

func (s *DBMetricsStorage) GetMetricByName(ctx context.Context, id string) (*models.Metrics, error) {
	row := s.dbObj.QueryRowContext(ctx, `
        SELECT id, mtype, delta, value, hash
        FROM metrics
        WHERE id = $1`, id)
	m, err := scanMetric(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get metric by name %s: %w", id, err)
	}
	return &m, nil
}

func (s *DBMetricsStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {
	if metric == nil {
		return fmt.Errorf("nil metric received")
	}
	_, err := s.dbObj.ExecContext(ctx, `
        INSERT INTO metrics (id, mtype, delta, value, hash)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id, mtype) DO UPDATE 
        SET 
            mtype = EXCLUDED.mtype,
            delta = EXCLUDED.delta,
            value = EXCLUDED.value,
            hash  = EXCLUDED.hash`,
		metric.ID,
		metric.MType,
		metric.Delta,
		metric.Value,
		metric.Hash,
	)
	if err != nil {
		return fmt.Errorf("update metric %s: %w", metric.ID, err)
	}

	return nil
}

func scanMetric(scanner interface {
	Scan(dest ...any) error
}) (models.Metrics, error) {
	var m models.Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64

	if err := scanner.Scan(&m.ID, &m.MType, &delta, &value, &m.Hash); err != nil {
		return m, err
	}
	if delta.Valid {
		m.Delta = &delta.Int64
	}
	if value.Valid {
		m.Value = &value.Float64
	}
	return m, nil
}
