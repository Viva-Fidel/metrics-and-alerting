package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/gin-gonic/gin"
)

// TrustedSubnetMiddleware отклоняет запросы, чей X-Real-IP не входит в доверенную подсеть.
// При пустом cidr проверки не выполняются.
func TrustedSubnetMiddleware(cidr string) (gin.HandlerFunc, error) {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return func(c *gin.Context) {
			c.Next()
		}, nil
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("parse trusted subnet %q: %w", cidr, err)
	}

	return func(c *gin.Context) {
		ip := net.ParseIP(c.GetHeader(security.RealIPHeader))
		if ip == nil || !network.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}, nil
}
