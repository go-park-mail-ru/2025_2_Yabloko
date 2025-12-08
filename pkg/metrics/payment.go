package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PaymentsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_created_total",
			Help: "Total number of created payments",
		},
		[]string{"status"},
	)

	PaymentsSucceededTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_succeeded_total",
			Help: "Total number of successfully completed payments",
		},
	)

	PaymentsFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_failed_total",
			Help: "Total number of failed payments by status",
		},
		[]string{"status"},
	)
)
