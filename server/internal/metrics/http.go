package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var HTTPRequestDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests in seconds",
		Buckets: []float64{
			0.001,
			0.003,
			0.005,
			0.01,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
		},
	},
	[]string{"method", "path", "status"},
)

var RedirectTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redirect_total",
		Help: "Total number of redirect requests",
	},
	[]string{"status"},
)
