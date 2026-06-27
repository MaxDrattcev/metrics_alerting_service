package flags

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckUnknownFlags_KnownFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String("a", "", "address")
	os.Args = []string{"test", "-a", "localhost:8080"}

	flag.CommandLine = fs
	require.NoError(t, CheckUnknownFlags())
}

func TestCheckUnknownFlags_UnknownFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String("a", "", "address")
	os.Args = []string{"test", "-unknown"}

	flag.CommandLine = fs
	require.Error(t, CheckUnknownFlags())
}
