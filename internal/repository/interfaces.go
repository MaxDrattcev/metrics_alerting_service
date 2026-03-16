package repository

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

type MetricsStorage interface {
	UpdateGauge(ctx context.Context, metric models.Metrics) error

	UpdateCounter(ctx context.Context, metric models.Metrics) error

	GetMetric(ctx context.Context, mType string, mName string) (models.Metrics, error)

	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)

	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
}

type FileStorage interface {
	WriteMetrics(metrics []models.Metrics) error

	ReadMetrics() ([]models.Metrics, error)
}
