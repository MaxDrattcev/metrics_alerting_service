package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const clientIPKey ctxKey = "client_ip"

func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), clientIPKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey).(string)
	return ip
}
