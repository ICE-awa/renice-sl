package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var MQPublishTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mq_publish_total",
		Help: "Total number of messages published to the JetStream",
	},
	[]string{"subject"},
)

var MQPublishDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "mq_publish_duration_seconds",
		Help: "Duration of publishing messages to the JetStream",
		Buckets: []float64{
			0.001,
			0.003,
			0.005,
			0.010,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
		},
	},
	[]string{"subject"},
)

var WorkerProcessDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "worker_process_duration_seconds",
		Help: "Duration of processing messages in workers",
		Buckets: []float64{
			0.001,
			0.003,
			0.005,
			0.010,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
		},
	},
	[]string{"subject"},
)
