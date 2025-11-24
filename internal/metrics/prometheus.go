package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "The total number of processed requests",
		},
		[]string{"method", "path", "status", "upstream"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "The duration of the requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "upstream"},
	)

	UpstreamErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_upstream_errors_total",
			Help: "The total number of upstream errors",
		},
		[]string{"upstream", "error_type"},
	)

	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_hits_total",
			Help: "The total number of rate limit hits",
		},
		[]string{"route", "ip"},
	)
)
