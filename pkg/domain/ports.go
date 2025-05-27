package domain

import (
	"context"
	"io"
	"time"
)

type JobStatus string

const (
	JobStatusUploading     JobStatus = "UPLOADING"
	JobStatusVirusChecking JobStatus = "VIRUS_CHECKING"
	JobStatusCompleted     JobStatus = "COMPLETED"
	JobStatusFailed        JobStatus = "FAILED"
	JobStatusDeleted       JobStatus = "DELETED"
)

type File struct {
	ID        string
	Size      int64
	CreatedAt time.Time
}

type UploadJob struct {
	ID        string    `json:"jobId"`
	Filename  string    `json:"filename,omitempty"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	FileID    string    `json:"fileId,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type FileStorage interface {
	Upload(ctx context.Context, fileID string, reader io.Reader) error
	Download(ctx context.Context, fileID string) (io.ReadCloser, error)
	Delete(ctx context.Context, fileID string) error
}

type UploadJobRepository interface {
	Create(ctx context.Context, job *UploadJob) error
	Get(ctx context.Context, jobID string) (*UploadJob, error)
	Update(ctx context.Context, job *UploadJob) error
	GetByFileID(ctx context.Context, fileID string) (*UploadJob, error)
}

type MetricsCollector interface {
	RecordUploadDuration(status string, duration time.Duration)
	RecordUploadSize(size int64)
}
