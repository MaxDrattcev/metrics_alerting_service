package agent

import (
	"context"
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
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

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

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcserver.TrustedSubnetInterceptor(""),
		),
	)
	proto.RegisterMetricsServer(grpcServer, grpcserver.NewMetricsGRPCServer(svc))

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(func() {
		grpcServer.Stop()
	})

	clientCfg := &config.Config{
		Client: config.ClientConfig{
			GRPCAddress: lis.Addr().String(),
		},
	}

	sender, err := NewGRPCSender(clientCfg)
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

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer(
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

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(func() {
		grpcServer.Stop()
	})

	sender, err := NewGRPCSender(&config.Config{
		Client: config.ClientConfig{GRPCAddress: lis.Addr().String()},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sender.Close()
	})

	value := 1.0
	err = sender.SendBatch(context.Background(), []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
	})
	require.NoError(t, err)

	if hostIP() != "" {
		require.Equal(t, hostIP(), receivedIP)
	}
}
