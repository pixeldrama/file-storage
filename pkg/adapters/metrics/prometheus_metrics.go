package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
	uploadDuration *prometheus.HistogramVec
	uploadSize     prometheus.Histogram
}

func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		uploadDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "file_upload_duration_seconds",
				Help:    "Duration of file uploads in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
		uploadSize: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "file_upload_size_bytes",
				Help:    "Size of uploaded files in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 1MB
			},
		),
	}

	prometheus.MustRegister(metrics.uploadDuration)
	prometheus.MustRegister(metrics.uploadSize)

	return metrics
}

func (m *PrometheusMetrics) RecordUploadDuration(status string, duration time.Duration) {
	m.uploadDuration.WithLabelValues(status).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) RecordUploadSize(size int64) {
	m.uploadSize.Observe(float64(size))
}
