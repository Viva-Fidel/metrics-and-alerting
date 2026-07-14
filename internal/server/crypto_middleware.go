package server

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/gin-gonic/gin"
)

// CryptoMiddleware расшифровывает тело входящего запроса с помощью приватного RSA-ключа
func CryptoMiddleware(privateKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if len(body) == 0 {
			c.Next()
			return
		}

		decrypted, err := security.Decrypt(body, privateKey)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(decrypted))
		c.Next()
	}
}
