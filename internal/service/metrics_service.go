package service

import (
	"context"
	"errors"
	"sort"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
)

var (
	ErrInvalidMetricType  = errors.New("invalid metric type provided")
	ErrMetricTypeMismatch = errors.New("metric with such name already exists with different type")
	ErrMetricNameNotExist = errors.New("metric with such name does not exist")
)

type MetricService struct {
	Storage repository.MetricsStorage
}

func NewMetricService(storage repository.MetricsStorage) *MetricService {
	return &MetricService{Storage: storage}
}

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
