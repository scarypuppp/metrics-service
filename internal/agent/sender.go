package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	models "github.com/scarypuppp/metrics-service/internal/model"
	"github.com/scarypuppp/metrics-service/internal/utils/hash"
	"github.com/scarypuppp/metrics-service/internal/utils/retry"
)

// Sender delivers metrics to the server over HTTP with gzip compression and optional HMAC signing.
type Sender struct {
	client      *resty.Client
	key         string
	certificate *x509.Certificate
}

// NewSender configures the HTTP client for the given server URL and returns a Sender.
func NewSender(
	client *resty.Client,
	baseURL string,
	key string,
	publicKey *x509.Certificate,
	host string,
) *Sender {
	client.SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", host)
	return &Sender{client, key, publicKey}
}

// SendMetric posts a single metric to the server as compressed JSON, retrying on network errors.
func (s *Sender) SendMetric(metric models.Metrics) error {
	return retry.Do(
		func() error {
			body, err := json.Marshal(metric)
			if err != nil {
				return fmt.Errorf("failed to marshal metric: %w", err)
			}
			bodyHash := hash.GetHash(body, s.key)
			// Сжимаем данные
			dataToSend, err := compressData(body)
			if err != nil {
				return fmt.Errorf("failed to compress data: %w", err)
			}
			// Шифруем данные
			if s.certificate != nil {
				dataToSend, err = encryptData(s.certificate, dataToSend)
				if err != nil {
					return fmt.Errorf("failed to encrypt data: %w", err)
				}
			}
			// Строим запрос
			req := s.client.R().
				SetHeader("Content-Encoding", "gzip").
				SetBody(dataToSend)
			if s.key != "" {
				req.SetHeader("HashSHA256", bodyHash)
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

func encryptData(certificate *x509.Certificate, data []byte) ([]byte, error) {
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, certificate.PublicKey.(*rsa.PublicKey), data)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}
