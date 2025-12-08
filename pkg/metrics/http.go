package metrics

import (
	"apple_backend/pkg/httpx"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)
)

func normalizePath(path string) string {
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}
	path = strings.Replace(path, "/orders/", "/orders/{id}", 1)
	path = strings.Replace(path, "/payments/order/", "/payments/order/{id}", 1)
	return path
}

func HTTPMetricsMiddleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := normalizePath(r.URL.Path)
		method := r.Method

		sw := httpx.NewStatusWriter(w)
		start := time.Now()

		next.ServeHTTP(sw, r)

		statusCode := sw.Status
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		labels := prometheus.Labels{
			"service": service,
			"method":  method,
			"path":    path,
			"status":  http.StatusText(statusCode),
		}

		HTTPRequestsTotal.With(labels).Inc()
		HTTPRequestDuration.With(labels).Observe(time.Since(start).Seconds())
	})
}
