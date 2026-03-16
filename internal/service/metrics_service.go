package service

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"strconv"
)

type metricsService struct {
	repo repository.MetricsStorage
	file repository.FileStorage
	cfg  *config.Config
}

func NewMetricsService(repo repository.MetricsStorage, file repository.FileStorage, cfg *config.Config) MetricsService {
	return &metricsService{
		repo: repo,
		file: file,
		cfg:  cfg,
	}
}

func (m *metricsService) UpdateGauge(ctx context.Context, mType string, mName string, mValue *float64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Value: mValue,
		Hash:  "",
	}

	if err := m.repo.UpdateGauge(ctx, metric); err != nil {
		return err
	}
	if *m.cfg.Server.StoreInterval == 0 {
		metrics, err := m.repo.GetAllMetrics(ctx)
		if err != nil {
			return err
		}
		if err := m.file.WriteMetrics(metrics); err != nil {
			return err
		}
	}
	return nil
}

func (m *metricsService) UpdateCounter(ctx context.Context, mType string, mName string, mValue *int64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Delta: mValue,
		Hash:  "",
	}
	if err := m.repo.UpdateCounter(ctx, metric); err != nil {
		return err
	}
	if *m.cfg.Server.StoreInterval == 0 {
		metrics, err := m.repo.GetAllMetrics(ctx)
		if err != nil {
			return err
		}
		if err := m.file.WriteMetrics(metrics); err != nil {
			return err
		}
	}
	return nil
}

func (m *metricsService) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	metric, err := m.repo.GetMetric(ctx, mType, mName)
	if err != nil {
		return "", err
	}
	if metric.MType == models.Gauge {
		if metric.Value == nil {
			return "", fmt.Errorf("gauge metric value is nil")
		}
		value := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		return value, nil
	}
	if metric.Delta == nil {
		return "", fmt.Errorf("counter metric delta is nil")
	}
	delta := strconv.FormatInt(*metric.Delta, 10)
	return delta, nil
}

func (m *metricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics, err := m.repo.GetAllMetrics(ctx)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (m *metricsService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	if err := m.repo.UpdateMetrics(ctx, metrics); err != nil {
		return err
	}
	return nil
}
