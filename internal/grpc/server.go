package grpc

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/netip"

	models "github.com/scarypuppp/metrics-service/internal/model"
	pb "github.com/scarypuppp/metrics-service/internal/proto"
	"github.com/scarypuppp/metrics-service/internal/service"
	"github.com/scarypuppp/metrics-service/internal/utils/hash"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	// realIPMetadataKey is the metadata key agents use to pass their IP address.
	realIPMetadataKey = "x-real-ip"
	// hashMetadataKey is the metadata key agents use to pass the HMAC signature of the request.
	hashMetadataKey = "hashsha256"
)

// MetricsServer объект GRPC сервера метрик.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	addr          string
	metricService *service.MetricService
	trustedSubnet netip.Prefix
	key           string
	logger        *zap.Logger
}

// NewMetricServer constructs a MetricsServer serving on addr, providing metrics via metricService.
func NewMetricServer(logger *zap.Logger, metricService *service.MetricService, trustedSubnet netip.Prefix, key string, addr string) *MetricsServer {
	return &MetricsServer{
		addr:          addr,
		metricService: metricService,
		trustedSubnet: trustedSubnet,
		key:           key,
		logger:        logger,
	}
}

func (ms *MetricsServer) Run() error {
	listen, err := net.Listen("tcp", ms.addr)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации listener: %w", err)
	}
	// Создаем gRPC сервер с интерцепторами проверки доверенной подсети и подписи запроса
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		trustedSubnetInterceptor(ms.trustedSubnet),
		validateHashInterceptor(ms.key),
	))
	// Регистрируем сервис
	pb.RegisterMetricsServer(s, ms)

	ms.logger.Info("grpc server running", zap.String("address", listen.Addr().String()))
	if err := s.Serve(listen); err != nil {
		return fmt.Errorf("ошибка при работе сервера: %w", err)
	}
	return nil
}

func trustedSubnetInterceptor(prefix netip.Prefix) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !prefix.IsValid() {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		ips := md.Get(realIPMetadataKey)
		if len(ips) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		ip, err := netip.ParseAddr(ips[0])
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "invalid x-real-ip metadata")
		}

		if !prefix.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip is not in trusted subnet")
		}

		return handler(ctx, req)
	}
}

// validateHashInterceptor проверяет HMAC-подпись запроса, если у сервера задан ключ.
func validateHashInterceptor(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if key == "" {
			return handler(ctx, req)
		}

		message, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "request is not a proto message")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing hashsha256 metadata")
		}
		hashes := md.Get(hashMetadataKey)
		if len(hashes) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing hashsha256 metadata")
		}

		body, err := proto.MarshalOptions{Deterministic: true}.Marshal(message)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal request")
		}

		expectedHash := hash.GetHash(body, key)
		if subtle.ConstantTimeCompare([]byte(hashes[0]), []byte(expectedHash)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid request hash")
		}

		return handler(ctx, req)
	}
}

func (ms *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	protoMetrics := in.GetMetrics()
	metrics := make([]models.Metrics, 0, len(protoMetrics))
	for _, m := range protoMetrics {
		metric, err := fromProtoMetric(m)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid metric %q: %v", m.GetId(), err)
		}
		metrics = append(metrics, metric)
	}

	if err := ms.metricService.UpsertMetrics(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save metrics: %v", err)
	}

	return pb.UpdateMetricsResponse_builder{}.Build(), nil
}

func fromProtoMetric(m *pb.Metric) (models.Metrics, error) {
	metric := models.Metrics{ID: m.GetId()}
	switch m.GetType() {
	case pb.Metric_GAUGE:
		metric.MType = models.Gauge
		value := m.GetValue()
		metric.Value = &value
	case pb.Metric_COUNTER:
		metric.MType = models.Counter
		delta := m.GetDelta()
		metric.Delta = &delta
	default:
		return models.Metrics{}, fmt.Errorf("unknown metric type: %v", m.GetType())
	}
	return metric, nil
}
