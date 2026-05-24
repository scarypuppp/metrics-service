package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scarypuppp/metrics-service/internal/agent"
	models "github.com/scarypuppp/metrics-service/internal/model"
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
	return append([]models.Metrics(nil), m.sent...) // возвращаем копию
}

// Тесты

func TestAgent_CollectsOnPollInterval(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := agent.NewAgent(collector, sender, 1, 999)
	go a.Run()
	time.Sleep(3 * time.Second)

	if collector.callCount.Load() == 0 {
		t.Error("expected CollectMetrics to be called, got 0")
	}
}

func TestAgent_SendsOnReportInterval(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := agent.NewAgent(collector, sender, 1, 2)

	go a.Run()
	time.Sleep(3 * time.Second)

	if len(sender.getSent()) == 0 {
		t.Error("expected SendMetric to be called, got 0 sends")
	}
}

func TestAgent_SendsCollectedMetrics(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := agent.NewAgent(collector, sender, 1, 2)

	go a.Run()
	time.Sleep(3 * time.Second)

	for _, m := range sender.getSent() {
		if m.ID == "TestGauge" {
			return
		}
	}
	t.Error("expected TestGauge to be sent, not found")
}
