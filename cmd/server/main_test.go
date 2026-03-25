package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

	type expectedOutput struct {
		metricValue string
		statusCode  int
	}

	tests := []struct {
		name           string
		inputArgs      inputArgs
		expectedOutput expectedOutput
	}{
		{
			name: "Set counter value #1",
			inputArgs: inputArgs{
				metricType: "counter",
				metricName: "metric1",
				metricVal:  "1",
			},
			expectedOutput: expectedOutput{
				metricValue: "1",
				statusCode:  200,
			},
		},
		{
			name: "Set gauge value #1",
			inputArgs: inputArgs{
				metricType: "gauge",
				metricName: "metric2",
				metricVal:  "3.14",
			},
			expectedOutput: expectedOutput{
				metricValue: "3.14",
				statusCode:  200,
			},
		},
		{
			name: "Empty metric name",
			inputArgs: inputArgs{
				metricType: "gauge",
				metricName: "",
				metricVal:  "3.14",
			},
			expectedOutput: expectedOutput{
				metricValue: "",
				statusCode:  404,
			},
		},
		{
			name: "Empty metric type",
			inputArgs: inputArgs{
				metricType: "",
				metricName: "metric1",
				metricVal:  "3.14",
			},
			expectedOutput: expectedOutput{
				metricValue: "",
				statusCode:  404,
			},
		},
		{
			name: "Unexpected metric value",
			inputArgs: inputArgs{
				metricType: "gauge",
				metricName: "metric1",
				metricVal:  "none",
			},
			expectedOutput: expectedOutput{
				metricValue: "",
				statusCode:  400,
			},
		},
		{
			name: "Invalid type",
			inputArgs: inputArgs{
				metricType: "unknown",
				metricName: "metric1",
				metricVal:  "none",
			},
			expectedOutput: expectedOutput{
				metricValue: "",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := repository.MemStorage{Metrics: make(map[string]models.Metrics)}
			metricService := service.MetricService{Storage: &storage}
			handler := handlers.CreateMetricHandler(metricService)

			request := httptest.NewRequest(http.MethodPost, "/update/{metricType}/{metricName}/{metricValue}", nil)
			request.SetPathValue("metricType", test.inputArgs.metricType)
			request.SetPathValue("metricName", test.inputArgs.metricName)
			request.SetPathValue("metricValue", test.inputArgs.metricVal)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			handler(w, request)
			res := w.Result()

			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			fmt.Println(string(resBody))
			require.NoError(t, err)

			require.Equal(t, test.expectedOutput.statusCode, res.StatusCode)
			if res.StatusCode == 200 {
				require.Equal(t, test.expectedOutput.metricValue, strings.TrimRight(string(resBody), "0"))
			}

		})
	}
}
