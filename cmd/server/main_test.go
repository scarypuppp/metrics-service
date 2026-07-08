package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/handler"
	"github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouter(metricService service.MetricService, publisher *audit.Publisher) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricByURLHandler(metricService, publisher))
	return r
}

func TestCreateMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid gauge metric",
			method:         http.MethodPost,
			url:            "/update/gauge/temperature/36.6",
			expectedStatus: http.StatusOK,
			expectedBody:   "36.6",
		},
		{
			name:           "valid counter metric",
			method:         http.MethodPost,
			url:            "/update/counter/hits/10",
			expectedStatus: http.StatusOK,
			expectedBody:   "10",
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			url:            "/update/gauge/temperature/36.6",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			url:            "/update/unknown/temperature/36.6",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid metric value",
			method:         http.MethodPost,
			url:            "/update/gauge/temperature/abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.MemMetricsStorage{Metrics: make(map[string]models.Metrics)}
			metricService := service.MetricService{Storage: &storage}
			publisher := audit.NewPublisher()

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			newRouter(metricService, publisher).ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestCreateMetricHandler_GaugeOverwritesOnUpdate(t *testing.T) {
	storage := repository.MemMetricsStorage{Metrics: make(map[string]models.Metrics)}
	metricService := service.MetricService{Storage: &storage}
	publisher := audit.NewPublisher()
	r := newRouter(metricService, publisher)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/10.5", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/99.9", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), "99.9")
}

func TestCreateMetricHandler_CounterAccumulatesOnUpdate(t *testing.T) {
	storage := repository.MemMetricsStorage{Metrics: make(map[string]models.Metrics)}
	metricService := service.MetricService{Storage: &storage}
	publisher := audit.NewPublisher()
	r := newRouter(metricService, publisher)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/hits/10", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest(http.MethodPost, "/update/counter/hits/5", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), "15")
}

func TestCreateMetricHandler_MetricTypeConflictReturnsError(t *testing.T) {
	storage := repository.MemMetricsStorage{Metrics: make(map[string]models.Metrics)}
	metricService := service.MetricService{Storage: &storage}
	publisher := audit.NewPublisher()
	r := newRouter(metricService, publisher)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/36.6", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest(http.MethodPost, "/update/counter/temperature/10", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
