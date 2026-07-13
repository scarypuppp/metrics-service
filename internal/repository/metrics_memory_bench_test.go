package repository

import (
	"context"
	"fmt"
	"testing"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

func newBenchStorage(b *testing.B, metricsCount int) *MemMetricsStorage {
	b.Helper()
	storage, err := NewMemMetricsStorage()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < metricsCount; i++ {
		value := float64(i)
		err := storage.UpdateMetric(context.Background(), &models.Metrics{
			ID:    fmt.Sprintf("gauge%d", i),
			MType: models.Gauge,
			Value: &value,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	return storage
}

func BenchmarkMemStorageUpdateMetric(b *testing.B) {
	storage := newBenchStorage(b, 0)
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
		if err := storage.UpdateMetric(ctx, &metric); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkMemStorageGetMetricByName(b *testing.B) {
	storage := newBenchStorage(b, 100)
	ctx := context.Background()
	var i int
	b.ResetTimer()
	for b.Loop() {
		if _, err := storage.GetMetricByName(ctx, fmt.Sprintf("gauge%d", i%100)); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkMemStorageGetAllMetrics(b *testing.B) {
	storage := newBenchStorage(b, 100)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		if _, err := storage.GetAllMetrics(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
