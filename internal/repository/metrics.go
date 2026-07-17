package repository

import (
	"context"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

// MetricsStorage abstracts metrics persistence with read, upsert and transaction support.
type MetricsStorage interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
	GetMetricByName(ctx context.Context, id string) (*models.Metrics, error)
	UpdateMetric(ctx context.Context, metric *models.Metrics) error
	BeginTx(ctx context.Context) (context.Context, func(error) error, error)
}
