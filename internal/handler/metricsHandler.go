package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func CreateMetricHandler(metricService service.MetricService) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		metricType := req.PathValue("metricType")
		metricName := req.PathValue("metricName")
		if metricName == "" || metricType == "" {
			http.Error(w, "metricName and metricType is required", http.StatusNotFound)
			return
		}
		metricValue := req.PathValue("metricValue")

		metric, createMeticErr := metricService.CreateMetric(metricName, metricType, metricValue)

		if createMeticErr != nil {
			switch {
			case errors.Is(createMeticErr, strconv.ErrSyntax):
				errorMessage := fmt.Sprintf("Error parsing metric: %s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
				return
			case errors.Is(createMeticErr, service.InvalidMetricTypeProvided):
				errorMessage := fmt.Sprintf("Error creating metric: %s", createMeticErr)
				http.Error(w, errorMessage, http.StatusBadRequest)
				return
			default:
				errorMessage := fmt.Sprintf("Error creating metric: %s", createMeticErr)
				http.Error(w, errorMessage, http.StatusInternalServerError)
				return
			}
		}

		var response string
		if metric.MType == models.Gauge {
			response = fmt.Sprintf("%f", *metric.Value)
		} else {
			response = fmt.Sprintf("%d", *metric.Delta)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}
}
