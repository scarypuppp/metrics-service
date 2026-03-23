package main

import (
	"net/http"

	handlers "github.com/scarypuppp/metrics-service/internal/handler"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func run() error {
	mux := http.NewServeMux()

	storage := repository.MemStorage{Metrics: make(map[string]models.Metrics)}
	metricService := service.MetricService{Storage: &storage}
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.CreateMetricHandler(metricService))

	return http.ListenAndServe(`:8080`, mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
