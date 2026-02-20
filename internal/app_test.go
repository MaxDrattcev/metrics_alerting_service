package internal

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewApp(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "localhost:8080",
		},
	}

	app := NewApp(cfg)

	require.NotNil(t, app)
	assert.NotNil(t, app.handler)
	assert.NotNil(t, app.router)
	assert.Equal(t, cfg, app.config)
}

func TestSetupRouter(t *testing.T) {
	mockHandler := &mockMetricsHandler{}

	router := SetupRouter(mockHandler, mockHandler)

	require.NotNil(t, router)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, mockHandler.called)
}

type mockMetricsHandler struct {
	called bool
}

func (m *mockMetricsHandler) Update(c *gin.Context) {
	m.called = true
	c.String(http.StatusOK, "")
}

func (m *mockMetricsHandler) GetMetric(c *gin.Context) {
	c.String(http.StatusOK, "123.45")
}

func (m *mockMetricsHandler) GetAllMetrics(c *gin.Context) {
	c.String(http.StatusOK, "metrics")
}
