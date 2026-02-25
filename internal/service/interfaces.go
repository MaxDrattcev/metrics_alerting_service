package service

import "github.com/MaxDrattcev/metrics_alerting_service/internal/models"

type MetricsService interface {
	UpdateGauge(string, string, *float64) error

	UpdateCounter(string, string, *int64) error

	GetMetric(string, string) (string, error)

	GetAllMetrics() ([]models.Metrics, error)

	WriteMetricsFile() error

	LoadMeticsFromFile() error
}
