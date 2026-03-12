package agent

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
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
			a.sendAllMetrics()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.Client.GetReportInterval()):
		}
	}
}

func (a *Agent) sendAllMetrics() {
	gauges := a.collector.GetAllGauges()
	for name, value := range gauges {
		if err := a.sender.SendGaugeJSON(name, value); err != nil {
			log.Printf("Failed to send gauge %s: %v", name, err)
		}
	}
	pollCount := a.collector.GetPollCount()
	if err := a.sender.SendCounterJSON("PollCount", pollCount); err != nil {
		log.Printf("Failed to send counter PollCounter: %v", err)
	}
}
