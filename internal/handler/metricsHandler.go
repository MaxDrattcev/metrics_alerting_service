package handler

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const (
	incorrectValue   = "Incorrect metric value"
	incorrectType    = "Incorrect metric type"
	methodNotAllowed = "Method not allowed"
)

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(service service.MetricsService) MetricsHandler {
	return &metricsHandler{
		service: service,
	}
}
func (m *metricsHandler) Update(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, methodNotAllowed)
		return
	}

	mType, mName, mValue, ok := m.parsePath(c)
	if !ok {
		return
	}

	if !m.validateParam(c, mType, mName, mValue) {
		return
	}

	if mType == models.Gauge {
		floatVal, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, incorrectValue)
			return
		}
		if err := m.service.UpdateGauge(mType, mName, &floatVal); err != nil {
			c.String(http.StatusBadRequest, incorrectValue)
			return
		}
	}

	if mType == models.Counter {
		intVal, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, incorrectValue)
			return
		}
		if err := m.service.UpdateCounter(mType, mName, &intVal); err != nil {
			c.String(http.StatusBadRequest, incorrectValue)
			return
		}
	}
	c.String(http.StatusOK, "")
}

func (m *metricsHandler) parsePath(c *gin.Context) (mType, mName, mValue string, ok bool) {
	mType = c.Param("type")
	mName = c.Param("name")
	mValue = c.Param("value")

	if mType == "" {
		c.String(http.StatusBadRequest, incorrectType)
		return "", "", "", false
	}
	return mType, mName, mValue, true
}

func (m *metricsHandler) validateParam(c *gin.Context, mType, mName, mValue string) bool {
	if mName == "" {
		c.String(http.StatusNotFound, "Name cannot be empty")
		return false
	}
	if mType == "" || mType != models.Gauge && mType != models.Counter {
		c.String(http.StatusBadRequest, incorrectType)
		return false
	}
	if mValue == "" {
		c.String(http.StatusNotFound, "Value cannot be empty")
		return false
	}
	return true
}

func (m *metricsHandler) GetMetric(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusMethodNotAllowed, methodNotAllowed)
		return
	}

	mType := c.Param("type")
	mName := c.Param("name")
	if mType == "" {
		c.String(http.StatusNotFound, "Type cannot be empty")
		return
	}
	if mName == "" {
		c.String(http.StatusNotFound, "Name cannot be empty")
		return
	}

	if mType != models.Gauge && mType != models.Counter {
		c.String(http.StatusNotFound, incorrectType)
		return
	}

	value, err := m.service.GetMetric(mType, mName)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.String(http.StatusOK, value)
}

func (m *metricsHandler) GetAllMetrics(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusMethodNotAllowed, methodNotAllowed)
		return
	}

	metrics, err := m.service.GetAllMetrics()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	type MetricView struct {
		Name  string
		Type  string
		Value string
	}

	views := make([]MetricView, 0, len(metrics))
	for _, metric := range metrics {
		view := MetricView{
			Name: metric.ID,
			Type: metric.MType,
		}

		if metric.MType == models.Gauge {
			if metric.Value != nil {
				view.Value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			} else {
				view.Value = "N/A"
			}
		} else if metric.MType == models.Counter {
			if metric.Delta != nil {
				view.Value = strconv.FormatInt(*metric.Delta, 10)
			} else {
				view.Value = "N/A"
			}
		} else {
			view.Value = "N/A"
		}

		views = append(views, view)
	}

	c.HTML(http.StatusOK, "metrics.html", gin.H{
		"metrics": views,
	})
}
