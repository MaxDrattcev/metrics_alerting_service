package internal

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/audit"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/scheduler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App — точка входа HTTP-приложения: роутер и конфигурация.
type App struct {
	handler  handler.MetricsHandler
	router   http.Handler
	config   *config.Config
	auditPub *audit.Publisher
}

// NewApp инициализирует хранилище, сервисы, handlers и HTTP-роутер.
func NewApp(cfg *config.Config, pool *pgxpool.Pool) *App {
	var metricsRepo repository.MetricsStorage
	if cfg.Server.DatabaseDSN != "" {
		metricsRepo = repository.NewPostgresStorage(pool)
	} else {
		metricsRepo = repository.NewMemStorage()
	}
	auditPub := audit.NewFromConfig(cfg.Server)

	metricsFile := repository.NewFileStorage(cfg.Server.FileStoragePath)
	metricsService := service.NewMetricsService(metricsRepo, metricsFile, cfg, auditPub)
	fileService := service.NewMetricsFileService(metricsRepo, metricsFile, cfg)

	ctxLoad, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fileService.LoadMeticsFromFile(ctxLoad); err != nil {
		log.Printf("load metrics from file: %v", err)
	}
	metricsScheduler := scheduler.NewMetricsScheduler(cfg, fileService)
	go metricsScheduler.RunWriteMetricsFile(context.Background())

	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsJSONHandler := handler.NewMetricsJSONHandler(metricsService, cfg)

	router := SetupRouter(metricsHandler, metricsJSONHandler, pool)

	return &App{
		handler:  metricsHandler,
		router:   router,
		config:   cfg,
		auditPub: auditPub,
	}
}

// Run запускает HTTP-сервер на адресе из конфигурации.
func (a *App) Run() error {
	log.Printf("Server starting on %s", a.config.Server.Address)
	log.Fatal(http.ListenAndServe(a.config.Server.Address, a.router))
	return nil
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	return a.auditPub.Close()
}
