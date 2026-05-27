package service

import (
	"context"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

// MetricsService — сервис обновления и чтения метрик.
type MetricsService interface {
	// UpdateGauge обновляет одну gauge-метрику.
	UpdateGauge(context.Context, string, string, *float64) error

	// UpdateCounter обновляет одну counter-метрику.
	UpdateCounter(context.Context, string, string, *int64) error

	// GetMetric возвращает значение метрики в виде строки.
	GetMetric(context.Context, string, string) (string, error)

	// GetAllMetrics возвращает все метрики из хранилища.
	GetAllMetrics(context.Context) ([]models.Metrics, error)

	// UpdateMetrics пакетно обновляет метрики (POST /updates).
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
}

// FileService — сервис периодической записи метрик в файл и загрузки при старте.
type FileService interface {
	// WriteMetricsFile сохраняет текущие метрики из хранилища в файл.
	WriteMetricsFile(ctx context.Context) error

	// LoadMeticsFromFile загружает метрики из файла в хранилище при restore.
	LoadMeticsFromFile(ctx context.Context) error
}
