package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/mocks"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetricsGRPCServer_UpdateMetrics_Success(t *testing.T) {
	mockSvc := mocks.NewMockMetricsService(t)
	server := NewMetricsGRPCServer(mockSvc)

	req := proto.UpdateMetricsRequest_builder{
		Metrics: []*proto.Metric{
			proto.Metric_builder{
				Id:    "Alloc",
				Type:  proto.Metric_GAUGE,
				Value: 123.45,
			}.Build(),
			proto.Metric_builder{
				Id:    "PollCount",
				Type:  proto.Metric_COUNTER,
				Delta: 5,
			}.Build(),
		},
	}.Build()

	mockSvc.EXPECT().
		UpdateMetrics(mock.Anything, mock.Anything).
		Run(func(_ context.Context, metrics []models.Metrics) {
			require.Len(t, metrics, 2)
			require.Equal(t, "Alloc", metrics[0].ID)
			require.Equal(t, models.Gauge, metrics[0].MType)
			require.NotNil(t, metrics[0].Value)
			require.Equal(t, 123.45, *metrics[0].Value)
			require.Equal(t, "PollCount", metrics[1].ID)
			require.Equal(t, models.Counter, metrics[1].MType)
			require.NotNil(t, metrics[1].Delta)
			require.Equal(t, int64(5), *metrics[1].Delta)
		}).
		Return(nil)

	resp, err := server.UpdateMetrics(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMetricsGRPCServer_UpdateMetrics_EmptyRequest(t *testing.T) {
	server := NewMetricsGRPCServer(mocks.NewMockMetricsService(t))

	_, err := server.UpdateMetrics(context.Background(), nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	req := proto.UpdateMetricsRequest_builder{Metrics: nil}.Build()
	_, err = server.UpdateMetrics(context.Background(), req)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestMetricsGRPCServer_UpdateMetrics_InvalidMetric(t *testing.T) {
	server := NewMetricsGRPCServer(mocks.NewMockMetricsService(t))

	req := proto.UpdateMetricsRequest_builder{
		Metrics: []*proto.Metric{
			proto.Metric_builder{
				Id:   "",
				Type: proto.Metric_GAUGE,
			}.Build(),
		},
	}.Build()

	_, err := server.UpdateMetrics(context.Background(), req)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestMetricsGRPCServer_UpdateMetrics_ServiceError(t *testing.T) {
	mockSvc := mocks.NewMockMetricsService(t)
	server := NewMetricsGRPCServer(mockSvc)

	req := proto.UpdateMetricsRequest_builder{
		Metrics: []*proto.Metric{
			proto.Metric_builder{
				Id:    "Alloc",
				Type:  proto.Metric_GAUGE,
				Value: 1.0,
			}.Build(),
		},
	}.Build()

	mockSvc.EXPECT().
		UpdateMetrics(mock.Anything, mock.Anything).
		Return(errors.New("db error"))

	_, err := server.UpdateMetrics(context.Background(), req)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestProtoToModel_Gauge(t *testing.T) {
	pm := proto.Metric_builder{
		Id:    "TestGauge",
		Type:  proto.Metric_GAUGE,
		Value: 42.5,
	}.Build()

	metric, err := protoToModel(pm)
	require.NoError(t, err)
	require.Equal(t, "TestGauge", metric.ID)
	require.Equal(t, models.Gauge, metric.MType)
	require.NotNil(t, metric.Value)
	require.Equal(t, 42.5, *metric.Value)
}

func TestProtoToModel_Counter(t *testing.T) {
	pm := proto.Metric_builder{
		Id:    "PollCount",
		Type:  proto.Metric_COUNTER,
		Delta: 10,
	}.Build()

	metric, err := protoToModel(pm)
	require.NoError(t, err)
	require.Equal(t, "PollCount", metric.ID)
	require.Equal(t, models.Counter, metric.MType)
	require.NotNil(t, metric.Delta)
	require.Equal(t, int64(10), *metric.Delta)
}
