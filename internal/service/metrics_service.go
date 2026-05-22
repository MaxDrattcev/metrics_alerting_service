package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/audit"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/middleware"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
)

type metricsService struct {
	repo  repository.MetricsStorage
	file  repository.FileStorage
	cfg   *config.Config
	audit *audit.Publisher
}

// NewMetricsService создаёт сервис метрик с указанным хранилищем, файлом и аудитом.
func NewMetricsService(repo repository.MetricsStorage, file repository.FileStorage, cfg *config.Config,
	auditPub *audit.Publisher) MetricsService {
	return &metricsService{
		repo:  repo,
		file:  file,
		cfg:   cfg,
		audit: auditPub,
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

	m.notifyAudit(ctx, []string{mName})

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

	m.notifyAudit(ctx, []string{mName})

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
	names := make([]string, 0, len(metrics))
	for _, met := range metrics {
		names = append(names, met.ID)
	}
	m.notifyAudit(ctx, names)
	return nil
}

func (m *metricsService) notifyAudit(ctx context.Context, metricNames []string) {
	if m.audit == nil {
		return
	}
	m.audit.Publish(audit.Event{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: middleware.ClientIPFromContext(ctx),
	})
}
