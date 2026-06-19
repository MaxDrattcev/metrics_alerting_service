package grpcserver

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// MetricsGRPCServer реализует gRPC-сервис Metrics.
type MetricsGRPCServer struct {
	proto.UnimplementedMetricsServer
	service service.MetricsService
}

// NewMetricsGRPCServer создаёт gRPC handler для метрик.
func NewMetricsGRPCServer(service service.MetricsService) *MetricsGRPCServer {
	return &MetricsGRPCServer{service: service}
}

// UpdateMetrics принимает батч метрик и сохраняет их через service layer.
func (s *MetricsGRPCServer) UpdateMetrics(ctx context.Context,
	req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if req == nil || len(req.GetMetrics()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "metrics list is empty")
	}

	metrics := make([]models.Metrics, 0, len(req.GetMetrics()))
	for _, pm := range req.GetMetrics() {
		metric, err := protoToModel(pm)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		metrics = append(metrics, metric)
	}

	if err := s.service.UpdateMetrics(ctx, metrics); err != nil {
		log.Printf("grpc UpdateMetrics: %v", err)
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func protoToModel(pm *proto.Metric) (models.Metrics, error) {
	if pm == nil {
		return models.Metrics{}, fmt.Errorf("metric is nil")
	}
	if pm.GetId() == "" {
		return models.Metrics{}, fmt.Errorf("metric id cannot be empty")
	}
	switch pm.GetType() {
	case proto.Metric_GAUGE:
		value := pm.GetValue()
		return models.Metrics{
			ID:    pm.GetId(),
			MType: models.Gauge,
			Value: &value,
		}, nil
	case proto.Metric_COUNTER:
		delta := pm.GetDelta()
		return models.Metrics{
			ID:    pm.GetId(),
			MType: models.Counter,
			Delta: &delta,
		}, nil
	default:
		return models.Metrics{}, fmt.Errorf("unsupported metric type")
	}
}
