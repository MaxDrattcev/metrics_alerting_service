package agent

import (
	"math/rand"
	"runtime"
	"sync"
)

type MetricsCollector struct {
	metrics   map[string]float64
	pollCount int64
	mu        sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]float64),
	}
}

func (c *MetricsCollector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pollCount++

	c.collectRuntimeMetrics()

	c.collectRandomMetrics()

}

func (c *MetricsCollector) collectRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.metrics["Alloc"] = float64(m.Alloc)
	c.metrics["BuckHashSys"] = float64(m.BuckHashSys)
	c.metrics["Frees"] = float64(m.Frees)
	c.metrics["GCCPUFraction"] = m.GCCPUFraction
	c.metrics["GCSys"] = float64(m.GCSys)
	c.metrics["HeapAlloc"] = float64(m.HeapAlloc)
	c.metrics["HeapIdle"] = float64(m.HeapIdle)
	c.metrics["HeapInuse"] = float64(m.HeapInuse)
	c.metrics["HeapObjects"] = float64(m.HeapObjects)
	c.metrics["HeapReleased"] = float64(m.HeapReleased)
	c.metrics["HeapSys"] = float64(m.HeapSys)
	c.metrics["LastGC"] = float64(m.LastGC)
	c.metrics["Lookups"] = float64(m.Lookups)
	c.metrics["MCacheInuse"] = float64(m.MCacheInuse)
	c.metrics["MCacheSys"] = float64(m.MCacheSys)
	c.metrics["MSpanInuse"] = float64(m.MSpanInuse)
	c.metrics["MSpanSys"] = float64(m.MSpanSys)
	c.metrics["Mallocs"] = float64(m.Mallocs)
	c.metrics["NextGC"] = float64(m.NextGC)
	c.metrics["NumForcedGC"] = float64(m.NumForcedGC)
	c.metrics["NumGC"] = float64(m.NumGC)
	c.metrics["OtherSys"] = float64(m.OtherSys)
	c.metrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	c.metrics["StackInuse"] = float64(m.StackInuse)
	c.metrics["StackSys"] = float64(m.StackSys)
	c.metrics["Sys"] = float64(m.Sys)
	c.metrics["TotalAlloc"] = float64(m.TotalAlloc)
}

func (c *MetricsCollector) collectRandomMetrics() {
	c.metrics["RandomValue"] = rand.Float64() * 1000.0
}

func (c *MetricsCollector) GetAllGauges() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]float64, len(c.metrics))
	for k, v := range c.metrics {
		result[k] = v
	}
	return result
}

func (c *MetricsCollector) GetPollCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pollCount
}
