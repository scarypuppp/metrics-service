package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/scarypuppp/metrics-service/internal/agent"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"go.uber.org/zap"
)

// Моки

// Collector
type mockCollector struct {
	callCount atomic.Int64
}

func (m *mockCollector) CollectMetrics(pollCountValue int64) []models.Metrics {
	m.callCount.Add(1)
	value := 1.0
	return []models.Metrics{
		{ID: "TestGauge", MType: models.Gauge, Value: &value},
	}
}

func (m *mockCollector) CollectCustomMetrics() ([]models.Metrics, error) {
	m.callCount.Add(1)
	value := 1.0
	return []models.Metrics{
		{ID: "TestCustomGauge", MType: models.Gauge, Value: &value},
	}, nil
}

// Sender
type mockSender struct {
	mu   sync.Mutex
	sent []models.Metrics
	err  error
}

func (m *mockSender) SendMetric(metric models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, metric)
	return m.err
}

func (m *mockSender) getSent() []models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]models.Metrics(nil), m.sent...)
}

// Тесты

func TestAgent_CollectsOnPollInterval(t *testing.T) {
	synctest.Run(func() {
		collector := &mockCollector{}
		sender := &mockSender{}
		logger := zap.NewNop()

		a := agent.NewAgent(collector, sender, *logger, 1, 999)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go a.RunCtx(ctx, 10)

		// Даём агенту запуститься и заблокироваться на таймерах
		synctest.Wait()

		// Прокручиваем фиктивное время на 3 секунды — должно сработать 3 тика
		time.Sleep(3 * time.Second)
		synctest.Wait()

		if collector.callCount.Load() == 0 {
			t.Error("expected CollectMetrics to be called, got 0")
		}
	})
}

func TestAgent_SendsOnReportInterval(t *testing.T) {
	synctest.Run(func() {
		collector := &mockCollector{}
		sender := &mockSender{}
		logger := zap.NewNop()

		// pollInterval=1s, reportInterval=2s
		a := agent.NewAgent(collector, sender, *logger, 1, 2)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go a.RunCtx(ctx, 10)

		synctest.Wait()
		time.Sleep(3 * time.Second)
		synctest.Wait()

		if len(sender.getSent()) == 0 {
			t.Error("expected SendMetric to be called, got 0 sends")
		}
	})
}

func TestAgent_SendsCollectedMetrics(t *testing.T) {
	synctest.Run(func() {
		collector := &mockCollector{}
		sender := &mockSender{}

		logger := zap.NewNop()

		a := agent.NewAgent(collector, sender, *logger, 1, 2)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go a.RunCtx(ctx, 10)

		synctest.Wait()
		time.Sleep(3 * time.Second)
		synctest.Wait()

		for _, m := range sender.getSent() {
			if m.ID == "TestGauge" {
				return
			}
		}
		t.Error("expected TestGauge to be sent, not found")
	})
}
