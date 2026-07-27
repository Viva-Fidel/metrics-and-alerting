package server

import (
	"context"
	"fmt"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetricsGRPCServer реализует pb.MetricsServer.
type MetricsGRPCServer struct {
	pb.UnimplementedMetricsServer
	metricsService *service.MetricsService
	auditPublisher *audit.Publisher
}

// NewMetricsGRPCServer создаёт gRPC-обработчик метрик.
func NewMetricsGRPCServer(metricsService *service.MetricsService, auditPublisher *audit.Publisher) *MetricsGRPCServer {
	return &MetricsGRPCServer{
		metricsService: metricsService,
		auditPublisher: auditPublisher,
	}
}

// UpdateMetrics принимает батч метрик и сохраняет их.
func (s *MetricsGRPCServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return &pb.UpdateMetricsResponse{}, nil
	}

	batch := make([]payload.MetricsJSON, 0, len(req.Metrics))
	names := make([]string, 0, len(req.Metrics))

	for _, m := range req.Metrics {
		if m == nil {
			continue
		}

		item, err := protoMetricToPayload(m)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		batch = append(batch, item)
		names = append(names, item.ID)
	}

	if err := s.metricsService.SetMetricsJSONBatch(batch); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if s.auditPublisher != nil {
		s.auditPublisher.Notify(names, clientIPFromContext(ctx))
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func protoMetricToPayload(m *pb.Metric) (payload.MetricsJSON, error) {
	switch m.Type {
	case pb.Metric_GAUGE:
		value := m.Value
		return payload.MetricsJSON{
			ID:    m.Id,
			MType: service.Gauge,
			Value: &value,
		}, nil
	case pb.Metric_COUNTER:
		delta := m.Delta
		return payload.MetricsJSON{
			ID:    m.Id,
			MType: service.Counter,
			Delta: &delta,
		}, nil
	default:
		return payload.MetricsJSON{}, fmt.Errorf("unknown metric type: %v", m.Type)
	}
}

func clientIPFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(security.RealIPMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
