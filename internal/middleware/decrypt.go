package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/crypto"
	"github.com/gin-gonic/gin"
)

// Decrypt расшифровывает тело POST-запроса, если задан путь к приватному ключу.
// Должен стоять ДО middleware.Compress().
func Decrypt(privateKeyPath string) gin.HandlerFunc {
	if privateKeyPath == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	privateKey, err := crypto.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost || c.Request.Body == nil {
			c.Next()
			return
		}

		encrypted, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(encrypted) == 0 {
			c.Next()
			return
		}

		decrypted, err := crypto.Decrypt(privateKey, encrypted)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(decrypted))
		c.Request.ContentLength = int64(len(decrypted))
		c.Next()
	}
}
