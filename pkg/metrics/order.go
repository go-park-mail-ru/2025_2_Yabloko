package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrdersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of created orders",
		},
		[]string{"store_id"},
	)

	OrdersCanceledTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_canceled_total",
			Help: "Total number of canceled orders",
		},
	)
)
