package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

// Agent координирует сбор и отправку метрик по таймерам.
type Agent struct {
	collector *MetricsCollector
	sender    *MetricsSender
	cfg       *config.Config
	jobs      chan []models.Metrics
	wg        sync.WaitGroup
}

// NewAgent создаёт агента с конфигурацией клиента.
func NewAgent(cfg *config.Config) (*Agent, error) {
	sender, err := NewMetricsSender(cfg)
	if err != nil {
		return nil, err
	}
	return &Agent{
		collector: NewMetricsCollector(),
		sender:    sender,
		cfg:       cfg,
		jobs:      make(chan []models.Metrics, 3),
	}, nil
}

// Start запускает сбор метрик по PollInterval, периодическую отправку
// снимка по ReportInterval и пул воркеров (RateLimit). Завершение — по отмене ctx.
func (a *Agent) Start(ctx context.Context) {
	a.startWorkers(a.cfg.Client.RateLimit)
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
			select {
			case <-ctx.Done():
				return
			case a.jobs <- snapshot:
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

func (a *Agent) startWorkers(poolSize int) {
	if poolSize <= 0 {
		poolSize = 1
	}
	for i := 0; i < poolSize; i++ {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			for {
				metrics, ok := <-a.jobs
				if !ok {
					return
				}
				ctxSend, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := a.sender.SendAllMetricsBuffer(ctxSend, metrics); err != nil {
					log.Printf("Failed to send buffer metrics: %v", err)
				}
				cancel()
			}
		}()
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

func (a *Agent) Shutdown(ctx context.Context) error {
	snapshot := a.buildSnapshot()
	if err := a.sender.SendAllMetricsBuffer(ctx, snapshot); err != nil {
		log.Printf("final metrics send : %v", err)
	}

	close(a.jobs)

	done := make(chan struct{})

	go func() {
		a.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
