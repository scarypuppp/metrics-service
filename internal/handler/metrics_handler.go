package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func RetrieveMetricsHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Обработка запроса
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		//Получение метрик
		metrics := metricService.GetAllMetrics()
		formattedMetrics := "<pre style=\"word-wrap: break-word; white-space: pre-wrap;\">"
		// Формирование ответа
		for _, m := range metrics {
			formattedMetrics += fmt.Sprintf("# HELP %s\n# TYPE %s %s\n%s %s\n", m.ID, m.ID, m.MType, m.ID, m.StringValue())
		}
		formattedMetrics += "</pre>"

		content := fmt.Sprintf(`<html><head><meta name="color-scheme" content="light dark"></head><body>%s</body></html>`, formattedMetrics)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}
}

func GetMetricHandler(metricService service.MetricService) http.HandlerFunc {
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
		metric, getMetricErr := metricService.GetByName(metricName)
		if getMetricErr != nil {
			switch {
			case errors.Is(getMetricErr, service.ErrMetricNameNotExist):
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusBadRequest)
			}
			return
		}

		// Формирование ответа
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metric.StringValue()))
	}
}

func UpdateMetricHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//Обработка запроса
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusNotFound)
			return
		}

		// Обновление/создание метрики
		metricValue := chi.URLParam(req, "metricValue")
		metric, createMeticErr := metricService.UpdateMetric(metricName, metricType, metricValue)

		slog.Info("GOT METRIC", "type", metricType, "name", metricName, "value", metricValue)
		if createMeticErr != nil {
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

		// Формирование ответа
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metric.StringValue()))
	}
}
