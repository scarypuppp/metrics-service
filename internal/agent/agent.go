package agent

import (
	"fmt"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type IMetricCollector interface {
	CollectMetrics(pollCountValue int64) []models.Metrics
}

type IMetricSender interface {
	SendMetric(metric models.Metrics) error
}

type Agent struct {
	collector      IMetricCollector
	sender         IMetricSender
	poolInterval   int64
	reportInterval int64
}

func NewAgent(collector IMetricCollector, sender IMetricSender, poolInterval int64, reportInterval int64) *Agent {
	return &Agent{
		collector:      collector,
		sender:         sender,
		poolInterval:   poolInterval,
		reportInterval: reportInterval,
	}
}

func (p *Agent) Run() {
	fmt.Println("Start pooling...")
	var poolCountValue int64 = 0
	var iterationCounter int64 = 1
	var metrics []models.Metrics
	for {
		if iterationCounter%p.poolInterval == 0 {
			metrics = p.handleCollectMetrics(&poolCountValue)
			poolCountValue++
		}
		if iterationCounter%p.reportInterval == 0 {
			p.handlerSendMetrics(metrics)
			poolCountValue = 0
		}
		time.Sleep(time.Second)
		iterationCounter++
	}
}

func (p *Agent) handleCollectMetrics(poolCountValue *int64) []models.Metrics {
	result := p.collector.CollectMetrics(*poolCountValue)
	return result
}

func (p *Agent) handlerSendMetrics(metrics []models.Metrics) {
	for _, metric := range metrics {
		err := p.sender.SendMetric(metric)
		if err != nil {
			fmt.Printf("Failed to send metric %s: %s\n", metric.ID, err)
		}
	}
}
