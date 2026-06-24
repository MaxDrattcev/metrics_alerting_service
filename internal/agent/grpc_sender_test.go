package agent

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/grpccreds"
	"net"
	"testing"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/grpcserver"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/mocks"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewGRPCSender_EmptyAddress(t *testing.T) {
	_, err := NewGRPCSender(&config.Config{})
	require.Error(t, err)
}

func TestModelToProto_Gauge(t *testing.T) {
	value := 12.34
	pm := modelToProto(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &value,
	})

	require.Equal(t, "Alloc", pm.GetId())
	require.Equal(t, proto.Metric_GAUGE, pm.GetType())
	require.Equal(t, 12.34, pm.GetValue())
}

func TestModelToProto_Counter(t *testing.T) {
	delta := int64(7)
	pm := modelToProto(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &delta,
	})

	require.Equal(t, "PollCount", pm.GetId())
	require.Equal(t, proto.Metric_COUNTER, pm.GetType())
	require.Equal(t, int64(7), pm.GetDelta())
}

func TestGRPCSender_SendBatch(t *testing.T) {
	storeInterval := int64(300)
	restore := false
	cfg := &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
			Restore:       &restore,
		},
	}
	repo := repository.NewMemStorage()
	mockFile := mocks.NewMockFileStorage(t)
	svc := service.NewMetricsService(repo, mockFile, cfg, nil)
	addr, certFile := startTestGRPCServer(t, func(s *grpc.Server) {
		proto.RegisterMetricsServer(s, grpcserver.NewMetricsGRPCServer(svc))
	})
	sender, err := NewGRPCSender(&config.Config{
		Client: config.ClientConfig{
			GRPCAddress: addr,
			GRPCCert:    certFile,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sender.Close()
	})
	value := 99.9
	delta := int64(3)
	err = sender.SendBatch(context.Background(), []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := repo.GetMetric(ctx, models.Gauge, "Alloc")
	require.NoError(t, err)
	require.NotNil(t, got.Value)
	require.Equal(t, 99.9, *got.Value)
	gotCounter, err := repo.GetMetric(ctx, models.Counter, "PollCount")
	require.NoError(t, err)
	require.NotNil(t, gotCounter.Delta)
	require.Equal(t, int64(3), *gotCounter.Delta)
}

func TestGRPCSender_SendBatch_SendsRealIPMetadata(t *testing.T) {
	var receivedIP string
	certFile, keyFile := grpccreds.WriteTestSelfSignedCert(t)
	serverCreds, err := grpccreds.ServerCredentials(certFile, keyFile)
	require.NoError(t, err)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(func(
			ctx context.Context,
			req any,
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (any, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			ips := md.Get("x-real-ip")
			if len(ips) > 0 {
				receivedIP = ips[0]
			}
			return &proto.UpdateMetricsResponse{}, nil
		}),
	)
	proto.RegisterMetricsServer(grpcServer, grpcserver.NewMetricsGRPCServer(mocks.NewMockMetricsService(t)))
	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })
	sender, err := NewGRPCSender(&config.Config{
		Client: config.ClientConfig{
			GRPCAddress: lis.Addr().String(),
			GRPCCert:    certFile,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sender.Close() })
	value := 1.0
	err = sender.SendBatch(context.Background(), []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
	})
	require.NoError(t, err)
	if hostIP() != "" {
		require.Equal(t, hostIP(), receivedIP)
	}
}

func startTestGRPCServer(t *testing.T, register func(*grpc.Server)) (addr, certFile string) {
	t.Helper()
	certFile, keyFile := grpccreds.WriteTestSelfSignedCert(t)
	creds, err := grpccreds.ServerCredentials(certFile, keyFile)
	require.NoError(t, err)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	register(grpcServer)
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(func() {
		grpcServer.Stop()
	})
	return lis.Addr().String(), certFile
}

func TestNewGRPCSender_EmptyCert(t *testing.T) {
	_, err := NewGRPCSender(&config.Config{
		Client: config.ClientConfig{
			GRPCAddress: "localhost:8081",
			GRPCCert:    "",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc tls")
}
