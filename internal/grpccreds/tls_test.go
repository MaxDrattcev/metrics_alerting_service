package grpccreds

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerCredentials_EmptyPaths(t *testing.T) {
	_, err := ServerCredentials("", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cert and key paths are required")
}

func TestClientCredentials_EmptyPath(t *testing.T) {
	_, err := ClientCredentials("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cert path is required")
}

func TestServerCredentials_InvalidFiles(t *testing.T) {
	_, err := ServerCredentials("no-such-cert.pem", "no-such-key.key")
	require.Error(t, err)
}

func TestClientCredentials_InvalidFile(t *testing.T) {
	_, err := ClientCredentials("no-such-cert.pem")
	require.Error(t, err)
}

func TestServerAndClientCredentials_WithRealCert(t *testing.T) {
	certFile, keyFile := WriteTestSelfSignedCert(t)

	serverCreds, err := ServerCredentials(certFile, keyFile)
	require.NoError(t, err)
	require.NotNil(t, serverCreds)

	clientCreds, err := ClientCredentials(certFile)
	require.NoError(t, err)
	require.NotNil(t, clientCreds)
}
