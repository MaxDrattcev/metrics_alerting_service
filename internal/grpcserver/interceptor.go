package grpcserver

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/middleware"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"strings"
	"time"
)

const realIPMetadataKey = "x-real-ip"

var grpcLog *zap.Logger

func init() {
	var err error
	grpcLog, err = zap.NewDevelopment()
	if err != nil {
		grpcLog = zap.NewNop()
	}
}

// LoggerInterceptor логирует gRPC-запросы: метод, статус, длительность, IP.
func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("code", code.String()),
			zap.Duration("duration", time.Since(start)),
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ips := md.Get(realIPMetadataKey); len(ips) > 0 {
				fields = append(fields, zap.String("x-real-ip", ips[0]))
			}
		}
		if r, ok := req.(*proto.UpdateMetricsRequest); ok {
			fields = append(fields, zap.Int("metrics_count", len(r.GetMetrics())))
		}
		if err != nil {
			grpcLog.Warn("grpc request", append(fields, zap.Error(err))...)
		} else {
			grpcLog.Info("grpc request", fields...)
		}
		return resp, err
	}
}

// TrustedSubnetInterceptor проверяет, что IP агента из metadata входит в trusted_subnet.
// При пустом trustedSubnet ограничения не применяются.
func TrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	trustedSubnet = strings.TrimSpace(trustedSubnet)
	if trustedSubnet == "" {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	_, network, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		// Некорректный CIDR в конфиге сервера.
		return func(
			ctx context.Context,
			req any,
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (any, error) {
			return nil, status.Error(codes.Internal, "invalid trusted subnet")
		}
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		ips := md.Get(realIPMetadataKey)
		if len(ips) == 0 || strings.TrimSpace(ips[0]) == "" {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		ipStr := strings.TrimSpace(ips[0])
		ip := net.ParseIP(ipStr)
		if ip == nil || !network.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip is not in trusted subnet")
		}

		ctx = middleware.ContextWithClientIP(ctx, ipStr)

		return handler(ctx, req)
	}
}
