package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
)

func TestGRPCUpdateMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := repository.NewMemRepository(logger, "", 0, false)
	svc := service.NewMetricsService(repo)
	pub := audit.NewPublisher(logger, "", "")
	t.Cleanup(func() { _ = pub.Close() })

	gs, err := NewGRPCServer("127.0.0.1:0", svc, pub, "127.0.0.0/8", logger)
	require.NoError(t, err)
	require.NotNil(t, gs)

	go func() { _ = gs.Serve() }()
	t.Cleanup(gs.Stop)

	addr := gs.listener.Addr().String()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := pb.NewMetricsClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(security.RealIPMetadataKey, "127.0.0.1"))

	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 1.5},
			{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 3},
		},
	})
	require.NoError(t, err)

	g, ok := repo.GetGauge("Alloc")
	require.True(t, ok)
	assert.Equal(t, 1.5, g)

	c, ok := repo.GetCounter("PollCount")
	require.True(t, ok)
	assert.Equal(t, int64(3), c)

	ctxBad := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(security.RealIPMetadataKey, "10.0.0.1"))
	_, err = client.UpdateMetrics(ctxBad, &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{{Id: "X", Type: pb.Metric_GAUGE, Value: 1}},
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestNewGRPCServer_EmptyAddress(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gs, err := NewGRPCServer("", nil, nil, "", logger)
	require.NoError(t, err)
	assert.Nil(t, gs)
}

func TestNewGRPCServer_Listen(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	repo := repository.NewMemRepository(logger, "", 0, false)
	svc := service.NewMetricsService(repo)
	gs, err := NewGRPCServer(addr, svc, nil, "", logger)
	require.NoError(t, err)
	require.NotNil(t, gs)

	go func() { _ = gs.Serve() }()
	time.Sleep(20 * time.Millisecond)
	gs.GracefulStop()
}
