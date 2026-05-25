package service

import (
	"context"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
)

func benchConfig(b *testing.B, storeInterval int64) *config.Config {
	b.Helper()

	return &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
			Restore:       ptrBool(false),
		},
	}
}

func ptrBool(v bool) *bool { return &v }

func serviceWithMem(b *testing.B, storeInterval int64) MetricsService {
	b.Helper()

	repo := repository.NewMemStorage()
	file := &noopFileStorage{}
	return NewMetricsService(repo, file, benchConfig(b, storeInterval), nil)
}

type noopFileStorage struct{}

func (n *noopFileStorage) WriteMetrics([]models.Metrics) error { return nil }

func (n *noopFileStorage) ReadMetrics() ([]models.Metrics, error) {
	return nil, nil
}

func benchmarkServiceBatch(b *testing.B) []models.Metrics {
	b.Helper()

	full := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}
	metrics := make([]models.Metrics, 0, len(full)+1)
	v := 100.0
	for _, name := range full {
		val := v
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &val})
	}
	d := int64(1)
	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &d})
	return metrics
}

func BenchmarkMetricsService_UpdateMetrics(b *testing.B) {
	svc := serviceWithMem(b, 300)
	ctx := context.Background()
	batch := benchmarkServiceBatch(b)

	b.ReportAllocs()
	for b.Loop() {
		metrics := make([]models.Metrics, len(batch))
		copy(metrics, batch)
		if err := svc.UpdateMetrics(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMetricsService_UpdateGauge(b *testing.B) {
	svc := serviceWithMem(b, 300)
	ctx := context.Background()
	v := 42.0

	b.ReportAllocs()
	for b.Loop() {
		if err := svc.UpdateGauge(ctx, models.Gauge, "Alloc", &v); err != nil {
			b.Fatal(err)
		}
	}
}
