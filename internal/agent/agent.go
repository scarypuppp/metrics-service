package agent

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
	"go.uber.org/zap"
)

// IMetricCollector interface for metric collecting operations.
type IMetricCollector interface {
	CollectMetrics(pollCountValue int64) []models.Metrics
	CollectCustomMetrics() ([]models.Metrics, error)
}

// IMetricSender interface for metric sending operations
type IMetricSender interface {
	SendMetric(metric models.Metrics) error
}

// Agent represents object collecting and sending metrics.
type Agent struct {
	collector        IMetricCollector
	sender           IMetricSender
	logger           zap.Logger
	mu               sync.RWMutex
	collectedMetrics []models.Metrics
	collecting       atomic.Bool
	pollTicker       *time.Ticker
	reportTicker     *time.Ticker
	pollCount        atomic.Int64
}

// NewAgent constructor method for Agent object.
func NewAgent(collector IMetricCollector, sender IMetricSender, logger zap.Logger, pollInterval int64, reportInterval int64) *Agent {
	return &Agent{
		collector:    collector,
		sender:       sender,
		logger:       logger,
		pollTicker:   time.NewTicker(time.Duration(pollInterval) * time.Second),
		reportTicker: time.NewTicker(time.Duration(reportInterval) * time.Second),
	}
}

// RunCtx method starts agent collecting and reporting metrics.
func (a *Agent) RunCtx(ctx context.Context, rateLimit int) {
	defer a.pollTicker.Stop()
	defer a.reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Stopping agent...")
			return
		case <-a.pollTicker.C:
			a.logger.Debug("Poll tick")
			a.runCollect()
		case <-a.reportTicker.C:
			a.logger.Debug("Report tick")
			a.runReport(ctx, rateLimit)
		}
	}
}

// collectRuntimeWorker represents worker that collects runtime metrics.
func (a *Agent) collectRuntimeWorker(resultsCh chan<- []models.Metrics) {
	defer close(resultsCh)
	defer a.logger.Info("collectRuntimeWorker done")
	a.logger.Info("collectRuntimeWorker started")
	resultsCh <- a.collector.CollectMetrics(a.pollCount.Load())
}

// collectCustomWorker represents worker that collects custom metrics.
func (a *Agent) collectCustomWorker(resultsCh chan<- []models.Metrics) {
	defer close(resultsCh)
	defer a.logger.Info("collectCustomWorker done")
	a.logger.Info("collectCustomWorker started")
	result, err := a.collector.CollectCustomMetrics()
	if err != nil {
		a.logger.Error("Error collecting custom metrics", zap.Error(err))
	}
	resultsCh <- result
}

// sendMetricsWorker represents worker that sends metrics to the server.
func (a *Agent) sendMetricsWorker(ctx context.Context, id int, sendInCh <-chan models.Metrics, sendOutCh chan<- error) {
	defer a.logger.Info("SendMetric worker done", zap.Int("id", id))
	a.logger.Info("SendMetric worker started", zap.Int("id", id))

	for metric := range sendInCh {
		select {
		case <-ctx.Done():
			return
		default:
			sendOutCh <- a.sender.SendMetric(metric)
		}
	}
}

// fanIn implementation of FanIn async pattern.
// Aggregates runtime metrics and custom metrics from various collecting workers.
func fanIn(channels ...<-chan []models.Metrics) <-chan []models.Metrics {
	mergedCh := make(chan []models.Metrics)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan []models.Metrics) {
			defer wg.Done()
			for metrics := range c {
				mergedCh <- metrics
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(mergedCh)
	}()

	return mergedCh
}

// runCollect entrypoint to start collect metrics.
func (a *Agent) runCollect() {
	if !a.collecting.CompareAndSwap(false, true) {
		a.logger.Info("Previous collection still running, skipping")
		return
	}

	runtimeCh := make(chan []models.Metrics, 1)
	customCh := make(chan []models.Metrics, 1)
	mergedCh := fanIn(runtimeCh, customCh)

	go a.collectRuntimeWorker(runtimeCh)
	go a.collectCustomWorker(customCh)

	go func() {
		defer a.collecting.Store(false)

		var collectedMetrics []models.Metrics
		for metrics := range mergedCh {
			collectedMetrics = append(collectedMetrics, metrics...)
		}

		a.mu.Lock()
		a.collectedMetrics = collectedMetrics
		a.mu.Unlock()

		a.pollCount.Add(1)
		a.logger.Info("Collection done", zap.Int("count", len(collectedMetrics)))
	}()
}

// runReport entrypoint to start send metrics to the server.
func (a *Agent) runReport(ctx context.Context, rateLimit int) {
	a.mu.RLock()
	metrics := make([]models.Metrics, len(a.collectedMetrics))
	copy(metrics, a.collectedMetrics)
	a.mu.RUnlock()

	if len(metrics) == 0 {
		a.logger.Info("No metrics to report, skipping")
		return
	}

	sendInCh := make(chan models.Metrics, rateLimit)
	sendOutCh := make(chan error, rateLimit)

	var workerWg sync.WaitGroup
	for i := 0; i < rateLimit; i++ {
		workerWg.Add(1)
		go func(id int) {
			defer workerWg.Done()
			a.sendMetricsWorker(ctx, id, sendInCh, sendOutCh)
		}(i)
	}

	go func() {
		for _, m := range metrics {
			select {
			case <-ctx.Done():
				return
			case sendInCh <- m:
			}
		}
		close(sendInCh)
	}()
	go func() {
		for err := range sendOutCh {
			if err != nil {
				a.logger.Error("ERROR SENDING METRIC", zap.Error(err))
			}
		}
	}()
	workerWg.Wait()
	close(sendOutCh)
}
