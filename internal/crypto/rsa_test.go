package crypto

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	pubPath := filepath.Join("..", "..", "keys", "agent_public.pem")
	privPath := filepath.Join("..", "..", "keys", "server_private.pem")

	pub, err := LoadPublicKey(pubPath)
	require.NoError(t, err)
	priv, err := LoadPrivateKey(privPath)
	require.NoError(t, err)

	data := []byte("metrics batch payload for test")
	encrypted, err := Encrypt(pub, data)
	require.NoError(t, err)

	decrypted, err := Decrypt(priv, encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestEncrypt_NilKey(t *testing.T) {
	_, err := Encrypt(nil, []byte("x"))
	require.Error(t, err)
}

func TestDecrypt_NilKey(t *testing.T) {
	_, err := Decrypt(nil, []byte("x"))
	require.Error(t, err)
}

func TestDecrypt_InvalidLength(t *testing.T) {
	privPath := filepath.Join("..", "..", "keys", "server_private.pem")
	priv, err := LoadPrivateKey(privPath)
	require.NoError(t, err)

	_, err = Decrypt(priv, []byte("short"))
	require.Error(t, err)
}

func TestLoadPublicKey_NotFound(t *testing.T) {
	_, err := LoadPublicKey("missing.pem")
	require.Error(t, err)
}
