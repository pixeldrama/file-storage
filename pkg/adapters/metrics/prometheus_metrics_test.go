package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics_RecordUploadDuration(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := &PrometheusMetrics{
		uploadDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "file_upload_duration_seconds",
				Help:    "Duration of file uploads in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
	}
	registry.MustRegister(metrics.uploadDuration)

	metrics.RecordUploadDuration("success", 2*time.Second)

	if count := testutil.CollectAndCount(metrics.uploadDuration); count != 1 {
		t.Errorf("Expected 1 observation, got %d", count)
	}
}

func TestPrometheusMetrics_RecordUploadSize(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := &PrometheusMetrics{
		uploadSize: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "file_upload_size_bytes",
				Help:    "Size of uploaded files in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
			},
		),
	}
	registry.MustRegister(metrics.uploadSize)

	metrics.RecordUploadSize(2048)

	if count := testutil.CollectAndCount(metrics.uploadSize); count != 1 {
		t.Errorf("Expected 1 observation, got %d", count)
	}
}

func TestNewPrometheusMetrics(t *testing.T) {
	metrics := NewPrometheusMetrics()
	if metrics == nil {
		t.Error("Expected non-nil metrics instance")
	}
	if metrics.uploadDuration == nil {
		t.Error("Expected non-nil uploadDuration metric")
	}
	if metrics.uploadSize == nil {
		t.Error("Expected non-nil uploadSize metric")
	}
}
