package agent

import (
	"context"
	"fmt"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCSender struct {
	cfg    *config.Config
	client proto.MetricsClient
	conn   *grpc.ClientConn
}

func NewGRPCSender(cfg *config.Config) (*GRPCSender, error) {
	if cfg.Client.GRPCAddress == "" {
		return nil, fmt.Errorf("grpc address is empty")
	}

	conn, err := grpc.NewClient(
		cfg.Client.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial grpc: %w", err)
	}

	return &GRPCSender{
		cfg:    cfg,
		client: proto.NewMetricsClient(conn),
		conn:   conn,
	}, nil
}

func (s *GRPCSender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *GRPCSender) SendBatch(ctx context.Context, metrics []models.Metrics) error {
	protoMetrics := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		protoMetrics = append(protoMetrics, modelToProto(m))
	}

	if ip := hostIP(); ip != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-real-ip", ip))
	}

	req := proto.UpdateMetricsRequest_builder{
		Metrics: protoMetrics,
	}.Build()
	_, err := s.client.UpdateMetrics(ctx, req)

	return err
}

func modelToProto(m models.Metrics) *proto.Metric {
	b := proto.Metric_builder{
		Id: m.ID,
	}
	switch m.MType {
	case models.Gauge:
		b.Type = proto.Metric_GAUGE
		if m.Value != nil {
			b.Value = *m.Value
		}
	case models.Counter:
		b.Type = proto.Metric_COUNTER
		if m.Delta != nil {
			b.Delta = *m.Delta
		}
	}
	return b.Build()
}
