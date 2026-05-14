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
	Storage repository.IMetricsStorage
}

func NewMetricService(storage repository.IMetricsStorage) *MetricService {
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
	existing, err := ms.Storage.GetMetricByName(ctx, metric.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.MType != metric.MType {
		return nil, ErrMetricTypeMismatch
	}
	if metric.MType == models.Counter && existing != nil {
		*metric.Delta += *existing.Delta
	}
	if err = ms.Storage.UpdateMetric(ctx, &metric); err != nil {
		return nil, err
	}
	return &metric, nil
}
