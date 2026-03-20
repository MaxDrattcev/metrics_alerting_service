package scheduler

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"log"
	"time"
)

type MetricsScheduler struct {
	cfg         *config.Config
	fileService service.FileService
}

func NewMetricsScheduler(cfg *config.Config, fileService service.FileService) *MetricsScheduler {
	return &MetricsScheduler{
		cfg:         cfg,
		fileService: fileService}
}

func (ms *MetricsScheduler) RunWriteMetricsFile(ctx context.Context) {
	if *ms.cfg.Server.StoreInterval == 0 {
		return
	}
	ticker := time.NewTicker(ms.cfg.Server.GetStoreInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := ms.fileService.WriteMetricsFile(writeCtx)
			cancel()
			if err != nil {
				log.Printf("failed to write metrics file: %v", err)
			}
		}
	}
}
