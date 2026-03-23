package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/scarypuppp/metrics-service/internal/handler"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMetricHandler(t *testing.T) {
	type inputArgs struct {
		metricType string
		metricName string
		metricVal  string
	}

	tests := []struct {
		name                string
		inputArgs           inputArgs
		expectedMetricValue string
	}{
		{
			name: "metric value",
			inputArgs: inputArgs{
				metricType: "counter",
				metricName: "metric1",
				metricVal:  "1",
			},
			expectedMetricValue: "1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := repository.MemStorage{Metrics: make(map[string]models.Metrics)}
			metricService := service.MetricService{Storage: &storage}
			handler := handlers.CreateMetricHandler(metricService)

			request := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu_usage/3.14", nil)
			request.SetPathValue("metric_type", test.inputArgs.metricType)
			request.SetPathValue("metric_name", test.inputArgs.metricName)
			request.SetPathValue("metric_value", test.inputArgs.metricVal)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)
			res := w.Result()

			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			fmt.Println(string(resBody))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}
