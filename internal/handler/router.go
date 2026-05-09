package handlers

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/metrics-service/internal/middlewares"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func GetAppRouter(
	metricService service.MetricService,
	db *sql.DB,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.LogResponse)
	r.Use(middlewares.CompressResponse)
	r.Use(middlewares.DecompressRequest)
	r.Use(middlewares.LogRequest)

	r.Route("/", func(r chi.Router) {
		r.Get("/", RetrieveMetricsHandler(metricService))
		r.Get("/ping", PingDatabaseHandler(db))
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
