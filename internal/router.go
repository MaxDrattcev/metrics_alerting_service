package internal

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"path/filepath"
)

func SetupRouter(metricsHandler handler.MetricsHandler, metricsJSONHandler handler.MetricsHandler, pool *pgxpool.Pool) http.Handler {
	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(middleware.Logger(), middleware.Compress())

	if files, err := filepath.Glob("templates/*"); err == nil && len(files) > 0 {
		router.LoadHTMLGlob("templates/*")
	}

	router.POST("/update/:type/:name/:value", metricsHandler.Update)
	router.GET("/value/:type/:name", metricsHandler.GetMetric)
	router.GET("/", metricsHandler.GetAllMetrics)

	router.POST("/update", metricsJSONHandler.Update)
	router.POST("/value", metricsJSONHandler.GetMetric)
	router.GET("/metrics", metricsJSONHandler.GetAllMetrics)

	router.GET("/ping", handler.PingDb(pool))

	return router
}
