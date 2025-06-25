package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"file-storage-go/pkg/domain"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type AzureBlobStorage struct {
	client        *azblob.Client
	containerName string
	metrics       domain.MetricsCollector
}

func NewAzureBlobStorage(accountName, serviceURL, accountKey, containerName string, metrics domain.MetricsCollector) (*AzureBlobStorage, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &AzureBlobStorage{
		client:        client,
		containerName: containerName,
		metrics:       metrics,
	}, nil
}

func (s *AzureBlobStorage) getBlobName(fileID string) string {
	return fileID
}

func (s *AzureBlobStorage) Upload(ctx context.Context, fileID string, reader io.Reader) error {
	start := time.Now()
	blobName := s.getBlobName(fileID)

	_, err := s.client.UploadStream(ctx, s.containerName, blobName, reader, nil)
	if err != nil {
		s.metrics.RecordUploadDuration("error", time.Since(start))
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.metrics.RecordUploadDuration("success", time.Since(start))
	return nil
}

func (s *AzureBlobStorage) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	blobName := s.getBlobName(fileID)

	downloadResponse, err := s.client.DownloadStream(ctx, s.containerName, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return downloadResponse.Body, nil
}

func (s *AzureBlobStorage) Delete(ctx context.Context, fileID string) error {
	blobName := s.getBlobName(fileID)

	_, err := s.client.DeleteBlob(ctx, s.containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
