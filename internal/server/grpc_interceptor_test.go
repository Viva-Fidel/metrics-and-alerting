package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
)

func TestTrustedSubnetUnaryInterceptor_EmptyAllowsAll(t *testing.T) {
	interceptor, err := TrustedSubnetUnaryInterceptor("")
	require.NoError(t, err)

	called := false
	_, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestTrustedSubnetUnaryInterceptor_AllowsInsideSubnet(t *testing.T) {
	interceptor, err := TrustedSubnetUnaryInterceptor("192.168.1.0/24")
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(security.RealIPMetadataKey, "192.168.1.10"))
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)
}

func TestTrustedSubnetUnaryInterceptor_RejectsOutsideSubnet(t *testing.T) {
	interceptor, err := TrustedSubnetUnaryInterceptor("192.168.1.0/24")
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(security.RealIPMetadataKey, "10.0.0.1"))
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnetUnaryInterceptor_RejectsMissingMetadata(t *testing.T) {
	interceptor, err := TrustedSubnetUnaryInterceptor("192.168.1.0/24")
	require.NoError(t, err)

	_, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnetUnaryInterceptor_InvalidCIDR(t *testing.T) {
	_, err := TrustedSubnetUnaryInterceptor("not-a-cidr")
	require.Error(t, err)
}
