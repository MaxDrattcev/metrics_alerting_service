package agent

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMetricsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector()

	collector.Collect()

	gauges := collector.GetAllGauges()

	require.NotEmpty(t, gauges, "Metrics should be collected")

	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	for _, metricName := range expectedMetrics {
		_, exists := gauges[metricName]
		assert.True(t, exists, "Metric %s should be collected", metricName)
	}
}

func TestMetricsCollector_PollCount(t *testing.T) {
	collector := NewMetricsCollector()

	initialCount := collector.GetPollCount()
	assert.Equal(t, int64(0), initialCount, "Initial PollCount should be 0")

	collector.Collect()
	count := collector.GetPollCount()
	assert.Equal(t, int64(1), count, "PollCount should be 1 after first Collect")

	collector.Collect()
	count = collector.GetPollCount()
	assert.Equal(t, int64(2), count, "PollCount should be 2 after second Collect")

	collector.Collect()
	count = collector.GetPollCount()
	assert.Equal(t, int64(3), count, "PollCount should be 3 after third Collect")
}

func TestMetricsCollector_RandomValue(t *testing.T) {
	collector := NewMetricsCollector()

	collector.Collect()

	gauges := collector.GetAllGauges()

	randomValue, exists := gauges["RandomValue"]
	require.True(t, exists, "RandomValue should be collected")

	assert.GreaterOrEqual(t, randomValue, 0.0, "RandomValue should be >= 0")
	assert.Less(t, randomValue, 1000.0, "RandomValue should be < 1000")
}

func TestMetricsCollector_GetAllGauges(t *testing.T) {
	collector := NewMetricsCollector()

	gauges := collector.GetAllGauges()
	assert.Empty(t, gauges, "Gauges should be empty before Collect")

	collector.Collect()

	gauges = collector.GetAllGauges()
	assert.NotEmpty(t, gauges, "Gauges should not be empty after Collect")

	gauges["test"] = 999.0
	gauges2 := collector.GetAllGauges()
	_, exists := gauges2["test"]
	assert.False(t, exists, "GetAllGauges should return a copy")
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	collector := NewMetricsCollector()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			collector.Collect()
			collector.GetAllGauges()
			collector.GetPollCount()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	gauges := collector.GetAllGauges()
	assert.NotEmpty(t, gauges, "Metrics should be collected even with concurrent access")

	count := collector.GetPollCount()
	assert.Equal(t, int64(10), count, "PollCount should be 10 after 10 concurrent Collect calls")
}
