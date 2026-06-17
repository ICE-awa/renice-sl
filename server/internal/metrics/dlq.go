package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DLQMessagesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dlq_messages_total",
		Help: "Total number of messages in the dead-letter queue",
	},
	[]string{"subject"},
)

var DLQRetryTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dlq_retry_total",
		Help: "Total number of retries for messages in the dead-letter queue",
	},
	[]string{"subject"},
)

var DLQResolvedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dlq_resolved_total",
		Help: "Total number of resolved messages in the dead-letter queue",
	},
	[]string{"subject"},
)
