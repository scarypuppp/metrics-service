package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/metrics-service/internal/middlewares"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func GetAppRouter() chi.Router {
	r := chi.NewRouter()

	storage := repository.NewMemStorage()
	metricService := service.MetricService{Storage: storage}

	r.Use(middlewares.LogResponse)
	r.Use(middlewares.LogRequest)

	r.Route("/", func(r chi.Router) {
		r.Get("/", RetrieveMetricsHandler(metricService))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", GetMetricHandler(metricService))
			r.Get("/{metricType}/{metricName}", GetMetricByURLHandler(metricService))
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", UpdateMetricHandler(metricService))
			r.Post("/{metricType}/{metricName}/{metricValue}", UpdateMetricByURLHandler(metricService))
		})
	})

	return r
}
