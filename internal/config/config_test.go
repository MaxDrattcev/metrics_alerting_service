package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientConfig_GetPollInterval(t *testing.T) {
	tests := []struct {
		name         string
		pollInterval int64
		want         time.Duration
	}{
		{
			name:         "2 seconds",
			pollInterval: 2,
			want:         2 * time.Second,
		},
		{
			name:         "10 seconds",
			pollInterval: 10,
			want:         10 * time.Second,
		},
		{
			name:         "zero",
			pollInterval: 0,
			want:         0 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{
				PollInterval: tt.pollInterval,
			}

			interval := cfg.GetPollInterval()
			assert.Equal(t, tt.want, interval)
		})
	}
}

func TestClientConfig_GetReportInterval(t *testing.T) {
	tests := []struct {
		name           string
		reportInterval int64
		want           time.Duration
	}{
		{
			name:           "10 seconds",
			reportInterval: 10,
			want:           10 * time.Second,
		},
		{
			name:           "30 seconds",
			reportInterval: 30,
			want:           30 * time.Second,
		},
		{
			name:           "zero",
			reportInterval: 0,
			want:           0 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{
				ReportInterval: tt.reportInterval,
			}

			interval := cfg.GetReportInterval()
			assert.Equal(t, tt.want, interval)
		})
	}
}

func TestServerConfig_GetStoreInterval(t *testing.T) {
	interval := int64(300)
	cfg := &ServerConfig{StoreInterval: &interval}
	assert.Equal(t, 300*time.Second, cfg.GetStoreInterval())
}

func TestLoadConfigJSON_Success(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	content := `{
		"server": {"address": "localhost:9090", "store_interval": 10, "store_file": "m.json", "restore": true},
		"client": {"address": "localhost:9090", "poll_interval": 2, "report_interval": 5}
	}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	cfg, err := LoadConfigJSON(path)
	require.NoError(t, err)
	assert.Equal(t, "localhost:9090", cfg.Server.Address)
	assert.Equal(t, int64(10), *cfg.Server.StoreInterval)
	assert.True(t, *cfg.Server.Restore)
	assert.Equal(t, int64(2), cfg.Client.PollInterval)
}

func TestLoadConfigJSON_FileNotFound(t *testing.T) {
	_, err := LoadConfigJSON("no-such-file.json")
	require.Error(t, err)
}

func TestLoadConfigJSON_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.json"
	require.NoError(t, os.WriteFile(path, []byte("{invalid"), 0644))

	_, err := LoadConfigJSON(path)
	require.Error(t, err)
}
