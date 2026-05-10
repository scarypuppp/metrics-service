package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/scarypuppp/metrics-service/internal/model"
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
		metrics, err := metricService.GetAllMetrics(req.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}
}

func UpdateMetricByURLHandler(metricService service.MetricService) http.HandlerFunc {
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
		metric, createMeticErr := metricService.UpsertMetric(req.Context(), metricName, metricType, metricValue)

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

func UpdateMetricHandler(metricService service.MetricService) http.HandlerFunc {
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

		// Обновление/создание метрики
		_, createMeticErr := metricService.UpsertMetric(req.Context(), metric.ID, metric.MType, metric.StringValue())

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
		w.WriteHeader(http.StatusOK)
	}
}
