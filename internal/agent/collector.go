package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type Collector struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]models.Metrics),
	}
}

func (c *Collector) CollectMetrics(pollCountValue int64) []models.Metrics {
	c.mu.RLock()
	runtimeMetrics := getMemStats()
	var returnMetrics []models.Metrics
	for metricName, metricValue := range runtimeMetrics {
		metric := models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Delta: nil,
			Value: &metricValue,
		}
		returnMetrics = append(returnMetrics, metric)
	}

	pollCountMetric := models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &pollCountValue,
		Value: nil,
	}
	returnMetrics = append(returnMetrics, pollCountMetric)

	randomValue := rand.Float64()
	randomValueMetric := models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Delta: nil,
		Value: &randomValue,
	}
	returnMetrics = append(returnMetrics, randomValueMetric)
	c.mu.RUnlock()
	c.updateMetrics(returnMetrics)
	return returnMetrics
}

func (c *Collector) CollectCustomMetrics() []models.Metrics {
	c.mu.RLock()
	var returnMetrics []models.Metrics
	customStats := getCustomStats()
	for metricName, metricValue := range customStats {
		metric := models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Delta: nil,
			Value: &metricValue,
		}
		returnMetrics = append(returnMetrics, metric)
	}
	c.mu.RUnlock()
	c.updateMetrics(returnMetrics)
	return returnMetrics
}

func (c *Collector) updateMetrics(metrics []models.Metrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range metrics {
		c.metrics[m.ID] = m
	}
}

func getCustomStats() map[string]float64 {
	vm, _ := mem.VirtualMemory()

	result := map[string]float64{
		"TotalMemory": float64(vm.Total),
		"FreeMemory":  float64(vm.Free),
	}

	perCore, _ := cpu.Percent(time.Second, true)
	for i, usage := range perCore {
		key := fmt.Sprintf("CPUutilization%d", i)
		result[key] = usage
	}

	return result
}

func getMemStats() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
	}
}
