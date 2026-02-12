package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			if c.GetHeader("Content-Type") != "text/plain" {
				c.String(http.StatusBadRequest, "Content-Type must be text/plain")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
