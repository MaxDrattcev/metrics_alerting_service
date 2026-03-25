package agent

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"math/rand"
	"runtime"
	"sync"
)

type MetricsCollector struct {
	metrics      map[string]float64
	pollCount    int64
	mu           sync.RWMutex
	prevCPUTimes cpu.TimesStat
	cpuReady     bool
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

	c.collectGopsutilMetrics()

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

func (c *MetricsCollector) collectGopsutilMetrics() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory info: %v", err)
		return
	}
	c.metrics["TotalMemory"] = float64(vm.Total)
	c.metrics["FreeMemory"] = float64(vm.Free)

	times, err := cpu.Times(false)
	if err != nil || len(times) == 0 {
		log.Printf("Error getting cpu times: %v", err)
		return
	}
	cur := times[0]
	if c.cpuReady {
		c.metrics["CPUutilization1"] = c.cpuUtilizationDelta(c.prevCPUTimes, cur)
	} else {
		c.metrics["CPUutilization1"] = 0
	}
	c.prevCPUTimes = cur
	c.cpuReady = true
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

func (c *MetricsCollector) cpuUtilizationDelta(prev, cur cpu.TimesStat) float64 {
	prevIdle := prev.Idle + prev.Iowait
	curIdle := cur.Idle + cur.Iowait
	prevTotal := prev.User + prev.System + prev.Nice + prev.Idle + prev.Iowait +
		prev.Irq + prev.Softirq + prev.Steal
	curTotal := cur.User + cur.System + cur.Nice + cur.Idle + cur.Iowait +
		cur.Irq + cur.Softirq + cur.Steal
	dTotal := curTotal - prevTotal
	if dTotal <= 0 {
		return 0
	}
	return (dTotal - (curIdle - prevIdle)) / dTotal
}
