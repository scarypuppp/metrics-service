package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/middlewares"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/service"
	"go.uber.org/zap"
)

func sendEvent(p *audit.Publisher, rt time.Time, ms []models.Metrics, addr string) {
	var metricNames []string
	for _, m := range ms {
		metricNames = append(metricNames, m.ID)
	}
	event := audit.Event{Timestamp: rt.Unix(), Metrics: metricNames, Address: addr}
	p.Publish(event)
}

// RetrieveMetricsHandler returns a GET handler rendering all metrics as an HTML page.
func RetrieveMetricsHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Обработка запроса
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		//Получение метрик
		metrics, err := metricService.GetAllMetrics(req.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// Формирование ответа
		var sb strings.Builder
		sb.Grow(128 + 80*len(metrics))
		sb.WriteString(`<html><head><meta name="color-scheme" content="light dark"></head><body>` +
			`<pre style="word-wrap: break-word; white-space: pre-wrap;">`)
		for _, m := range metrics {
			fmt.Fprintf(&sb, "# HELP %s\n# TYPE %s %s\n%s %s\n", m.ID, m.ID, m.MType, m.ID, m.StringValue())
		}
		sb.WriteString(`</pre></body></html>`)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, sb.String())
	}
}

// GetMetricByURLHandler returns a GET handler serving a single metric value from URL path parameters.
func GetMetricByURLHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//Обработка запроса
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusBadRequest)
			return
		}

		// Получение метрики
		metric, getMetricErr := metricService.GetByName(req.Context(), metricName)
		if getMetricErr != nil {
			switch {
			case errors.Is(getMetricErr, service.ErrMetricNameNotExist):
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Формирование ответа
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metric.StringValue()))
	}
}

// GetMetricHandler returns a POST handler serving a single metric as JSON for a JSON request body.
func GetMetricHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var metric models.Metrics
		var buffer bytes.Buffer
		_, err := buffer.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buffer.Bytes(), &metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Получение метрики
		existingMetric, getMetricErr := metricService.GetByName(req.Context(), metric.ID)
		if getMetricErr != nil {
			switch {
			case errors.Is(getMetricErr, service.ErrMetricNameNotExist):
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		responseData, err := json.Marshal(existingMetric)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}
}

// UpdateMetricByURLHandler returns a POST handler upserting a metric from URL path parameters and publishing an audit event.
func UpdateMetricByURLHandler(metricService service.MetricService, publisher *audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestIP, ok := req.Context().Value(middlewares.RequestIPKey).(string)
		if !ok {
			requestIP = "unknown"
		}
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusBadRequest)
			return
		}
		metricValue := chi.URLParam(req, "metricValue")
		// Обновление/создание метрики
		metric, err := models.NewMetric(metricName, metricType, metricValue)
		if err != nil {
			http.Error(w, "error building metric entity", http.StatusBadRequest)
			return
		}
		_, createMeticErr := metricService.UpsertMetric(req.Context(), *metric)

		if createMeticErr != nil {
			zap.S().Error("Error upsert metric", zap.Error(err))
			switch {
			case errors.Is(createMeticErr, strconv.ErrSyntax):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			case errors.Is(createMeticErr, service.ErrInvalidMetricType):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			case errors.Is(createMeticErr, service.ErrMetricTypeMismatch):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			default:
				errorMessage := fmt.Sprintf("Error creating metric: %s", createMeticErr)
				http.Error(w, errorMessage, http.StatusInternalServerError)
			}
			return
		}

		sendEvent(publisher, time.Now(), []models.Metrics{*metric}, requestIP)

		// Формирование ответа
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metric.StringValue()))
	}
}

// UpdateMetricHandler returns a POST handler upserting a metric from a JSON body and publishing an audit event.
func UpdateMetricHandler(metricService service.MetricService, publisher *audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestIP, ok := req.Context().Value(middlewares.RequestIPKey).(string)
		if !ok {
			requestIP = "unknown"
		}
		var metric models.Metrics
		var buffer bytes.Buffer
		_, err := buffer.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buffer.Bytes(), &metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Обновление/создание метрики
		_, createMeticErr := metricService.UpsertMetric(req.Context(), metric)

		if createMeticErr != nil {
			zap.S().Error("Error upsert metric", zap.Error(err))
			switch {
			case errors.Is(createMeticErr, strconv.ErrSyntax):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			case errors.Is(createMeticErr, service.ErrInvalidMetricType):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			case errors.Is(createMeticErr, service.ErrMetricTypeMismatch):
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			default:
				errorMessage := fmt.Sprintf("Error creating metric: %s", createMeticErr)
				http.Error(w, errorMessage, http.StatusInternalServerError)
			}
			return
		}

		sendEvent(publisher, time.Now(), []models.Metrics{metric}, requestIP)

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateMetricsHandler returns a POST handler upserting a batch of metrics from a JSON array body.
func UpdateMetricsHandler(metricService service.MetricService, publisher *audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestIP, ok := req.Context().Value(middlewares.RequestIPKey).(string)
		if !ok {
			requestIP = "unknown"
		}
		var metrics []models.Metrics
		var buffer bytes.Buffer

		_, err := buffer.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buffer.Bytes(), &metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = metricService.UpsertMetrics(req.Context(), metrics); err != nil {
			zap.S().Error("Error upsert metrics", zap.Error(err))
			switch {
			case errors.Is(err, service.ErrInvalidMetricType):
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, service.ErrMetricTypeMismatch):
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, fmt.Sprintf("Error updating metrics: %s", err), http.StatusInternalServerError)
			}
			return
		}

		sendEvent(publisher, time.Now(), metrics, requestIP)

		w.WriteHeader(http.StatusOK)
	}
}
