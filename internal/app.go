package internal

import (
	"context"
	"errors"
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
	handler         handler.MetricsHandler
	router          http.Handler
	config          *config.Config
	auditPub        *audit.Publisher
	httpServer      *http.Server
	fileService     service.FileService
	schedulerCancel context.CancelFunc
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
	schedCtx, schedulerCancel := context.WithCancel(context.Background())
	go metricsScheduler.RunWriteMetricsFile(schedCtx)

	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsJSONHandler := handler.NewMetricsJSONHandler(metricsService, cfg)

	router := SetupRouter(metricsHandler, metricsJSONHandler, pool, cfg)

	return &App{
		handler:         metricsHandler,
		router:          router,
		config:          cfg,
		auditPub:        auditPub,
		fileService:     fileService,
		schedulerCancel: schedulerCancel,
		httpServer: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: router,
		},
	}
}

// Run запускает HTTP-сервер на адресе из конфигурации.
func (a *App) Run() error {
	log.Printf("Server starting on %s", a.config.Server.Address)
	err := a.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func (a *App) Shutdown(ctx context.Context) error {
	if a.schedulerCancel != nil {
		a.schedulerCancel()
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	if a.fileService != nil {
		if err := a.fileService.WriteMetricsFile(ctx); err != nil {
			return err
		}
	}
	return a.Close()
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	return a.auditPub.Close()
}
