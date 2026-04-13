package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestRate     *prometheus.CounterVec
	RequestErrors   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		RequestRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "chat_service_requests_total",
				Help: "Количество запросов",
			},
			[]string{"method", "endpoint"},
		),

		RequestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "chat_service_errors_total",
				Help: "Количество ошибок",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "chat_service_request_duration_seconds",
				Help:    "Время выполнения",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
	}
}

func (m *Metrics) RecordRequest(method, endpoint string, duration time.Duration, err error) {
	m.RequestRate.WithLabelValues(method, endpoint).Inc()
	m.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if err != nil {
		m.RequestErrors.WithLabelValues(method, endpoint, err.Error()).Inc()
	}
}
