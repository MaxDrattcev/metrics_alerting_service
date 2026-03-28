package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHashSHA256(bodyBytes []byte, key string) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	_, err := mac.Write(bodyBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}
