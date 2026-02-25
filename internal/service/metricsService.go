package service

import (
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

func (m *metricsService) UpdateGauge(mType string, mName string, mValue *float64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Value: mValue,
		Hash:  "",
	}

	if err := m.repo.UpdateGauge(metric); err != nil {
		return err
	}
	if *m.cfg.Server.StoreInterval == 0 {
		metrics, err := m.repo.GetAllMetrics()
		if err != nil {
			return err
		}
		if err := m.file.WriteMetrics(metrics); err != nil {
			return err
		}
	}
	return nil
}

func (m *metricsService) UpdateCounter(mType string, mName string, mValue *int64) error {
	var metric = models.Metrics{
		ID:    mName,
		MType: mType,
		Delta: mValue,
		Hash:  "",
	}
	if err := m.repo.UpdateCounter(metric); err != nil {
		return err
	}
	if *m.cfg.Server.StoreInterval == 0 {
		metrics, err := m.repo.GetAllMetrics()
		if err != nil {
			return err
		}
		if err := m.file.WriteMetrics(metrics); err != nil {
			return err
		}
	}
	return nil
}

func (m *metricsService) GetMetric(mType string, mName string) (string, error) {
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

func (m *metricsService) GetAllMetrics() ([]models.Metrics, error) {
	metrics, err := m.repo.GetAllMetrics()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (m *metricsService) WriteMetricsFile() error {
	metrics, err := m.repo.GetAllMetrics()
	if err != nil {
		return err
	}
	if err := m.file.WriteMetrics(metrics); err != nil {
		return err
	}
	return nil
}

func (m *metricsService) LoadMeticsFromFile() error {
	if !*m.cfg.Server.Restore {
		return nil
	}
	metrics, err := m.file.ReadMetrics()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if metric.MType == models.Gauge {
			if err := m.repo.UpdateGauge(metric); err != nil {
				return err
			}
		} else {
			if err := m.repo.UpdateCounter(metric); err != nil {
				return err
			}
		}
	}
	return nil
}
