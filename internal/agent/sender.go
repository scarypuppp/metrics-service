package agent

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type Sender struct {
	Client  *http.Client
	BaseURL string
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	var stringValue = ""
	switch metric.MType {
	case models.Gauge:
		stringValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case models.Counter:
		stringValue = strconv.FormatInt(*metric.Delta, 10)
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	url := fmt.Sprintf("%s/update/%s/%s/%s", s.BaseURL, metric.MType, metric.ID, stringValue)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("failed to make request:", err)
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := s.Client.Do(request)
	if err != nil {
		fmt.Println("failed to send metric:", err) // просто логируем, не паникуем
		return err
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("failed to read response: %w", err)
		return err
	}
	responseText := string(bodyBytes)
	if response.StatusCode != http.StatusOK {
		fmt.Println("failed to send metric", response.StatusCode, "-", responseText)
	}
	return nil
}
