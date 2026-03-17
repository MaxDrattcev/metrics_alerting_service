package internal

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/scheduler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"time"
)

type App struct {
	handler handler.MetricsHandler
	router  http.Handler
	config  *config.Config
}

func NewApp(cfg *config.Config, pool *pgxpool.Pool) *App {
	var metricsRepo repository.MetricsStorage
	if cfg.Server.DataBaseDSN != "" {
		metricsRepo = repository.NewPostgresStorage(pool)
	} else {
		metricsRepo = repository.NewMemStorage()
	}

	metricsFile := repository.NewFileStorage(cfg.Server.FileStoragePath)
	metricsService := service.NewMetricsService(metricsRepo, metricsFile, cfg)
	fileService := service.NewMetricsFileService(metricsRepo, metricsFile, cfg)

	ctxLoad, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fileService.LoadMeticsFromFile(ctxLoad); err != nil {
		log.Printf("load metrics from file: %v", err)
	}
	metricsScheduler := scheduler.NewMetricsScheduler(cfg, fileService)
	go metricsScheduler.RunWriteMetricsFile(context.Background())

	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsJSONHandler := handler.NewMetricsJSONHandler(metricsService)

	router := SetupRouter(metricsHandler, metricsJSONHandler, pool)

	return &App{
		handler: metricsHandler,
		router:  router,
		config:  cfg,
	}
}

func (a *App) Run() error {
	log.Printf("Server starting on %s", a.config.Server.Address)
	log.Fatal(http.ListenAndServe(a.config.Server.Address, a.router))
	return nil
}
