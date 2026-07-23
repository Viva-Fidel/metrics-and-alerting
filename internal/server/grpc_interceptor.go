package server

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnetUnaryInterceptor проверяет, что IP агента из метаданных x-real-ip
// принадлежит доверенной подсети. При пустом cidr проверки не выполняются.
func TrustedSubnetUnaryInterceptor(cidr string) (grpc.UnaryServerInterceptor, error) {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return func(
			ctx context.Context,
			req any,
			_ *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (any, error) {
			return handler(ctx, req)
		}, nil
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("parse trusted subnet %q: %w", cidr, err)
	}

	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get(security.RealIPMetadataKey)
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		ip := net.ParseIP(strings.TrimSpace(values[0]))
		if ip == nil || !network.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip not in trusted subnet")
		}

		return handler(ctx, req)
	}, nil
}
