package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type Sender struct {
	client  *http.Client
	baseURL string
}

func NewSender(client *http.Client, baseUrl string) *Sender {
	return &Sender{client, baseUrl}
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	url := fmt.Sprintf("%s/update/", s.baseURL)

	body, err := json.Marshal(metric)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("failed to make request:", err)
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := s.client.Do(request)
	if err != nil {
		fmt.Println("failed to send metric:", err)
		return err
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("failed to read response:", err)
		return err
	}
	responseText := string(bodyBytes)
	if response.StatusCode != http.StatusOK {
		fmt.Println("failed to send metric", response.StatusCode, "-", responseText)
	}
	return nil
}
