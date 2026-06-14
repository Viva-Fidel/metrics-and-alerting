package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/gin-gonic/gin"
)

type hashResponseWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *hashResponseWriter) Write(data []byte) (int, error) {
	if _, err := w.body.Write(data); err != nil {
		return 0, err
	}
	return w.ResponseWriter.Write(data)
}

// HashMiddleware проверяет заголовок HashSHA256 входящих запросов и добавляет хеш в ответ
func HashMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key == "" {
			c.Next()
			return
		}

		if c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			if got := c.GetHeader(security.HashHeader); got != "" {
				if want := security.BuildHash(body, key); got != want {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}

		writer := &hashResponseWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		c.Header(security.HashHeader, security.BuildHash(writer.body.Bytes(), key))
	}
}
