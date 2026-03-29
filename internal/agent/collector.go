package agent

import (
	"math/rand/v2"
	"runtime"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type Collector struct{}

func (c *Collector) CollectMetrics(pollCountValue int64) []models.Metrics {
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

	return returnMetrics
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
