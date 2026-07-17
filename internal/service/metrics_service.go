package service

import (
	"context"
	"errors"
	"sort"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
)

// Errors returned by MetricService operations.
var (
	ErrInvalidMetricType  = errors.New("invalid metric type provided")
	ErrMetricTypeMismatch = errors.New("metric with such name already exists with different type")
	ErrMetricNameNotExist = errors.New("metric with such name does not exist")
)

// MetricService implements business logic for reading and updating metrics on top of a MetricsStorage.
type MetricService struct {
	Storage repository.MetricsStorage
}

// NewMetricService creates a MetricService backed by the given storage.
func NewMetricService(storage repository.MetricsStorage) *MetricService {
	return &MetricService{Storage: storage}
}

// GetAllMetrics returns all stored metrics sorted by id.
func (ms *MetricService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics, err := ms.Storage.GetAllMetrics(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}

// GetByName returns the metric with the given id or ErrMetricNameNotExist if it is missing.
func (ms *MetricService) GetByName(ctx context.Context, id string) (*models.Metrics, error) {
	metric, err := ms.Storage.GetMetricByName(ctx, id)
	if err != nil {
		return nil, err
	}
	if metric == nil {
		return nil, ErrMetricNameNotExist
	}
	return metric, nil
}

// UpsertMetric creates or updates a single metric in a transaction, accumulating counter deltas.
func (ms *MetricService) UpsertMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	txCtx, done, err := ms.Storage.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	err = ms.upsertMetric(txCtx, metric)
	if err = done(err); err != nil {
		return nil, err
	}
	return &metric, nil
}

// UpsertMetrics creates or updates a batch of metrics in one transaction, merging duplicates by id.
func (ms *MetricService) UpsertMetrics(ctx context.Context, metrics []models.Metrics) error {
	merged := make(map[string]models.Metrics)
	for _, m := range metrics {
		key := m.ID
		if existing, ok := merged[key]; ok {
			if existing.MType != m.MType {
				return ErrMetricTypeMismatch
			}
			if m.MType == models.Counter {
				newDelta := *existing.Delta + *m.Delta
				existing.Delta = &newDelta
				merged[key] = existing
				continue
			}
		}
		merged[key] = m
	}

	txCtx, done, err := ms.Storage.BeginTx(ctx)
	if err != nil {
		return err
	}

	for _, metric := range merged {
		if err = ms.upsertMetric(txCtx, metric); err != nil {
			return done(err)
		}
	}
	return done(nil)
}

func (ms *MetricService) upsertMetric(ctx context.Context, metric models.Metrics) error {
	existing, err := ms.Storage.GetMetricByName(ctx, metric.ID)
	if err != nil {
		return err
	}
	if existing != nil && existing.MType != metric.MType {
		return ErrMetricTypeMismatch
	}
	if metric.MType == models.Counter && existing != nil {
		*metric.Delta += *existing.Delta
	}
	return ms.Storage.UpdateMetric(ctx, &metric)
}
