package models

import (
	"errors"
	"fmt"
	"strconv"
)

// Supported metric types.
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// ErrWrongMetricType is returned when a metric type is neither counter nor gauge.
var ErrWrongMetricType = errors.New("unknown metric type")

// ErrInvalidName is returned when a metric name is empty.
var ErrInvalidName = errors.New("invalid name provided")

// NewMetric builds a Metrics entity from a name, type and string value, validating and parsing the value.
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

// Metrics represents a single metric: Delta is used for counters, Value for gauges.
type Metrics struct {
	ID    string   `json:"id" db:"id"`
	MType string   `json:"type" db:"mtype"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty" db:"hash"`
}

// StringValue returns the metric value formatted as a string according to its type.
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
