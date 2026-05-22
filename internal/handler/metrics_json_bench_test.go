package handler

import (
	"encoding/json"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

func benchmarkUpdatesJSON() []byte {
	v := 123.45
	d := int64(1)
	names := []string{
		"Alloc", "BuckHashSys", "Frees", "HeapAlloc", "HeapSys",
		"RandomValue", "Sys", "TotalAlloc", "Mallocs", "NumGC",
	}
	metrics := make([]models.Metrics, 0, 29)
	for _, name := range names {
		val := v
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &val})
	}
	metrics = append(metrics, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &d})
	body, err := json.Marshal(metrics)
	if err != nil {
		panic(err)
	}
	return body
}

func BenchmarkUnmarshalUpdatesBody(b *testing.B) {
	body := benchmarkUpdatesJSON()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var metrics []models.Metrics
		if err := json.Unmarshal(body, &metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalUpdatesResponse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(map[string]any{}); err != nil {
			b.Fatal(err)
		}
	}
}
