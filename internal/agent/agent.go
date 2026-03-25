package agent

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"log"
	"sync"
	"time"
)

type Agent struct {
	collector *MetricsCollector
	sender    *MetricsSender
	cfg       *config.Config
	jobs      chan []models.Metrics
	wg        sync.WaitGroup
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		collector: NewMetricsCollector(),
		sender:    NewMetricsSender(cfg),
		cfg:       cfg,
		jobs:      make(chan []models.Metrics, 3),
	}
}

func (a *Agent) Start(ctx context.Context) {
	a.startWorkers(ctx, a.cfg.Client.RateLimit)
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
	ticker := time.NewTicker(a.cfg.Client.GetReportInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := a.buildSnapshot()
			rateLimit := a.cfg.Client.RateLimit
			if rateLimit == 0 {
				rateLimit = 1
			}

			for i := 0; i < rateLimit; i++ {
				select {
				case <-ctx.Done():
					return
				case a.jobs <- snapshot:
				}
			}
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

func (a *Agent) startWorkers(ctx context.Context, poolSize int) {
	if poolSize <= 0 {
		poolSize = 1
	}
	for i := 0; i < poolSize; i++ {
		a.wg.Add(1)
		go func(WorkerID int) {
			defer a.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case metrics, ok := <-a.jobs:
					if !ok {
						return
					}
					ctxSend, cancel := context.WithTimeout(ctx, 5*time.Second)
					_ = a.sender.SendAllMetricsBuffer(ctxSend, metrics)
					cancel()
				}
			}
		}(i)
	}
}

func (a *Agent) buildSnapshot() []models.Metrics {
	var metrics []models.Metrics
	gauges := a.collector.GetAllGauges()
	for name, value := range gauges {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &value})
	}

	pollCount := a.collector.GetPollCount()
	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &pollCount})
	return metrics
}
