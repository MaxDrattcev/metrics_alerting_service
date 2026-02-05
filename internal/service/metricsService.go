package service

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"strconv"
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

func (m metricsService) GetMetric(mType string, mName string) (string, error) {
	metric, err := m.repo.GetMetric(mType, mName)
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

func (m metricsService) GetAllMetrics() ([]models.Metrics, error) {
	metrics, err := m.repo.GetAllMetrics()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
