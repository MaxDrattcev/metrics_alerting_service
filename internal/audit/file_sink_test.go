package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileSink_NotifyAndClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	sink, err := NewFileSink(path)
	require.NoError(t, err)

	err = sink.Notify(Event{
		TS:        123,
		Metrics:   []string{"Alloc", "PollCount"},
		IPAddress: "192.168.0.107",
	})
	require.NoError(t, err)
	require.NoError(t, sink.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), `"Alloc"`)
	require.Contains(t, string(data), `"192.168.0.107"`)
}

func TestFileSink_NotifyAfterClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	sink, err := NewFileSink(path)
	require.NoError(t, err)
	require.NoError(t, sink.Close())

	err = sink.Notify(Event{TS: 1})
	require.Error(t, err)
}
