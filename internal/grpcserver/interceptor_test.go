package grpcserver

import (
	"context"
	"go.uber.org/zap"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func callInterceptor(
	t *testing.T,
	interceptor grpc.UnaryServerInterceptor,
	ctx context.Context,
) (any, error) {
	t.Helper()

	handler := func(ctx context.Context, req any) (any, error) {
		return middleware.ClientIPFromContext(ctx), nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"}
	return interceptor(ctx, nil, info, handler)
}

func TestTrustedSubnetInterceptor_EmptySubnet(t *testing.T) {
	interceptors := UnaryInterceptors(zap.NewNop(), "")
	require.Len(t, interceptors, 1)
}

func TestUnaryInterceptors_WithSubnet(t *testing.T) {
	interceptors := UnaryInterceptors(zap.NewNop(), "192.168.0.0/24")
	require.Len(t, interceptors, 2)
}

func TestTrustedSubnetInterceptor_AllowedIP(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.0.0/24")

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "192.168.0.107"),
	)

	resp, err := callInterceptor(t, interceptor, ctx)
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.107", resp)
}

func TestTrustedSubnetInterceptor_ForbiddenIP(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.0.0/24")

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "8.8.8.8"),
	)

	_, err := callInterceptor(t, interceptor, ctx)
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnetInterceptor_MissingMetadata(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.0.0/24")

	_, err := callInterceptor(t, interceptor, context.Background())
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnetInterceptor_MissingRealIP(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.0.0/24")

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "value"))

	_, err := callInterceptor(t, interceptor, ctx)
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnetInterceptor_InvalidCIDR(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("invalid-cidr")

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "192.168.0.107"),
	)

	_, err := callInterceptor(t, interceptor, ctx)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestLoggerInterceptor_PassesThrough(t *testing.T) {
	interceptor := LoggerInterceptor(zap.NewNop())

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"}
	resp, err := interceptor(context.Background(), nil, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}
