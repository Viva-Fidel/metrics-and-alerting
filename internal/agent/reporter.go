package agent

import (
	"context"
	"crypto/rsa"
	"log/slog"

	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"resty.dev/v3"
)

// MetricsReporter отвечает за отправку батча собранных метрик.
// Нужен, чтобы вызывающий код не передавал в Run сразу два клиента
// (один из которых всегда nil).
type MetricsReporter interface {
	Report(ctx context.Context, metrics *Metrics) error
}

type httpReporter struct {
	logger    *slog.Logger
	client    *resty.Client
	hashKey   string
	publicKey *rsa.PublicKey
}

func NewHTTPReporter(logger *slog.Logger, client *resty.Client, hashKey string, publicKey *rsa.PublicKey) MetricsReporter {
	return &httpReporter{
		logger:    logger,
		client:    client,
		hashKey:   hashKey,
		publicKey: publicKey,
	}
}

func (r *httpReporter) Report(ctx context.Context, metrics *Metrics) error {
	if r == nil || r.client == nil {
		return nil
	}
	ReportMetrics(ctx, r.client, metrics, r.hashKey, r.publicKey)
	return nil
}

type grpcReporter struct {
	logger    *slog.Logger
	client    pb.MetricsClient
	hostIP    string
}

func NewGRPCReporter(logger *slog.Logger, client pb.MetricsClient, hostIP string) MetricsReporter {
	return &grpcReporter{
		logger: logger,
		client: client,
		hostIP: hostIP,
	}
}

func (r *grpcReporter) Report(ctx context.Context, metrics *Metrics) error {
	if r == nil || r.client == nil {
		return nil
	}
	return ReportMetricsGRPC(ctx, r.client, metrics, r.hostIP)
}

