package repository

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"sync"
)

type MemStorage struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

func NewMemStorage() MetricsStorage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (m *MemStorage) UpdateGauge(ctx context.Context, metric models.Metrics) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if metric.Value == nil {
		return fmt.Errorf("gauge metric requires value")
	}
	key := m.key(metric.ID, metric.MType)
	m.metrics[key] = metric
	return nil
}

func (m *MemStorage) UpdateCounter(ctx context.Context, metric models.Metrics) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.key(metric.ID, metric.MType)
	if m.exists(key) {
		existing := m.metrics[key]
		if existing.Delta != nil {
			*metric.Delta += *existing.Delta
		}
	}
	m.metrics[key] = metric
	return nil
}

func (m *MemStorage) key(mName, mType string) string {
	return fmt.Sprintf("%s:%s", mName, mType)
}

func (m *MemStorage) exists(key string) bool {
	_, exists := m.metrics[key]
	return exists
}

func (m *MemStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return models.Metrics{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.key(mName, mType)
	metric, ok := m.metrics[key]
	if !ok {
		return models.Metrics{}, fmt.Errorf("metric not found")
	}
	return metric, nil
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return []models.Metrics{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := make([]models.Metrics, 0, len(m.metrics))

	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
