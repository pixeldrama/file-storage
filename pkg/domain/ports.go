package domain

import (
	"context"
	"io"
	"time"
)

// File represents a file in the system
type File struct {
	ID        string
	Size      int64
	CreatedAt time.Time
}

// UploadJob represents an upload job in the system
type UploadJob struct {
	ID        string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	FileID    string
	Error     string
}

// FileStorage defines the port for file storage operations
type FileStorage interface {
	Upload(ctx context.Context, fileID string, reader io.Reader) error
	Download(ctx context.Context, fileID string) (io.ReadCloser, error)
	Delete(ctx context.Context, fileID string) error
}

// UploadJobRepository defines the port for upload job persistence
type UploadJobRepository interface {
	Create(ctx context.Context, job *UploadJob) error
	Get(ctx context.Context, jobID string) (*UploadJob, error)
	Update(ctx context.Context, job *UploadJob) error
}

// MetricsCollector defines the port for metrics collection
type MetricsCollector interface {
	RecordUploadDuration(status string, duration time.Duration)
	RecordUploadSize(size int64)
}
