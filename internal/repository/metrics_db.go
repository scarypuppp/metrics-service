package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/metrics-service/internal/model"
)

type DBMetricsStorage struct {
	Metrics map[string]models.Metrics
	dbObj   *sqlx.DB
}

func NewDBMetricsStorage(dbObj *sqlx.DB) *DBMetricsStorage {
	return &DBMetricsStorage{
		Metrics: make(map[string]models.Metrics),
		dbObj:   dbObj,
	}
}

func (s *DBMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics
	err := s.dbObj.SelectContext(ctx, &metrics, `
        SELECT id, mtype, delta, value, hash
        FROM metrics
        ORDER BY id`)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *DBMetricsStorage) GetMetricByName(ctx context.Context, id string) (*models.Metrics, error) {
	var m models.Metrics
	err := s.dbObj.GetContext(ctx, &m, `
        SELECT id, mtype, delta, value, hash
        FROM metrics
        WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get metric by name %s: %w", id, err)
	}
	return &m, nil
}

func (s *DBMetricsStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {
	tx, err := s.dbObj.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if metric == nil {
		return fmt.Errorf("nil metric received")
	}

	stmt, err := tx.PreparexContext(ctx, `
        INSERT INTO metrics (id, mtype, delta, value, hash)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id, mtype) DO UPDATE 
        SET 
            mtype = EXCLUDED.mtype,
            delta = EXCLUDED.delta,
            value = EXCLUDED.value,
            hash  = EXCLUDED.hash`)

	_, err = stmt.Exec(metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		return fmt.Errorf("update metric %s: %w", metric.ID, err)
	}

	return tx.Commit()
}

func (s *DBMetricsStorage) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	tx := s.dbObj.MustBegin()
	defer tx.Rollback()

	stmt, err := tx.PrepareNamedContext(ctx, `
	INSERT INTO metrics (id, mtype, delta, value, hash)
	VALUES (:id, :mtype, :delta, :value, :hash)
	ON CONFLICT (id, mtype) DO UPDATE 
	SET 
	    mtype = EXCLUDED.mtype,
		delta = EXCLUDED.delta,
		value = EXCLUDED.value,
		hash  = EXCLUDED.hash`)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, metrics)
	if err != nil {
		return err
	}
	return nil
}
