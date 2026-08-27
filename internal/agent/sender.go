package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	models "github.com/scarypuppp/metrics-service/internal/model"
	pb "github.com/scarypuppp/metrics-service/internal/proto"
	"github.com/scarypuppp/metrics-service/internal/utils/hash"
	"github.com/scarypuppp/metrics-service/internal/utils/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Метаданные, которые агент передаёт серверу вместе с gRPC запросом.
const (
	realIPMetadataKey = "x-real-ip"
	hashMetadataKey   = "hashsha256"
)

type TransportType string

const (
	TransportHTTP TransportType = "HTTP"
	TransportGRPC TransportType = "GRPC"
)

// Sender delivers metrics to the server over HTTP with gzip compression and optional HMAC signing or over gRPC.
type Sender struct {
	client      *resty.Client
	grpcClient  pb.MetricsClient
	key         string
	certificate *x509.Certificate
	host        string
	transport   TransportType
}

// NewSender configures the HTTP client for the given server URL and returns a Sender.
func NewSender(
	client *resty.Client,
	grpcClient *grpc.ClientConn,
	baseURL string,
	key string,
	publicKey *x509.Certificate,
	host string,
	transport TransportType,
) *Sender {
	if transport == TransportHTTP {
		client.SetBaseURL(baseURL).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetHeader("X-Real-IP", host)
	}

	var metricsClient pb.MetricsClient
	if transport == TransportGRPC && grpcClient != nil {
		metricsClient = pb.NewMetricsClient(grpcClient)
	}

	return &Sender{client, metricsClient, key, publicKey, host, transport}
}

// SendMetrics delivers a batch of metrics to the server.
func (s *Sender) SendMetrics(metrics []models.Metrics) error {
	switch s.transport {
	case TransportHTTP:
		for _, metric := range metrics {
			if err := s.sendMetricHTTP(metric); err != nil {
				return err
			}
		}
		return nil
	case TransportGRPC:
		return s.sendMetricsGRPC(metrics)
	default:
		return fmt.Errorf("invalid transport type: %s", s.transport)
	}
}

func (s *Sender) sendMetricHTTP(metric models.Metrics) error {
	return retry.Do(
		func() error {
			body, err := json.Marshal(metric)
			if err != nil {
				return fmt.Errorf("failed to marshal metric: %w", err)
			}
			bodyHash := hash.GetHash(body, s.key)
			dataToSend, err := s.prepareData(body)
			if err != nil {
				return fmt.Errorf("failed to prepare data: %w", err)
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

func (s *Sender) sendMetricsGRPC(metrics []models.Metrics) error {
	protoMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		protoMetric, err := toProtoMetric(metric)
		if err != nil {
			return fmt.Errorf("failed to convert metric: %w", err)
		}
		protoMetrics = append(protoMetrics, protoMetric)
	}
	req := pb.UpdateMetricsRequest_builder{Metrics: protoMetrics}.Build()

	ctx := metadata.AppendToOutgoingContext(context.Background(), realIPMetadataKey, s.host)
	if s.key != "" {
		body, err := proto.MarshalOptions{Deterministic: true}.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, hashMetadataKey, hash.GetHash(body, s.key))
	}

	return retry.Do(
		func() error {
			if _, err := s.grpcClient.UpdateMetrics(ctx, req); err != nil {
				return fmt.Errorf("failed to send metrics via grpc: %w", err)
			}
			return nil
		},
		retry.IsNetworkError,
	)
}

func toProtoMetric(m models.Metrics) (*pb.Metric, error) {
	b := pb.Metric_builder{Id: m.ID}
	switch m.MType {
	case models.Gauge:
		b.Type = pb.Metric_GAUGE
		if m.Value != nil {
			b.Value = *m.Value
		}
	case models.Counter:
		b.Type = pb.Metric_COUNTER
		if m.Delta != nil {
			b.Delta = *m.Delta
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", m.MType)
	}
	return b.Build(), nil
}

func (s *Sender) prepareData(data []byte) ([]byte, error) {
	// Сжимаем данные
	dataToSend, err := compressData(data)
	if err != nil {
		return nil, err
	}
	// Шифруем данные
	if s.certificate != nil {
		dataToSend, err = encryptData(s.certificate, dataToSend)
		if err != nil {
			return nil, err
		}
	}
	return dataToSend, nil
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
