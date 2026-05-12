package models

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var ErrWrongMetricType = errors.New("unknown metric type")
var ErrInvalidName = errors.New("invalid name provided")

func NewMetric(name, mType, stringValue string) (*Metrics, error) {
	if name == "" {
		return nil, ErrInvalidName
	}

	switch mType {
	case Gauge:
		value, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid gauge value %q: %w", stringValue, err)
		}
		return &Metrics{
			ID:    name,
			MType: Gauge,
			Value: &value,
		}, nil

	case Counter:
		delta, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid counter value %q: %w", stringValue, err)
		}
		return &Metrics{
			ID:    name,
			MType: Counter,
			Delta: &delta,
		}, nil

	default:
		return nil, ErrWrongMetricType
	}
}

type Metrics struct {
	ID    string   `json:"id" db:"id"`
	MType string   `json:"type" db:"mtype"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty" db:"hash"`
}

func (m *Metrics) StringValue() string {
	switch m.MType {
	case Gauge:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case Counter:
		return strconv.FormatInt(*m.Delta, 10)
	default:
		panic(ErrWrongMetricType)
	}
}
