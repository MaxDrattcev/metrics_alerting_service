package agent

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"log"
	"time"
)

type Agent struct {
	collector *MetricsCollector
	sender    *MetricsSender
	cfg       *config.Config
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		collector: NewMetricsCollector(),
		sender:    NewMetricsSender(cfg),
		cfg:       cfg,
	}
}

func (a *Agent) Start(ctx context.Context) {
	go a.startCollecting(ctx)
	go a.startReporting(ctx)
}

func (a *Agent) startCollecting(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.collector.Collect()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.Client.GetPollInterval()):
		}
	}
}

func (a *Agent) startReporting(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.sendMetricsBuffer(ctx)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.Client.GetReportInterval()):
		}
	}
}

func (a *Agent) sendAllMetrics(ctx context.Context) {
	gauges := a.collector.GetAllGauges()
	for name, value := range gauges {
		if err := a.sender.SendGaugeJSON(ctx, name, value); err != nil {
			log.Printf("Failed to send gauge %s: %v", name, err)
		}
	}
	pollCount := a.collector.GetPollCount()
	if err := a.sender.SendCounterJSON(ctx, "PollCount", pollCount); err != nil {
		log.Printf("Failed to send counter PollCounter: %v", err)
	}
}

func (a *Agent) sendMetricsBuffer(ctx context.Context) {
	var metrics []models.Metrics
	gauges := a.collector.GetAllGauges()
	for name, value := range gauges {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &value})
	}
	pollCount := a.collector.GetPollCount()
	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &pollCount})

	if err := a.sender.SendAllMetricsBuffer(ctx, metrics); err != nil {
		log.Printf("Failed to send buffer metrics: %v", err)
	}
}
