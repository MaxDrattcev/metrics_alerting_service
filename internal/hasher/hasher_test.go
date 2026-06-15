package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeHashSHA256(t *testing.T) {
	hash1, err := ComputeHashSHA256([]byte("body"), "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	hash2, err := ComputeHashSHA256([]byte("body"), "secret")
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	hash3, err := ComputeHashSHA256([]byte("other"), "secret")
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}
