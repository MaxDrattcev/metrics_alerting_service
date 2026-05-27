package handler

import (
	"github.com/gin-gonic/gin"
)

// MetricsHandler — HTTP-обработчики API метрик.
type MetricsHandler interface {
	// Update обрабатывает обновление одной метрики (JSON или legacy path).
	Update(c *gin.Context)
	// GetMetric возвращает значение одной метрики.
	GetMetric(c *gin.Context)
	// GetAllMetrics возвращает все метрики.
	GetAllMetrics(c *gin.Context)
	// UpdateMetrics обрабатывает пакетное обновление (POST /updates).
	UpdateMetrics(c *gin.Context)
}
