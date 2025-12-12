package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RecommendHomeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recommendation_service_home_total",
			Help: "Total number of /recommend/home responses",
		},
		[]string{"mode"}, // profile | random | empty
	)

	RecommendHomeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "recommendation_service_home_duration_seconds",
			Help:    "/recommend/home handler duration seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"mode"},
	)
)
