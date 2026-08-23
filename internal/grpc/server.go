package grpc

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	models "github.com/scarypuppp/metrics-service/internal/model"
	pb "github.com/scarypuppp/metrics-service/internal/proto"
	"github.com/scarypuppp/metrics-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// realIPMetadataKey is the metadata key agents use to pass their IP address.
const realIPMetadataKey = "x-real-ip"

// MetricsServer объект GRPC сервера метрик.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	addr          string
	metricService *service.MetricService
	trustedSubnet netip.Prefix
	logger        *zap.Logger
}

// NewMetricServer constructs a MetricsServer serving on addr, providing metrics via metricService.
func NewMetricServer(logger *zap.Logger, metricService *service.MetricService, trustedSubnet netip.Prefix, addr string) *MetricsServer {
	return &MetricsServer{
		addr:          addr,
		metricService: metricService,
		trustedSubnet: trustedSubnet,
		logger:        logger,
	}
}

func (ms *MetricsServer) Run() error {
	listen, err := net.Listen("tcp", ms.addr)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации listener: %w", err)
	}
	// Создаем gRPC сервер с интерцептором проверки доверенной подсети
	s := grpc.NewServer(grpc.UnaryInterceptor(trustedSubnetInterceptor(ms.trustedSubnet)))
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
