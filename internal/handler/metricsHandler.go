package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func RetrieveMetricsHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Request received")
		// Обработка запроса
		if req.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		// Формирование ответа
		metrics := *metricService.GetAllMetrics()
		formattedMetrics := "<ul>"

		for _, m := range metrics {
			formattedMetrics += fmt.Sprintf("<li>%s (%s): %s</li>", m.ID, m.MType, m.StringValue())
		}
		formattedMetrics += "</ul>"

		content := fmt.Sprintf("<html><body><h1>Метрики:</h1>%s</body></html>", formattedMetrics)
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
		metricType := strings.ToLower(chi.URLParam(req, "metricType"))
		metricName := strings.ToLower(chi.URLParam(req, "metricName"))
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusNotFound)
			return
		}

		// Получение матрики
		metric, getMetricErr := metricService.GetByName(metricName)
		if getMetricErr != nil {
			if errors.Is(getMetricErr, service.ErrMetricNameNotExist) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			} else {
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

func UpdateMetricHandler(metricService service.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//Обработка запроса
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		metricType := strings.ToLower(chi.URLParam(req, "metricType"))
		metricName := strings.ToLower(chi.URLParam(req, "metricName"))
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusNotFound)
			return
		}

		// Обновление/создание метрики
		metricValue := strings.ToLower(chi.URLParam(req, "metricValue"))
		metric, createMeticErr := metricService.UpdateMetric(metricName, metricType, metricValue)

		fmt.Printf("GOT METRIC: %s %s %s\n", metricType, metricName, metricValue)
		if createMeticErr != nil {
			if errors.Is(createMeticErr, strconv.ErrSyntax) ||
				errors.Is(createMeticErr, service.ErrInvalidMetricType) ||
				errors.Is(createMeticErr, service.ErrMetricTypeMismatch) {
				errorMessage := fmt.Sprintf("%s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
			} else {
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
