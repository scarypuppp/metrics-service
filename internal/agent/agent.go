package agent

import (
	"fmt"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

const poolInterval int64 = 2
const reportInterval int64 = 10

type Agent struct {
	Collector *Collector
	Sender    *Sender
}

func (p *Agent) Run() {
	fmt.Println("Start pooling...")
	var poolCountValue int64 = 0
	var iterationCounter int64 = 1
	var metrics []models.Metrics
	for {
		if iterationCounter%poolInterval == 0 {
			metrics = p.handleCollectMetrics(&poolCountValue)
			poolCountValue++
		}
		if iterationCounter%reportInterval == 0 {
			p.handlerSendMetrics(metrics)
		}
		time.Sleep(time.Second)
		iterationCounter++
	}
}

func (p *Agent) handleCollectMetrics(poolCountValue *int64) []models.Metrics {
	result := p.Collector.CollectMetrics(*poolCountValue)
	return result
}

func (p *Agent) handlerSendMetrics(metrics []models.Metrics) {
	for _, metric := range metrics {
		err := p.Sender.SendMetric(metric)
		if err != nil {
			fmt.Printf("Failed to send metric %s: %s\n", metric.ID, err)
		}
	}
}
