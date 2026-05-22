package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
)

// MetricsScheduler запускает фоновую запись метрик по StoreInterval.
type MetricsScheduler struct {
	cfg         *config.Config
	fileService service.FileService
}

// NewMetricsScheduler создаёт планировщик записи в файл.
func NewMetricsScheduler(cfg *config.Config, fileService service.FileService) *MetricsScheduler {
	return &MetricsScheduler{
		cfg:         cfg,
		fileService: fileService}
}

// RunWriteMetricsFile в цикле вызывает WriteMetricsFile с интервалом из конфига.
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
