package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const clientIPKey ctxKey = "client_ip"

// ClientIP сохраняет IP клиента в context.Request для последующего аудита.
func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), clientIPKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// ClientIPFromContext возвращает IP клиента из context (пустая строка, если не задан).
func ClientIPFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(clientIPKey).(string)
	if !ok {
		return ""
	}
	return ip
}

// ContextWithClientIP сохраняет IP в context (для gRPC и аудита).
func ContextWithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}
