package repository

import "github.com/scarypuppp/metrics-service/internal/model"

type IMemStorage interface {
	GetMetric(id string) *models.Metrics
	SetMetric(metric *models.Metrics) error
}
