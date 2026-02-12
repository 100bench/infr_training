package relay

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	outboxEventsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_events_processed_total",
		Help: "Number of outbox events processed",
	})	

	outboxErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_errors_total",
		Help: "Number of outbox errors",
	},
	[]string{"error_type"},
)
	outboxLagSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_lag_seconds",
		Help: "Age of oldest unsent outbox event in seconds",
	})

	outboxBatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "outbox_batch_size",
		Help:    "Number of events per relay batch",
		Buckets: []float64{1, 5, 10, 25, 50, 100},
    })

)