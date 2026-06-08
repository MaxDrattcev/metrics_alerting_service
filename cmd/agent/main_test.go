package main

import (
	"os"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_FlagsOverrideJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte(`{
		"client": {"address": "json:8080", "poll_interval": 1, "report_interval": 1}
	}`), 0644))

	cfg, err := initConfig(environmentvar.EnvVar{}, AgentFlags{
		Config:  path,
		Address: "flag:9090",
	})
	require.NoError(t, err)
	assert.Equal(t, "flag:9090", cfg.Client.Address)
}

func TestInitConfig_EnvHighestPriority(t *testing.T) {
	cfg, err := initConfig(environmentvar.EnvVar{Address: "env:7070"}, AgentFlags{
		Address: "flag:9090",
	})
	require.NoError(t, err)
	assert.Equal(t, "env:7070", cfg.Client.Address)
}
