package middleware

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
)

func TrustedSubnet(trustedSubnet string) gin.HandlerFunc {
	if trustedSubnet == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	_, network, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return func(c *gin.Context) {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
	return func(c *gin.Context) {
		ipStr := strings.TrimSpace(c.GetHeader("X-Real-IP"))
		if ipStr == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		ip := net.ParseIP(ipStr)
		if ip == nil || !network.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
