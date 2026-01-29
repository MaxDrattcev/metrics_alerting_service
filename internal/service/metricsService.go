package service

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
)

type metricsService struct {
	repo repository.MetricsStorage
}

func NewMetricsService(repo repository.MetricsStorage) MetricsService {
	return metricsService{
		repo: repo,
	}
}

func (m metricsService) UpdateGauge(mType string, mName string, mValue *float64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Value: mValue,
		Hash:  "",
	}

	if err := m.repo.UpdateGauge(metric); err != nil {
		return err
	}
	return nil
}

func (m metricsService) UpdateCounter(mType string, mName string, mValue *int64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Delta: mValue,
		Hash:  "",
	}
	if err := m.repo.UpdateCounter(metric); err != nil {
		return err
	}
	return nil
}
