package repository

import "github.com/MaxDrattcev/metrics_alerting_service/internal/models"

type MetricsStorage interface {
	UpdateGauge(metric models.Metrics) error

	UpdateCounter(metric models.Metrics) error
}
