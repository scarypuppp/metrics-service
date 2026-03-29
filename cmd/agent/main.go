package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/scarypuppp/metrics-service/internal/model"
)

var poolInterval int64 = 2
var reportInterval int64 = 10

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

func collectMetrics(pollCountValue *int64) []models.Metrics {
	runtimeMetrics := getMemStats()
	var returnMetrics []models.Metrics
	for metricName, metricValue := range runtimeMetrics {
		metric := models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Delta: nil,
			Value: &metricValue,
			Hash:  "",
		}
		returnMetrics = append(returnMetrics, metric)
	}

	pollCountMetric := models.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: pollCountValue,
		Value: nil,
		Hash:  "",
	}
	returnMetrics = append(returnMetrics, pollCountMetric)

	randomValue := rand.Float64()
	randomValueMetric := models.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Delta: nil,
		Value: &randomValue,
		Hash:  "",
	}
	returnMetrics = append(returnMetrics, randomValueMetric)

	return returnMetrics
}

func sendMetric(client *http.Client, metric models.Metrics) error {
	var stringValue = ""
	if metric.MType == models.Gauge {
		stringValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}
	if metric.MType == models.Counter {
		stringValue = strconv.FormatInt(*metric.Delta, 10)
	}
	if stringValue == "" {
		panic("YOO")
	}
	url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%s", metric.MType, metric.ID, stringValue)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		panic(fmt.Errorf("failed to make request: %w", err))
	}
	response, err := client.Do(request)
	if err != nil {
		panic(fmt.Errorf("failed to send metric: %w", err))
	}
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read response: %w", err))
	}
	responseText := string(bodyBytes)
	if response.StatusCode != 200 {
		fmt.Println("Error sending metric:", response.StatusCode, "-", responseText)
	}
	err = response.Body.Close()
	if err != nil {
		panic(fmt.Errorf("failed to close response body: %w", err))
	}
	return nil
}

func main() {
	client := &http.Client{}
	fmt.Println("Start pooling...")
	var poolCountValue int64 = 0
	var iterationCounter int64 = 0
	var metrics []models.Metrics
	for {
		if iterationCounter%poolInterval == 0 {
			metrics = collectMetrics(&poolCountValue)
		}
		if iterationCounter%reportInterval == 0 {
			for _, metric := range metrics {
				sendMetric(client, metric)
			}
		}
		time.Sleep(time.Second)
		iterationCounter += 1
	}
}
