package agent

import (
	"context"

	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"google.golang.org/grpc/metadata"
)

// ReportMetricsGRPC отправляет собранные метрики на сервер по gRPC батчем.
func ReportMetricsGRPC(ctx context.Context, client pb.MetricsClient, metrics *Metrics, hostIP string) {
	gauge, counter := metrics.Snapshot()
	batch := make([]*pb.Metric, 0, len(gauge)+len(counter))

	for name, value := range gauge {
		batch = append(batch, &pb.Metric{
			Id:    name,
			Type:  pb.Metric_GAUGE,
			Value: value,
		})
	}

	for name, value := range counter {
		batch = append(batch, &pb.Metric{
			Id:    name,
			Type:  pb.Metric_COUNTER,
			Delta: value,
		})
	}

	if len(batch) == 0 {
		return
	}

	_ = sendBatchMetricsGRPC(ctx, client, batch, hostIP)
}

func sendBatchMetricsGRPC(ctx context.Context, client pb.MetricsClient, batch []*pb.Metric, hostIP string) bool {
	if client == nil {
		return false
	}

	md := metadata.Pairs(security.RealIPMetadataKey, hostIP)
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: batch})
	return err == nil
}
