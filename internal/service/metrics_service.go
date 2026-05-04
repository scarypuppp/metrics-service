package service

import (
	"errors"
	"sort"
	"strconv"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

var ErrInvalidMetricType = errors.New("invalid metric type provided")
var ErrMetricTypeMismatch = errors.New("metric with such name already exists with other type")
var ErrMetricNameNotExist = errors.New("metric with such does not exist")

type IMemStorage interface {
	All() []models.Metrics
	GetByName(id string) *models.Metrics
	Set(metric *models.Metrics) error
}

type MetricService struct {
	Storage IMemStorage
}

func (ms *MetricService) GetAllMetrics() []models.Metrics {
	metrics := ms.Storage.All()
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (ms *MetricService) GetByName(id string) (*models.Metrics, error) {
	metric := ms.Storage.GetByName(id)
	if metric == nil {
		return nil, ErrMetricNameNotExist
	}
	return metric, nil
}

func (ms *MetricService) UpsertMetric(name string, MType string, stringValue string) (*models.Metrics, error) {
	existingMetric := ms.Storage.GetByName(name)

	if existingMetric != nil && existingMetric.MType != MType {
		return nil, ErrMetricTypeMismatch
	}

	switch MType {
	case models.Gauge:
		value, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return nil, err
		}
		if existingMetric != nil {
			existingMetric.Value = &value
		} else {
			existingMetric = &models.Metrics{
				ID:    name,
				MType: MType,
				Value: &value,
			}
		}
		err = ms.Storage.Set(existingMetric)
		if err != nil {
			return nil, err
		}
		return existingMetric, nil
	case models.Counter:
		value, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return nil, err
		}
		if existingMetric != nil {
			*existingMetric.Delta += value
		} else {
			existingMetric = &models.Metrics{
				ID:    name,
				MType: MType,
				Delta: &value,
			}
		}
		err = ms.Storage.Set(existingMetric)
		if err != nil {
			return nil, err
		}
		return existingMetric, nil
	default:
		return nil, ErrInvalidMetricType
	}
}
