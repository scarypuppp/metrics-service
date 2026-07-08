package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/utils/retry"
)

type DBMetricsStorage struct {
	dbObj *sqlx.DB
}

func NewDBMetricsStorage(dbObj *sqlx.DB) *DBMetricsStorage {
	return &DBMetricsStorage{dbObj: dbObj}
}

func (s *DBMetricsStorage) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, _ := ctx.Value(txKey{}).(*sqlx.Tx); tx != nil {
		return tx
	}
	return s.dbObj
}

func (s *DBMetricsStorage) BeginTx(ctx context.Context) (context.Context, func(error) error, error) {
	tx, err := s.dbObj.BeginTxx(ctx, nil)
	if err != nil {
		return ctx, nil, err
	}
	txCtx := WithTx(ctx, tx)
	doneFn := func(err error) error {
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}
	return txCtx, doneFn, nil
}

func (s *DBMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics
	err := retry.Do(func() error {
		metrics = make([]models.Metrics, 0)
		return sqlx.SelectContext(ctx, s.getExecutor(ctx), &metrics, `
            SELECT id, mtype, delta, value, hash FROM metrics ORDER BY id`)
	}, retry.IsPgConnectionError)
	return metrics, err
}

func (s *DBMetricsStorage) GetMetricByName(ctx context.Context, id string) (*models.Metrics, error) {
	var m models.Metrics
	err := retry.Do(func() error {
		return sqlx.GetContext(ctx, s.getExecutor(ctx), &m, `
            SELECT id, mtype, delta, value, hash FROM metrics WHERE id = $1`, id)
	}, retry.IsPgConnectionError)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	const upsertQuery = `
        INSERT INTO metrics (id, mtype, delta, value, hash)
        VALUES (:id, :mtype, :delta, :value, :hash)
        ON CONFLICT (id, mtype) DO UPDATE
        SET mtype = EXCLUDED.mtype,
            delta = EXCLUDED.delta,
            value = EXCLUDED.value,
            hash  = EXCLUDED.hash`
	return retry.Do(func() error {
		_, err := sqlx.NamedExecContext(ctx, s.getExecutor(ctx), upsertQuery, metric)
		if err != nil {
			return fmt.Errorf("upsert metric %s: %w", metric.ID, err)
		}
		return nil
	}, retry.IsPgConnectionError)
}
