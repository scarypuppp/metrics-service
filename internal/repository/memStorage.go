package repository

import (
	"fmt"

	"github.com/scarypuppp/metrics-service/internal/model"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Metrics: make(map[string]models.Metrics)}
}

func (s *MemStorage) All() *[]models.Metrics {
	metrics := make([]models.Metrics, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}
	return &metrics
}

func (s *MemStorage) GetByName(id string) *models.Metrics {
	metric, exists := s.Metrics[id]
	if !exists {
		return nil
	}
	return &metric
}

func (s *MemStorage) Set(metric *models.Metrics) error {
	if metric == nil {
		return fmt.Errorf("nil metric recieved")
	}
	s.Metrics[metric.ID] = *metric

	return nil
}
