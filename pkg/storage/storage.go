package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	uploadDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_upload_duration_seconds",
			Help:    "Duration of file uploads in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
	uploadSize = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "file_upload_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 1MB
		},
	)
)

func init() {
	prometheus.MustRegister(uploadDuration)
	prometheus.MustRegister(uploadSize)
}

type Storage struct {
	client        *azblob.Client
	containerName string
}

func NewStorage(accountURL, accountKey, containerName string) (*Storage, error) {
	credential, err := azblob.NewSharedKeyCredential(accountURL, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(accountURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Storage{
		client:        client,
		containerName: containerName,
	}, nil
}

func (s *Storage) UploadFile(ctx context.Context, fileID string, reader io.Reader) error {
	start := time.Now()
	blobName := fmt.Sprintf("%s", fileID)

	// Upload the file
	_, err := s.client.UploadStream(ctx, s.containerName, blobName, reader, nil)
	if err != nil {
		uploadDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
		return fmt.Errorf("failed to upload file: %w", err)
	}

	uploadDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())
	return nil
}

func (s *Storage) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	blobName := fmt.Sprintf("%s", fileID)

	// Download the file
	downloadResponse, err := s.client.DownloadStream(ctx, s.containerName, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return downloadResponse.Body, nil
}

func (s *Storage) DeleteFile(ctx context.Context, fileID string) error {
	blobName := fmt.Sprintf("%s", fileID)

	_, err := s.client.DeleteBlob(ctx, s.containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
