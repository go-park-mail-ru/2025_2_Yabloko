package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Общее количество созданных заказов по магазинам
	OrdersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_orders_created_total",
			Help: "Total number of successfully created orders per store",
		},
		[]string{"store_id"},
	)

	// Общее количество отмененных заказов по магазинам
	OrdersCanceledTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_orders_canceled_total",
			Help: "Total number of canceled orders per store",
		},
		[]string{"store_id"},
	)

	// Количество запросов GetOrder по магазинам
	OrdersGetTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_orders_get_total",
			Help: "Total number of GetOrder requests per store count",
		},
		[]string{"stores_count"},
	)

	// Длительность GetOrder запросов
	OrdersGetDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_service_orders_get_duration_seconds",
			Help:    "Order get request duration seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"stores_count"},
	)
)
