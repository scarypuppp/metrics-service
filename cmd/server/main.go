package main

import (
	"net/http"

	handlers "github.com/scarypuppp/metrics-service/internal/handler"
)

func run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.CreateMetricHandler)

	return http.ListenAndServe(`:8080`, mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
