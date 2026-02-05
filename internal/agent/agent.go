package agent

import (
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

func (a *Agent) Start() {
	go a.startCollecting()

	go a.startReporting()
}

func (a *Agent) startCollecting() {
	for {
		a.collector.Collect()
		time.Sleep(a.cfg.Client.GetPollInterval())
	}
}

func (a *Agent) startReporting() {
	for {
		a.sendAllMetrics()
		time.Sleep(a.cfg.Client.GetReportInterval())
	}
}

func (a *Agent) sendAllMetrics() {
	gauges := a.collector.GetAllGauges()
	for name, value := range gauges {
		if err := a.sender.SendGauge(name, value); err != nil {
			log.Printf("Failed to send gauge %s: %v", name, err)
		}
	}
	pollCount := a.collector.GetPollCount()
	if err := a.sender.SendCounter("PollCount", pollCount); err != nil {
		log.Printf("Failed to send counter PollCounter: %v", err)
	}
}
