package server

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"google.golang.org/grpc"
)

// GRPCServer оборачивает gRPC-сервер метрик.
type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	logger   *slog.Logger
}

// NewGRPCServer создаёт gRPC-сервер с сервисом Metrics и проверкой trusted_subnet.
func NewGRPCServer(
	addr string,
	metricsService *service.MetricsService,
	auditPublisher *audit.Publisher,
	trustedSubnet string,
	logger *slog.Logger,
) (*GRPCServer, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return &GRPCServer{logger: logger}, nil
	}

	interceptor, err := TrustedSubnetUnaryInterceptor(trustedSubnet)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen gRPC %q: %w", addr, err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	pb.RegisterMetricsServer(srv, NewMetricsGRPCServer(metricsService, auditPublisher))

	return &GRPCServer{
		server:   srv,
		listener: listener,
		logger:   logger,
	}, nil
}

// Serve запускает обработку входящих gRPC-соединений.
func (s *GRPCServer) Serve() error {
	if s == nil || s.server == nil || s.listener == nil {
		return nil
	}
	s.logger.Info("gRPC server listening", slog.String("addr", s.listener.Addr().String()))
	return s.server.Serve(s.listener)
}

// GracefulStop останавливает gRPC-сервер с ожиданием активных запросов.
func (s *GRPCServer) GracefulStop() {
	if s == nil || s.server == nil {
		return
	}
	s.server.GracefulStop()
}

// Stop принудительно останавливает gRPC-сервер.
func (s *GRPCServer) Stop() {
	if s == nil || s.server == nil {
		return
	}
	s.server.Stop()
}
