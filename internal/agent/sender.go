package agent

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/utils/retry"
)

type Sender struct {
	client  *resty.Client
	baseURL string
}

func NewSender(client *resty.Client, baseURL string) *Sender {
	client.SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetContentLength(true)

	return &Sender{client, baseURL}
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	return retry.Do(
		func() error {
			resp, err := s.client.R().
				SetBody(metric).
				Post("/update/")
			if err != nil {
				return fmt.Errorf("failed to send metric: %w", err)
			}
			if resp.IsError() {
				return fmt.Errorf("server returned %d: %s", resp.StatusCode(), resp.String())
			}
			return nil
		},
		retry.IsNetworkError,
	)
}
