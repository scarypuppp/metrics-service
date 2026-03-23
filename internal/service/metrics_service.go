package service

import (
	"fmt"
	"strconv"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
)

type MetricService struct {
	Storage repository.IMemStorage
}

type IMetricService interface {
	CreateMetric(name string, MType string, stringValue string) (*models.Metrics, error)
}

func (ms *MetricService) CreateMetric(name string, MType string, stringValue string) (*models.Metrics, error) {

	metric := (*models.Metrics)(nil)

	existingMetric := ms.Storage.GetMetric(name)

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
		return nil, fmt.Errorf("wrong metric type passed")
	}

	err := ms.Storage.SetMetric(metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}
