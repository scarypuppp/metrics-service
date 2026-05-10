package repository

import (
	"context"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type IMetricsStorage interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
	GetMetricByName(ctx context.Context, id string) (*models.Metrics, error)
	UpdateMetric(ctx context.Context, metric *models.Metrics) error
}
