package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	cfg := &config.Config{Client: config.ClientConfig{Address: "localhost:8080"}}
	agt, err := NewAgent(cfg)
	require.NoError(t, err)
	require.NotNil(t, agt)
}

func TestAgent_buildSnapshot(t *testing.T) {
	cfg := &config.Config{Client: config.ClientConfig{Address: "localhost:8080"}}
	agt, err := NewAgent(cfg)
	require.NoError(t, err)

	agt.collector.Collect()
	snapshot := agt.buildSnapshot()

	require.NotEmpty(t, snapshot)
	var hasPollCount bool
	for _, m := range snapshot {
		if m.ID == "PollCount" {
			hasPollCount = true
			assert.Equal(t, models.Counter, m.MType)
		}
	}
	assert.True(t, hasPollCount)
}

func TestAgent_Shutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host, err := url.Parse(server.URL)
	require.NoError(t, err)

	cfg := &config.Config{
		Client: config.ClientConfig{
			Address:        host.Host,
			PollInterval:   1,
			ReportInterval: 1,
			RateLimit:      1,
		},
	}

	agt, err := NewAgent(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	agt.Start(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	err = agt.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

func TestMetricsSender_SendAllMetricsBuffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/updates", r.URL.Path)
		require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		bodyReader, err := decodeGzipBody(r)
		require.NoError(t, err)
		var metrics []models.Metrics
		require.NoError(t, json.NewDecoder(bodyReader).Decode(&metrics))
		require.NotEmpty(t, metrics)
		require.Equal(t, "TestGauge", metrics[0].ID)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	host, err := url.Parse(server.URL)
	require.NoError(t, err)
	cfg := &config.Config{Client: config.ClientConfig{Address: host.Host}}
	sender, err := NewMetricsSender(cfg)
	require.NoError(t, err)
	value := 1.23
	metrics := []models.Metrics{
		{ID: "TestGauge", MType: models.Gauge, Value: &value},
	}
	err = sender.SendAllMetricsBuffer(t.Context(), metrics)
	require.NoError(t, err)
}
