package middleware

import (
	"github.com/ICE-awa/renice-sl/internal/metrics"
	"github.com/gin-gonic/gin"
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
	}
}
