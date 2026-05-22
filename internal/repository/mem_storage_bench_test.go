package repository

import (
	"context"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

func benchmarkBatchMetrics() []models.Metrics {
	gaugeNames := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}
	metrics := make([]models.Metrics, 0, len(gaugeNames)+1)
	v := 123.45
	for _, name := range gaugeNames {
		val := v
		metrics = append(metrics, models.Metrics{
			ID: name, MType: models.Gauge, Value: &val,
		})
	}
	d := int64(1)
	metrics = append(metrics, models.Metrics{
		ID: "PollCount", MType: models.Counter, Delta: &d,
	})
	return metrics
}

func fillMemStorage(b *testing.B, n int) *MemStorage {
	b.Helper()
	s := NewMemStorage().(*MemStorage)
	batch := benchmarkBatchMetrics()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		_ = s.UpdateMetrics(ctx, batch)
	}
	return s
}

func BenchmarkMemStorage_UpdateGauge(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	v := 42.0
	metric := models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &v}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.UpdateGauge(ctx, metric); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemStorage_UpdateCounter(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	d := int64(1)
	metric := models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &d}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.UpdateCounter(ctx, metric); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemStorage_UpdateMetrics(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	batch := benchmarkBatchMetrics()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := make([]models.Metrics, len(batch))
		copy(metrics, batch)
		if err := s.UpdateMetrics(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemStorage_GetAllMetrics(b *testing.B) {
	s := fillMemStorage(b, 100)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.GetAllMetrics(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
