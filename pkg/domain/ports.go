package domain

import (
	"context"
	"io"
	"time"
)

type JobStatus string

const (
	JobStatusPending           JobStatus = "PENDING"
	JobStatusUploading         JobStatus = "UPLOADING"
	JobStatusVirusCheckPending JobStatus = "VIRUS_CHECK_PENDING"
	JobStatusVirusChecking     JobStatus = "VIRUS_CHECK_IN_PROGRESS"
	JobStatusCompleted         JobStatus = "COMPLETED"
	JobStatusFailed            JobStatus = "FAILED"
	JobStatusDeleted           JobStatus = "DELETED"
)

type File struct {
	ID        string
	Size      int64
	CreatedAt time.Time
}

type UploadJob struct {
	ID              string    `json:"jobId"`
	CreatedByUserId string    `json:"createdByUserId"`
	Filename        string    `json:"filename,omitempty"`
	Status          JobStatus `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	FileID          string    `json:"fileId,omitempty"`
	Error           string    `json:"error,omitempty"`
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
	GetByStatus(ctx context.Context, status JobStatus) ([]*UploadJob, error)
}

type MetricsCollector interface {
	RecordUploadDuration(status string, duration time.Duration)
	RecordUploadSize(size int64)
	RecordVirusCheckDuration(status string, duration time.Duration)
}

type VirusChecker interface {
	// CheckFile checks if a file contains a virus.
	// Returns true if the file is clean, false if it contains a virus.
	CheckFile(ctx context.Context, reader io.Reader) (bool, error)
}

type AuthorizationService interface {
	Authorize(ctx context.Context, userID string, resourceType string, resourceID string, action string) (bool, error)
}

type FileAuthorization interface {
	AuthorizeUploadFile(userID, fileType, linkedResourceType, linkedResourceID string) (bool, error)
	AuthorizeReadFile(userID, fileID string) (bool, error)
	AuthorizeDeleteFile(userID, fileID string) (bool, error)
}
