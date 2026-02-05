package repository

import "github.com/MaxDrattcev/metrics_alerting_service/internal/models"

type MetricsStorage interface {
	UpdateGauge(metric models.Metrics) error

	UpdateCounter(metric models.Metrics) error

	GetMetric(mType string, mName string) (models.Metrics, error)

	GetAllMetrics() ([]models.Metrics, error)
}
