package buildinfo

import (
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Print()

	require.NoError(t, w.Close())
	os.Stdout = old

	out, _ := io.ReadAll(r)
	assert.Contains(t, string(out), "Build version:")
}
