package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheHitsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Number of cache hits for L1 cache",
	},
	[]string{"cache_level"},)

	cacheMissesTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Number of cache misses for L1 cache",
	})

	cacheL2ErrorsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_l2_errors_total",
		Help: "Number of errors when accessing L2 cache",
	})

)