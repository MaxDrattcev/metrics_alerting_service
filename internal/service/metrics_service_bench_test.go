package service

import (
	"context"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
)

func benchConfig(storeInterval int64) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
			Restore:       ptrBool(false),
		},
	}
}

func ptrBool(v bool) *bool { return &v }

func serviceWithMem(storeInterval int64) MetricsService {
	repo := repository.NewMemStorage()
	file := &noopFileStorage{}
	return NewMetricsService(repo, file, benchConfig(storeInterval), nil) // nil = без аудита (iter16)
}

type noopFileStorage struct{}

func (n *noopFileStorage) WriteMetrics([]models.Metrics) error { return nil }
func (n *noopFileStorage) ReadMetrics() ([]models.Metrics, error) {
	return nil, nil
}

func benchmarkServiceBatch() []models.Metrics {
	gaugeNames := []string{
		"Alloc", "Frees", "HeapAlloc", "RandomValue", "Sys", "TotalAlloc",
	}
	metrics := make([]models.Metrics, 0, 30)
	v := 100.0
	for _, name := range gaugeNames {
		val := v
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &val})
	}

	full := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}
	metrics = metrics[:0]
	for _, name := range full {
		val := v
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &val})
	}
	d := int64(1)
	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &d})
	return metrics
}

func BenchmarkMetricsService_UpdateMetrics(b *testing.B) {
	svc := serviceWithMem(300)
	ctx := context.Background()
	batch := benchmarkServiceBatch()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := make([]models.Metrics, len(batch))
		copy(metrics, batch)
		if err := svc.UpdateMetrics(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMetricsService_UpdateGauge(b *testing.B) {
	svc := serviceWithMem(300)
	ctx := context.Background()
	v := 42.0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := svc.UpdateGauge(ctx, models.Gauge, "Alloc", &v); err != nil {
			b.Fatal(err)
		}
	}
}
