package agent

import (
	"context"
	"log"
	"sync"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type IMetricCollector interface {
	CollectMetrics(pollCountValue int64) []models.Metrics
	CollectCustomMetrics() []models.Metrics
}

type IMetricSender interface {
	SendMetric(metric models.Metrics) error
}

type Agent struct {
	collector    IMetricCollector
	sender       IMetricSender
	metricsPool  sync.Pool
	poolTicker   *time.Ticker
	reportTicker *time.Ticker
	pollCount    int64
}

func NewAgent(collector IMetricCollector, sender IMetricSender, poolInterval int64, reportInterval int64) *Agent {
	return &Agent{
		collector:    collector,
		sender:       sender,
		poolTicker:   time.NewTicker(time.Duration(poolInterval) * time.Second),
		reportTicker: time.NewTicker(time.Duration(reportInterval) * time.Second),
	}
}

func (p *Agent) RunCollectMetricsWorker(collectType string, resultsCh chan<- models.Metrics) {
	var metrics []models.Metrics
	switch collectType {
	case "1":
		log.Println("Collecting Runtime")
		metrics = p.collector.CollectMetrics(p.pollCount)
	case "2":
		log.Println("Collecting Custom")
		metrics = p.collector.CollectCustomMetrics()
	}
	for _, m := range metrics {
		resultsCh <- m
	}
}

func (p *Agent) RunSendMetricWorker(id int, metricsCh <-chan models.Metrics, sendResultsCh chan<- error) {
	log.Printf("SendMetric worker %d started", id)
	currMetric := <-metricsCh
	log.Printf("SendMetric worker %d metric %s", id, currMetric.ID)
	time.Sleep(10 * time.Millisecond)
	sendResultsCh <- nil
	log.Printf("SendMetric worker %d done", id)
}

func (p *Agent) RunCtx(ctx context.Context, rateLimit int) {
	defer p.poolTicker.Stop()
	defer p.reportTicker.Stop()

	collectMetricsOutCh := make(chan models.Metrics)
	sendInCh := make(chan models.Metrics)
	sendOutCh := make(chan error)

	defer close(collectMetricsOutCh)
	defer close(sendInCh)
	defer close(sendOutCh)

	var collectedMetrics []models.Metrics

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping agent...")
			return
		case <-p.poolTicker.C:
			log.Println("POOL!")
			go func() {
				for m := range collectMetricsOutCh {
					collectedMetrics = append(collectedMetrics, m)
				}
			}()
			go p.RunCollectMetricsWorker("1", collectMetricsOutCh)
			go p.RunCollectMetricsWorker("2", collectMetricsOutCh)
		case <-p.reportTicker.C:
			log.Println("REPORT!")
			for i := 0; i < rateLimit; i++ {
				go p.RunSendMetricWorker(i, sendInCh, sendOutCh)
			}
		}
	}
}
