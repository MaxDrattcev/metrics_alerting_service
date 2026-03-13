package service

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

type MetricsService interface {
	UpdateGauge(context.Context, string, string, *float64) error

	UpdateCounter(context.Context, string, string, *int64) error

	GetMetric(context.Context, string, string) (string, error)

	GetAllMetrics(context.Context) ([]models.Metrics, error)

	WriteMetricsFile(ctx context.Context) error

	LoadMeticsFromFile(ctx context.Context) error
}
