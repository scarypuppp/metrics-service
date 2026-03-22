package main

import (
	"fmt"
	"net/http"
)

type Metric struct {
	metricType string
	metricName string
}

type GaugeMetric struct {
	Metric
	value float64
}

type CounterMetric struct {
	Metric
	value int64
}

type MemStorageRepo struct {
	metrics map[string][]Metric
}

func getMetricsByType() {

}

func handleMetric(w http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	fmt.Println(metricType, metricName, metricValue)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handleMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
