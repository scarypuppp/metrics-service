package handler

import (
	"fmt"
	"net/http"

	"github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func CreateMetricHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("content-Type") != "text/plain" {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}
	metricValue := req.PathValue("metricValue")

	storage := repository.MemStorage{Metrics: make(map[string]models.Metrics)}
	metricService := service.MetricService{Storage: &storage}
	metric, err := metricService.CreateMetric(metricName, metricType, metricValue)

	if err != nil {
		errorMessage := fmt.Sprintf("Error creating metric: %s", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	fmt.Println(*metric.Delta)

	w.WriteHeader(http.StatusOK)
}
