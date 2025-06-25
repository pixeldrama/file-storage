package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

// GinLoggerMiddleware returns a Gin middleware that logs requests using the provided slog.Logger
func GinLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Skip logging for certain paths (like health checks)
		path := c.Request.URL.Path
		if path == "/health" || path == "/metrics" {
			return
		}

		// Get error if exists
		var errs []string
		for _, err := range c.Errors {
			errs = append(errs, err.Error())
		}

		// Log the request details in ECS format
		attrs := []slog.Attr{
			slog.String("http.request.method", c.Request.Method),
			slog.String("http.request.path", path),
			slog.Int("http.response.status_code", c.Writer.Status()),
			slog.String("client.ip", c.ClientIP()),
			slog.Int("http.response.body.bytes", c.Writer.Size()),
		}

		if len(errs) > 0 {
			attrs = append(attrs, slog.String("error.message", strings.Join(errs, "; ")))
		}

		if c.Writer.Status() >= 400 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "HTTP request failed", attrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP request completed", attrs...)
		}
	}
}
