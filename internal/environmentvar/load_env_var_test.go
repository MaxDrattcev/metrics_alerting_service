package environmentvar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvVar_Success(t *testing.T) {
	t.Setenv("ADDRESS", "localhost:8080")
	t.Setenv("GRPC_ADDRESS", "localhost:8081")
	t.Setenv("TRUSTED_SUBNET", "192.168.0.0/24")
	t.Setenv("REPORT_INTERVAL", "10")
	t.Setenv("POLL_INTERVAL", "2")
	t.Setenv("RATE_LIMIT", "5")

	env, err := LoadEnvVar()
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", env.Address)
	assert.Equal(t, "localhost:8081", env.GRPCAddress)
	assert.Equal(t, "192.168.0.0/24", env.TrustedSubnet)
	assert.Equal(t, int64(10), env.ReportInterval)
	assert.Equal(t, int64(2), env.PollInterval)
	assert.Equal(t, 5, env.RateLimit)
}
