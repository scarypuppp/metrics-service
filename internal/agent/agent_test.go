package agent

import (
	"testing"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

// Моки

type mockCollector struct {
	callCount int
}

func (m *mockCollector) CollectMetrics(pollCountValue int64) []models.Metrics {
	m.callCount++
	value := 1.0
	return []models.Metrics{
		{ID: "TestGauge", MType: models.Gauge, Value: &value},
	}
}

type mockSender struct {
	sent []models.Metrics
	err  error
}

func (m *mockSender) SendMetric(metric models.Metrics) error {
	m.sent = append(m.sent, metric)
	return m.err
}

// Тесты

func TestAgent_CollectsOnPollInterval(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := NewAgent(collector, sender, 1, 999)
	go a.Run()
	time.Sleep(3 * time.Second)

	if collector.callCount == 0 {
		t.Error("expected CollectMetrics to be called, got 0")
	}
}

func TestAgent_SendsOnReportInterval(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := NewAgent(collector, sender, 1, 2)

	go a.Run()
	time.Sleep(3 * time.Second)

	if len(sender.sent) == 0 {
		t.Error("expected SendMetric to be called, got 0 sends")
	}
}

func TestAgent_SendsCollectedMetrics(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSender{}

	a := NewAgent(collector, sender, 1, 2)

	go a.Run()
	time.Sleep(3 * time.Second)

	for _, m := range sender.sent {
		if m.ID == "TestGauge" {
			return
		}
	}
	t.Error("expected TestGauge to be sent, not found")
}
