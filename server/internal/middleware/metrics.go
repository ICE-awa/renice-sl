package middleware

import (
	"github.com/ICE-awa/renice-sl/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"strconv"
	"time"
)

func HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "/metrics" {
			c.Next()
			return
		}

		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)
		start := time.Now()

		c.Next()

		if path == "" {
			path = "unknown"
		}

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		metrics.HTTPRequestDurationSeconds.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Observe(duration)

		if status[0] == '5' {
			slog.Error("http request failed",
				slog.String("request_id", requestID),
				slog.String("method", c.Request.Method),
				slog.String("path", path),
				slog.String("status", status),
				slog.Float64("duration", duration),
			)
		}
	}
}
