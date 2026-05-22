package repository

import (
	"context"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

// MetricsStorage — интерфейс персистентного хранилища метрик.
type MetricsStorage interface {
	// UpdateGauge сохраняет или обновляет gauge-метрику.
	UpdateGauge(ctx context.Context, metric models.Metrics) error

	// UpdateCounter сохраняет или обновляет counter-метрику (с накоплением delta).
	UpdateCounter(ctx context.Context, metric models.Metrics) error

	// GetMetric возвращает метрику по типу и имени.
	GetMetric(ctx context.Context, mType string, mName string) (models.Metrics, error)

	// GetAllMetrics возвращает все сохранённые метрики.
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)

	// UpdateMetrics пакетно обновляет список метрик.
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
}

// FileStorage — интерфейс файлового снимка метрик (JSON).
type FileStorage interface {
	// WriteMetrics записывает срез метрик в файл.
	WriteMetrics(metrics []models.Metrics) error

	// ReadMetrics читает метрики из файла.
	ReadMetrics() ([]models.Metrics, error)
}
