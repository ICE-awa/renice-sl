package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var RedisCacheHitTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_cache_hit_total",
		Help: "Total number of Redis cache hits",
	},
	[]string{"operation"},
)

var RedisCacheMissTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_cache_miss_total",
		Help: "Total number of Redis cache misses",
	},
	[]string{"operation"},
)
