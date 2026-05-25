package agent

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type IMetricCollector interface {
	GetCollectedMetrics() []models.Metrics
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
	pollCount    atomic.Int64
}

func NewAgent(collector IMetricCollector, sender IMetricSender, poolInterval int64, reportInterval int64) *Agent {
	agent := Agent{
		collector:    collector,
		sender:       sender,
		poolTicker:   time.NewTicker(time.Duration(poolInterval) * time.Second),
		reportTicker: time.NewTicker(time.Duration(reportInterval) * time.Second),
	}
	agent.metricsPool = sync.Pool{
		New: func() interface{} {
			return make([]models.Metrics, 0)
		},
	}
	return &agent
}

func (p *Agent) CollectRuntimeWorker(resultsCh chan<- []models.Metrics) {
	log.Printf("CollectRuntimeWorker started")
	resultsCh <- p.collector.CollectMetrics(p.pollCount.Load())
	log.Printf("CollectRuntimeWorker done")
}

func (p *Agent) CollectCustomWorker(resultsCh chan<- []models.Metrics) {
	log.Printf("CollectCustomWorker started")
	resultsCh <- p.collector.CollectCustomMetrics()
	log.Printf("CollectCustomWorker done")
}

func (p *Agent) RunSendMetricWorker(id int, metricsCh <-chan models.Metrics, sendResultsCh chan<- error) {
	log.Printf("SendMetric worker %d started", id)
	currMetric := <-metricsCh
	sendResultsCh <- p.sender.SendMetric(currMetric)
	log.Printf("SendMetric worker %d done", id)
}

func (p *Agent) RunCtx(ctx context.Context, rateLimit int) {
	defer p.poolTicker.Stop()
	defer p.reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping agent...")
			return
		case <-p.poolTicker.C:
			log.Println("POOL!")
			collectMetricsOutCh := make(chan []models.Metrics, 2)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				p.CollectRuntimeWorker(collectMetricsOutCh)
				wg.Done()
			}()
			go func() {
				p.CollectCustomWorker(collectMetricsOutCh)
				wg.Done()
			}()
			go func() {
				var collectedMetrics []models.Metrics
				for metrics := range collectMetricsOutCh {
					collectedMetrics = append(collectedMetrics, metrics...)
				}
				p.metricsPool.Put(collectedMetrics)
			}()
			go func() {
				wg.Wait()
				p.pollCount.Add(1)
				close(collectMetricsOutCh)
			}()
		case <-p.reportTicker.C:
			metrics := p.metricsPool.Get().([]models.Metrics)
			if len(metrics) == 0 {
				break
			}
			sendInCh := make(chan models.Metrics, len(metrics))
			sendOutCh := make(chan error, len(metrics))
			for i := 0; i < rateLimit; i++ {
				go p.RunSendMetricWorker(i, sendInCh, sendOutCh)
			}
			go func() {
				for _, m := range metrics {
					sendInCh <- m
				}
			}()
			go func() {
				for err := range sendOutCh {
					if err != nil {
						log.Printf("ERROR SENDING METRIC: %s", err)
					}
				}
				close(sendInCh)
				close(sendOutCh)
			}()
		}
	}
}
