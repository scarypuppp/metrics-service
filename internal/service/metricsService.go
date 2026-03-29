package service

import (
	"errors"
	"strconv"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
)

var ErrInvalidMetricType = errors.New("invalid metric type provided")
var ErrMetricTypeMismatch = errors.New("metric with such name already exists with other type")

type MetricService struct {
	Storage repository.IMemStorage
}

type IMetricService interface {
	CreateMetric(name string, MType string, stringValue string) (*models.Metrics, error)
}

func (ms *MetricService) CreateMetric(name string, MType string, stringValue string) (*models.Metrics, error) {

	metric := (*models.Metrics)(nil)

	existingMetric := ms.Storage.GetMetric(name)

	if existingMetric != nil && existingMetric.MType != MType {
		return nil, ErrMetricTypeMismatch
	}

	if MType == models.Gauge {
		value, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return nil, err
		}
		if existingMetric != nil {
			existingMetric.Value = &value
			return existingMetric, nil
		}

		metric = &models.Metrics{
			ID:    name,
			MType: MType,
			Delta: nil,
			Value: &value,
			Hash:  "",
		}
	} else if MType == models.Counter {
		value, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return nil, err
		}
		if existingMetric != nil {
			*existingMetric.Delta += value
			return existingMetric, nil
		}

		metric = &models.Metrics{
			ID:    name,
			MType: MType,
			Delta: &value,
			Value: nil,
			Hash:  "",
		}
	} else {
		return nil, ErrInvalidMetricType
	}

	err := ms.Storage.SetMetric(metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}
