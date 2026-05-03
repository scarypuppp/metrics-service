package agent

import (
	"bytes"
	"compress/gzip"
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
	if err != nil {
		fmt.Println("failed make json:", err)
	}
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		fmt.Println("failed to compress data:", err)
	}
	_, err = gz.Write(body)
	if err != nil {
		fmt.Println("failed to write compressed data:", err)
	}
	err = gz.Close()
	if err != nil {
		fmt.Println("failed to compress data:", err)
	}

	request, err := http.NewRequest(http.MethodPost, url, &b)
	if err != nil {
		fmt.Println("failed to make request:", err)
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	request.Header.Set("Content-Encoding", "gzip")
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
