package handler

import (
	"github.com/gin-gonic/gin"
)

type MetricsHandler interface {
	Update(c *gin.Context)

	GetMetric(c *gin.Context)

	GetAllMetrics(c *gin.Context)
}
