package handler

import (
	"encoding/json"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

type metricsJsonHandler struct {
	service service.MetricsService
}

func NewMetricsJsonHandler(service service.MetricsService) MetricsHandler {
	return &metricsJsonHandler{
		service: service,
	}
}

func (m *metricsJsonHandler) Update(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": methodNotAllowed})
		return
	}
	if c.GetHeader("Content-Type") != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/json"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer c.Request.Body.Close()

	var metric models.Metrics
	if err := json.Unmarshal(body, &metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !m.validateRequest(c, metric) {
		return
	}
	if metric.MType == models.Gauge {
		if err := m.service.UpdateGauge(metric.MType, metric.ID, metric.Value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if metric.MType == models.Counter {
		if err := m.service.UpdateCounter(metric.MType, metric.ID, metric.Delta); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (m *metricsJsonHandler) validateRequest(c *gin.Context, metric models.Metrics) bool {
	if metric.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Name cannot be empty"})
		return false
	}
	if metric.MType == "" || metric.MType != models.Gauge && metric.MType != models.Counter {
		c.JSON(http.StatusBadRequest, gin.H{"error": incorrectType})
		return false
	}
	if metric.MType == models.Gauge {
		if metric.Value == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Value cannot be empty"})
			return false
		}
	}
	if metric.MType == models.Counter {
		if metric.Delta == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Delta cannot be empty"})
			return false
		}
	}
	return true
}

func (m *metricsJsonHandler) GetMetric(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": methodNotAllowed})
		return
	}
	if c.GetHeader("Content-Type") != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/json"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer c.Request.Body.Close()
	var metric models.Metrics
	if err := json.Unmarshal(body, &metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if metric.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Name cannot be empty"})
		return
	}
	if metric.MType == "" || metric.MType != models.Gauge && metric.MType != models.Counter {
		c.JSON(http.StatusBadRequest, gin.H{"error": incorrectType})
		return
	}

	value, err := m.service.GetMetric(metric.MType, metric.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if metric.MType == models.Gauge {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		metric.Value = &f
		metric.Delta = nil
	}
	if metric.MType == models.Counter {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		metric.Delta = &i
		metric.Value = nil
	}
	metric.Hash = ""
	c.JSON(http.StatusOK, metric)

}

func (m *metricsJsonHandler) GetAllMetrics(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": methodNotAllowed})
		return
	}
	metrics, err := m.service.GetAllMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, metrics)
}
