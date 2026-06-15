package main

import (
	"os"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_FromJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte(`{
		"server": {
			"address": "json:8080",
			"store_interval": 300,
			"store_file": "metrics.json",
			"restore": true,
			"database_dsn": "postgres://localhost/db",
			"key": "json-key",
			"audit_file": "/tmp/audit.log",
			"audit_url": "http://audit.local/log",
			"crypto_key": "keys/server_private.pem"
		}
	}`), 0644))

	cfg, err := initConfig(environmentvar.EnvVar{}, ServerFlags{Config: path})
	require.NoError(t, err)

	assert.Equal(t, "json:8080", cfg.Server.Address)
	require.NotNil(t, cfg.Server.StoreInterval)
	assert.Equal(t, int64(300), *cfg.Server.StoreInterval)
	assert.Equal(t, "metrics.json", cfg.Server.FileStoragePath)
	require.NotNil(t, cfg.Server.Restore)
	assert.True(t, *cfg.Server.Restore)
	assert.Equal(t, "postgres://localhost/db", cfg.Server.DatabaseDSN)
	assert.Equal(t, "json-key", cfg.Server.Key)
	assert.Equal(t, "/tmp/audit.log", cfg.Server.AuditFile)
	assert.Equal(t, "http://audit.local/log", cfg.Server.AuditURL)
	assert.Equal(t, "keys/server_private.pem", cfg.Server.CryptoKey)
}

func TestInitConfig_FlagsOverrideJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte(`{
		"server": {"address": "json:8080", "store_interval": 300, "restore": false}
	}`), 0644))

	restore := true
	cfg, err := initConfig(environmentvar.EnvVar{}, ServerFlags{
		Config:          path,
		Address:         "flag:9090",
		StoreInterval:   60,
		FileStoragePath: "/tmp/flag.json",
		Restore:         restore,
		DatabaseDSN:     "postgres://flag/db",
		Key:             "flag-key",
		AuditFile:       "/tmp/flag-audit.log",
		AuditURL:        "http://flag.audit/log",
		CryptoKey:       "keys/flag.pem",
	})
	require.NoError(t, err)

	assert.Equal(t, "flag:9090", cfg.Server.Address)
	require.NotNil(t, cfg.Server.StoreInterval)
	assert.Equal(t, int64(60), *cfg.Server.StoreInterval)
	assert.Equal(t, "/tmp/flag.json", cfg.Server.FileStoragePath)
	require.NotNil(t, cfg.Server.Restore)
	assert.True(t, *cfg.Server.Restore)
	assert.Equal(t, "postgres://flag/db", cfg.Server.DatabaseDSN)
	assert.Equal(t, "flag-key", cfg.Server.Key)
	assert.Equal(t, "/tmp/flag-audit.log", cfg.Server.AuditFile)
	assert.Equal(t, "http://flag.audit/log", cfg.Server.AuditURL)
	assert.Equal(t, "keys/flag.pem", cfg.Server.CryptoKey)
}

func TestInitConfig_EnvHighestPriority(t *testing.T) {
	storeInterval := int64(120)
	restore := false

	cfg, err := initConfig(environmentvar.EnvVar{
		Address:         "env:7070",
		StoreInterval:   &storeInterval,
		FileStoragePath: "/tmp/env.json",
		Restore:         &restore,
		DatabaseDSN:     "postgres://env/db",
		Key:             "env-key",
		AuditFile:       "/tmp/env-audit.log",
		AuditURL:        "http://env.audit/log",
		CryptoKeyServer: "keys/env.pem",
	}, ServerFlags{
		Address:         "flag:9090",
		StoreInterval:   60,
		FileStoragePath: "/tmp/flag.json",
		Restore:         true,
		DatabaseDSN:     "postgres://flag/db",
		Key:             "flag-key",
		AuditFile:       "/tmp/flag-audit.log",
		AuditURL:        "http://flag.audit/log",
		CryptoKey:       "keys/flag.pem",
	})
	require.NoError(t, err)

	assert.Equal(t, "env:7070", cfg.Server.Address)
	require.NotNil(t, cfg.Server.StoreInterval)
	assert.Equal(t, int64(120), *cfg.Server.StoreInterval)
	assert.Equal(t, "/tmp/env.json", cfg.Server.FileStoragePath)
	require.NotNil(t, cfg.Server.Restore)
	assert.False(t, *cfg.Server.Restore)
	assert.Equal(t, "postgres://env/db", cfg.Server.DatabaseDSN)
	assert.Equal(t, "env-key", cfg.Server.Key)
	assert.Equal(t, "/tmp/env-audit.log", cfg.Server.AuditFile)
	assert.Equal(t, "http://env.audit/log", cfg.Server.AuditURL)
	assert.Equal(t, "keys/env.pem", cfg.Server.CryptoKey)
}

func TestInitConfig_ConfigServerEnvOverridesFlag(t *testing.T) {
	dir := t.TempDir()
	envPath := dir + "/env.json"
	flagPath := dir + "/flag.json"

	require.NoError(t, os.WriteFile(envPath, []byte(`{
		"server": {"address": "env-json:8080"}
	}`), 0644))
	require.NoError(t, os.WriteFile(flagPath, []byte(`{
		"server": {"address": "flag-json:8080"}
	}`), 0644))

	cfg, err := initConfig(environmentvar.EnvVar{
		ConfigServer: envPath,
	}, ServerFlags{
		Config: flagPath,
	})
	require.NoError(t, err)
	assert.Equal(t, "env-json:8080", cfg.Server.Address)
}

func TestInitConfig_FlagsOnlyWithoutJSON(t *testing.T) {
	cfg, err := initConfig(environmentvar.EnvVar{}, ServerFlags{
		Address:         "localhost:8080",
		FileStoragePath: "metrics.json",
	})
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", cfg.Server.Address)
	assert.Equal(t, "metrics.json", cfg.Server.FileStoragePath)
	assert.Nil(t, cfg.Server.StoreInterval)
	assert.Nil(t, cfg.Server.Restore)
}

func TestInitConfig_StoreIntervalZeroNotApplied(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte(`{
		"server": {"store_interval": 300}
	}`), 0644))

	cfg, err := initConfig(environmentvar.EnvVar{}, ServerFlags{
		Config:        path,
		StoreInterval: 0,
	})
	require.NoError(t, err)
	require.NotNil(t, cfg.Server.StoreInterval)
	assert.Equal(t, int64(300), *cfg.Server.StoreInterval)
}

func TestInitConfig_RestoreFalseNotAppliedFromFlag(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte(`{
		"server": {"restore": true}
	}`), 0644))

	// flags.Restore == false (дефолт) — флаг не перезаписывает JSON
	cfg, err := initConfig(environmentvar.EnvVar{}, ServerFlags{
		Config:  path,
		Restore: false,
	})
	require.NoError(t, err)
	require.NotNil(t, cfg.Server.Restore)
	assert.True(t, *cfg.Server.Restore)
}

func TestInitConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.json"
	require.NoError(t, os.WriteFile(path, []byte("{invalid"), 0644))

	_, err := initConfig(environmentvar.EnvVar{}, ServerFlags{Config: path})
	require.Error(t, err)
}
