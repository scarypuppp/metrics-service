package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"go.uber.org/zap"
)

// MemMetricsStorage is an in-memory MetricsStorage with optional file persistence.
type MemMetricsStorage struct {
	Metrics map[string]models.Metrics

	fileName      string
	storeInterval time.Duration

	mu sync.RWMutex
}

// Option configures a MemMetricsStorage during construction.
type Option func(*MemMetricsStorage) error

// WithFile enables file persistence: optional restore on start and periodic or synchronous saving.
func WithFile(ctx context.Context, fileName string, storeIntervalSec int, restore bool) Option {
	return func(s *MemMetricsStorage) error {
		s.fileName = fileName
		s.storeInterval = time.Duration(storeIntervalSec) * time.Second
		if restore {
			err := s.restoreFromFile()
			if err != nil {
				return err
			}
		}
		if s.storeInterval > 0 {
			s.startPeriodicSave(ctx)
		}
		return nil
	}
}

// NewMemMetricsStorage creates an in-memory storage and applies the given options.
func NewMemMetricsStorage(opts ...Option) (*MemMetricsStorage, error) {
	s := &MemMetricsStorage{Metrics: make(map[string]models.Metrics)}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}
	return s, nil
}

func (s *MemMetricsStorage) startPeriodicSave(ctx context.Context) {
	ticker := time.NewTicker(s.storeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.saveToFileWithLog()
			case <-ctx.Done():
				s.saveToFileWithLog()
				return
			}
		}
	}()
}

// BeginTx satisfies MetricsStorage; in-memory storage has no transactions, so it is a no-op.
func (s *MemMetricsStorage) BeginTx(ctx context.Context) (context.Context, func(error) error, error) {
	doneFn := func(err error) error {
		return err
	}
	return ctx, doneFn, nil
}

// GetAllMetrics returns all stored metrics.
func (s *MemMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	metrics := make([]models.Metrics, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// GetMetricByName returns the metric with the given id, or nil if it does not exist.
func (s *MemMetricsStorage) GetMetricByName(ctx context.Context, id string) (*models.Metrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if metric, exists := s.Metrics[id]; exists {
		return &metric, nil
	}
	return nil, nil
}

// UpdateMetric stores the metric, saving to file synchronously when configured with a zero store interval.
func (s *MemMetricsStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {
	if metric == nil {
		return fmt.Errorf("nil metric received")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Metrics[metric.ID] = *metric

	if s.fileName != "" && s.storeInterval == 0 {
		return s.saveToFileNoLock()
	}
	return nil
}

func (s *MemMetricsStorage) saveToFile() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveToFileNoLock()
}

func (s *MemMetricsStorage) saveToFileNoLock() error {
	file, err := os.OpenFile(s.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(s.Metrics)
}

func (s *MemMetricsStorage) saveToFileWithLog() {
	if err := s.saveToFile(); err != nil {
		zap.S().Error("failed to save metrics to file",
			zap.Error(err),
			zap.String("filename", s.fileName),
		)
	} else {
		zap.S().Info("metrics successfully saved to file", zap.String("filename", s.fileName))
	}
}

func (s *MemMetricsStorage) restoreFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.fileName, os.O_RDONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&s.Metrics)
}
