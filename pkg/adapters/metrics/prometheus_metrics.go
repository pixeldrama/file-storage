package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
	uploadDuration     *prometheus.HistogramVec
	uploadSize         prometheus.Histogram
	virusCheckDuration *prometheus.HistogramVec
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
				Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
			},
		),
		virusCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "virus_check_duration_seconds",
				Help:    "Duration of virus checks in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
	}

	prometheus.MustRegister(metrics.uploadDuration)
	prometheus.MustRegister(metrics.uploadSize)
	prometheus.MustRegister(metrics.virusCheckDuration)

	return metrics
}

func (m *PrometheusMetrics) RecordUploadDuration(status string, duration time.Duration) {
	m.uploadDuration.WithLabelValues(status).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) RecordUploadSize(size int64) {
	m.uploadSize.Observe(float64(size))
}

func (m *PrometheusMetrics) RecordVirusCheckDuration(status string, duration time.Duration) {
	m.virusCheckDuration.WithLabelValues(status).Observe(duration.Seconds())
}
