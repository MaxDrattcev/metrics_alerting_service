package handler

import (
	"encoding/json"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/hasher"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strconv"
)

type metricsJSONHandler struct {
	service service.MetricsService
	cfg     *config.Config
}

func NewMetricsJSONHandler(service service.MetricsService, cfg *config.Config) MetricsHandler {
	return &metricsJSONHandler{
		service: service,
		cfg:     cfg,
	}
}

func (m *metricsJSONHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

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

	if m.cfg.Server.Key != "" && c.GetHeader("HashSHA256") != "" {
		hash, err := hasher.ComputeHashSHA256(body, m.cfg.Server.Key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if hash != c.GetHeader("HashSHA256") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash"})
			return
		}
	}

	var metric models.Metrics
	if err := json.Unmarshal(body, &metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !m.validateRequest(c, metric) {
		return
	}
	if metric.MType == models.Gauge {
		if err := m.service.UpdateGauge(ctx, metric.MType, metric.ID, metric.Value); err != nil {
			log.Printf("UpdateGauge: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
	}
	if metric.MType == models.Counter {
		if err := m.service.UpdateCounter(ctx, metric.MType, metric.ID, metric.Delta); err != nil {
			log.Printf("UpdateCounter: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (m *metricsJSONHandler) validateRequest(c *gin.Context, metric models.Metrics) bool {
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

func (m *metricsJSONHandler) GetMetric(c *gin.Context) {
	ctx := c.Request.Context()

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

	value, err := m.service.GetMetric(ctx, metric.MType, metric.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if metric.MType == models.Gauge {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("ParseFloat in GetMetric: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		metric.Value = &f
		metric.Delta = nil
	}
	if metric.MType == models.Counter {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Printf("ParseInt in GetMetric: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		metric.Delta = &i
		metric.Value = nil
	}
	metric.Hash = ""
	c.JSON(http.StatusOK, metric)
}

func (m *metricsJSONHandler) GetAllMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := m.service.GetAllMetrics(ctx)
	if err != nil {
		log.Printf("GetAllMetrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (m *metricsJSONHandler) UpdateMetrics(c *gin.Context) {
	ctx := c.Request.Context()

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
	if m.cfg.Server.Key != "" && c.GetHeader("HashSHA256") != "" {
		hash, err := hasher.ComputeHashSHA256(body, m.cfg.Server.Key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if hash != c.GetHeader("HashSHA256") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash"})
			return
		}
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(body, &metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, metric := range metrics {
		if !m.validateRequest(c, metric) {
			return
		}
	}

	if err := m.service.UpdateMetrics(ctx, metrics); err != nil {
		log.Printf("UpdateMetrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	c.Status(http.StatusOK)
}
