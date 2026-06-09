// Package logging предоставляет middleware для структурированного логирования HTTP-запросов.
package logging

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware логирует метод, URI, статус и время обработки каждого запроса.
func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("uri", path),
			slog.Duration("latency", time.Since(start)),
			slog.Int("status", c.Writer.Status()),
			slog.Int("body_size", c.Writer.Size()),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error", slog.String("error", err.Error()))
			}
		}
	}
}
