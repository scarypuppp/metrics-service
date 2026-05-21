package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/utils/retry"
)

type Sender struct {
	client *resty.Client
	key    string
}

func NewSender(client *resty.Client, baseURL string, key string) *Sender {
	client.SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetContentLength(true)
	return &Sender{client, key}
}

func (s *Sender) SendMetric(metric models.Metrics) error {
	return retry.Do(
		func() error {
			body, err := json.Marshal(metric)
			if err != nil {
				return fmt.Errorf("failed to marshal metric: %w", err)
			}
			// Сжимаем данные
			compressedBody, err := compressData(body)
			if err != nil {
				return fmt.Errorf("failed to compress data: %w", err)
			}
			// Строим запрос
			req := s.client.R().
				SetHeader("Content-Encoding", "gzip").
				SetBody(compressedBody)
			if s.key != "" {
				req.SetHeader("HashSHA256", getHash(compressedBody, s.key))
			}
			// Делаем запрос
			resp, err := req.Post("/update/")
			if err != nil {
				return fmt.Errorf("failed to send metric: %w", err)
			}
			// Валидируем ответ
			if resp.IsError() {
				return fmt.Errorf("server returned %d: %s", resp.StatusCode(), resp.String())
			}
			return nil
		},
		retry.IsNetworkError,
	)
}

func getHash(body []byte, key string) string {
	h := sha256.New()
	h.Write(body)
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func compressData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write(data)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
