package service

import (
	"context"
	"fmt"
	"testing"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
)

func newBenchService(b *testing.B) *MetricService {
	b.Helper()
	storage, err := repository.NewMemMetricsStorage()
	if err != nil {
		b.Fatal(err)
	}
	return NewMetricService(storage)
}

func BenchmarkServiceUpsertMetricGauge(b *testing.B) {
	ms := newBenchService(b)
	ctx := context.Background()
	value := 42.5
	var i int
	b.ResetTimer()
	for b.Loop() {
		metric := models.Metrics{
			ID:    fmt.Sprintf("gauge%d", i%100),
			MType: models.Gauge,
			Value: &value,
		}
		if _, err := ms.UpsertMetric(ctx, metric); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkServiceUpsertMetricCounter(b *testing.B) {
	ms := newBenchService(b)
	ctx := context.Background()
	var i int
	b.ResetTimer()
	for b.Loop() {
		delta := int64(1)
		metric := models.Metrics{
			ID:    fmt.Sprintf("counter%d", i%100),
			MType: models.Counter,
			Delta: &delta,
		}
		if _, err := ms.UpsertMetric(ctx, metric); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkServiceUpsertMetricsBatch(b *testing.B) {
	ms := newBenchService(b)
	ctx := context.Background()

	const batchSize = 50
	batch := make([]models.Metrics, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		value := float64(i)
		batch = append(batch, models.Metrics{
			ID:    fmt.Sprintf("gauge%d", i),
			MType: models.Gauge,
			Value: &value,
		})
	}

	b.ResetTimer()
	for b.Loop() {
		if err := ms.UpsertMetrics(ctx, batch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceGetAllMetrics(b *testing.B) {
	ms := newBenchService(b)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		value := float64(i)
		_, err := ms.UpsertMetric(ctx, models.Metrics{
			ID:    fmt.Sprintf("gauge%d", i),
			MType: models.Gauge,
			Value: &value,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := ms.GetAllMetrics(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
