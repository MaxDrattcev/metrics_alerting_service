package audit

import (
	"path/filepath"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewFromConfig_FileOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	pub := NewFromConfig(config.ServerConfig{
		AuditFile: path,
	})
	require.NotNil(t, pub)
	require.NoError(t, pub.Close())
}

func TestNewFromConfig_Empty(t *testing.T) {
	require.Nil(t, NewFromConfig(config.ServerConfig{}))
}

func TestNewFromConfig_HTTPOnly(t *testing.T) {
	pub := NewFromConfig(config.ServerConfig{
		AuditURL: "http://localhost:9999/audit",
	})
	require.NotNil(t, pub)
	require.NoError(t, pub.Close())
}
