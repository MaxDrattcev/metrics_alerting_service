package internal

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"log"
	"net/http"
)

type App struct {
	handler handler.MetricsHandler
	router  http.Handler
	config  *config.Config
}

func NewApp(cfg *config.Config) *App {
	metricsRepo := repository.NewMetricsStorage()
	metricsService := service.NewMetricsService(metricsRepo)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsJSONHandler := handler.NewMetricsJSONHandler(metricsService)

	router := SetupRouter(metricsHandler, metricsJSONHandler)

	return &App{
		handler: metricsHandler,
		router:  router,
		config:  cfg,
	}
}

func (a *App) Run() error {
	log.Printf("Server starting on %s", a.config.Server.Address)
	return http.ListenAndServe(a.config.Server.Address, a.router)
}
