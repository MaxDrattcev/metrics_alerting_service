package internal

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"net/http"
	"path/filepath"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter регистрирует маршруты API метрик и middleware.
func SetupRouter(
	metricsHandler handler.MetricsHandler,
	metricsJSONHandler handler.MetricsHandler,
	pool *pgxpool.Pool,
	cfg *config.Config,
) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(
		middleware.Logger(),
		middleware.Decrypt(cfg.Server.CryptoKey),
		middleware.Compress(),
		middleware.ClientIP(),
	)
	if files, err := filepath.Glob("templates/*"); err == nil && len(files) > 0 {
		router.LoadHTMLGlob("templates/*")
	}
	router.POST("/update/:type/:name/:value", metricsHandler.Update)
	router.GET("/value/:type/:name", metricsHandler.GetMetric)
	router.GET("/", metricsHandler.GetAllMetrics)
	router.POST("/update", middleware.TrustedSubnet(cfg.Server.TrustedSubnet), metricsJSONHandler.Update)
	router.POST("/update/", middleware.TrustedSubnet(cfg.Server.TrustedSubnet), metricsJSONHandler.Update)
	router.POST("/value", metricsJSONHandler.GetMetric)
	router.POST("/value/", metricsJSONHandler.GetMetric)
	router.GET("/metrics", metricsJSONHandler.GetAllMetrics)
	router.POST("/updates", middleware.TrustedSubnet(cfg.Server.TrustedSubnet), metricsJSONHandler.UpdateMetrics)
	router.POST("/updates/", middleware.TrustedSubnet(cfg.Server.TrustedSubnet), metricsJSONHandler.UpdateMetrics)
	router.GET("/ping", handler.PingDB(pool))
	return router
}
