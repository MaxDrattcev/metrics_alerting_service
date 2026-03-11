package scheduler

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"log"
	"time"
)

type MetricsScheduler struct {
	cfg           *config.Config
	metricService service.MetricsService
}

func NewMetricsScheduler(cfg *config.Config, metricService service.MetricsService) *MetricsScheduler {
	return &MetricsScheduler{
		cfg:           cfg,
		metricService: metricService}
}

func (ms *MetricsScheduler) RunWriteMetricsFile() {
	if *ms.cfg.Server.StoreInterval == 0 {
		return
	}
	for {
		time.Sleep(ms.cfg.Server.GetStoreInterval())
		if err := ms.metricService.WriteMetricsFile(); err != nil {
			log.Printf("failed to write metrics file: %v", err)
		}
	}
}
