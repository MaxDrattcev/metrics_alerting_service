package service

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
)

type MetricsFileService struct {
	repo repository.MetricsStorage
	file repository.FileStorage
	cfg  *config.Config
}

func NewMetricsFileService(repo repository.MetricsStorage, file repository.FileStorage, cfg *config.Config) FileService {
	return &MetricsFileService{
		repo: repo,
		file: file,
		cfg:  cfg,
	}
}

func (m *MetricsFileService) WriteMetricsFile(ctx context.Context) error {
	metrics, err := m.repo.GetAllMetrics(ctx)
	if err != nil {
		return err
	}
	if err := m.file.WriteMetrics(metrics); err != nil {
		return err
	}
	return nil
}

func (m *MetricsFileService) LoadMeticsFromFile(ctx context.Context) error {
	if !*m.cfg.Server.Restore {
		return nil
	}
	metrics, err := m.file.ReadMetrics()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if metric.MType == models.Gauge {
			if err := m.repo.UpdateGauge(ctx, metric); err != nil {
				return err
			}
		} else {
			if err := m.repo.UpdateCounter(ctx, metric); err != nil {
				return err
			}
		}
	}
	return nil
}
