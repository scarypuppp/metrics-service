package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/scarypuppp/metrics-service/internal/model"
)

type MemStorage struct {
	Metrics   map[string]models.Metrics
	FileName  string
	saveOnSet bool
	mu        sync.RWMutex
}

func NewMemStorage(fileName string, saveOnSet bool) *MemStorage {
	return &MemStorage{
		Metrics:   make(map[string]models.Metrics),
		FileName:  fileName,
		saveOnSet: saveOnSet,
	}
}

func (s *MemStorage) All() []models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

func (s *MemStorage) GetByName(id string) *models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metrics[metric.ID] = *metric
	if s.saveOnSet {
		return s.saveToFile()
	}
	return nil
}

func (s *MemStorage) SaveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveToFile()
}

func (s *MemStorage) RestoreFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.FileName, os.O_RDONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	metrics := make(map[string]models.Metrics)
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}

	s.Metrics = metrics
	return nil
}

func (s *MemStorage) saveToFile() error {
	file, err := os.OpenFile(s.FileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(s.Metrics)
}
