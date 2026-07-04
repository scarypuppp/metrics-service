package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/middlewares"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func GetAppRouter(
	key string,
	metricService service.MetricService,
	publisher *audit.Publisher,
	db *sqlx.DB,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.LogResponse)
	r.Use(middlewares.CompressResponse)
	r.Use(middlewares.DecompressRequest)
	if key != "" {
		r.Use(middlewares.ValidateRequestHash(key))
	}
	r.Use(middlewares.LogRequest)

	r.Route("/", func(r chi.Router) {
		r.Get("/", RetrieveMetricsHandler(metricService))
		r.Get("/ping", PingDatabaseHandler(db))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", GetMetricHandler(metricService))
			r.Get("/{metricType}/{metricName}", GetMetricByURLHandler(metricService))
		})
		r.Group(func(r chi.Router) {
			r.Use(middlewares.GetRequestIP)
			r.Route("/update", func(r chi.Router) {
				r.Post("/", UpdateMetricHandler(metricService, publisher))
				r.Post("/{metricType}/{metricName}/{metricValue}", UpdateMetricByURLHandler(metricService, publisher))
			})
			r.Post("/updates/", UpdateMetricsHandler(metricService, publisher))
		})
	})
	return r
}
