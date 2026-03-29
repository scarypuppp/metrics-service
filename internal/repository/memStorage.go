package repository

import (
	"fmt"

	"github.com/scarypuppp/metrics-service/internal/model"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
}

func (s *MemStorage) GetMetric(id string) *models.Metrics {
	metric, exists := s.Metrics[id]
	if !exists {
		return nil
	}
	return &metric
}

func (s *MemStorage) SetMetric(metric *models.Metrics) error {
	if metric == nil {
		return fmt.Errorf("nil metric recieved")
	}
	s.Metrics[metric.ID] = *metric

	return nil
}
