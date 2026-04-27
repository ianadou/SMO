package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// healthLivePath and healthReadyPath are excluded from request logging
// because they are hit roughly every 10 seconds by Docker's HEALTHCHECK
// and Dockhand. Logging those requests would drown the log stream in
// noise without providing value.
const (
	healthLivePath  = "/health/live"
	healthReadyPath = "/health/ready"
)

// SLogLogger returns a Gin middleware that logs each completed request
// as a single structured slog entry. The level is mapped from the
// response status: 5xx → ERROR, anything else → INFO.
//
// The middleware MUST be registered after RequestID so the request_id
// field can be looked up from the request context.
func SLogLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if path == healthLivePath || path == healthReadyPath {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		}

		logger.LogAttrs(
			c.Request.Context(),
			level,
			"request completed",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.String("request_id", RequestIDFromContext(c.Request.Context())),
			slog.String("remote_ip", c.ClientIP()),
		)
	}
}
