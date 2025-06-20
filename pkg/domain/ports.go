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

type FileInfo struct {
	ID                 string    `json:"id"`
	Filename           string    `json:"filename,omitempty"`
	FileType           string    `json:"fileType,omitempty"`
	LinkedResourceType string    `json:"linkedResourceType,omitempty"`
	LinkedResourceID   string    `json:"linkedResourceID,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type UploadJob struct {
	ID              string    `json:"jobId"`
	CreatedByUserId string    `json:"createdByUserId"`
	FileID          string    `json:"fileId,omitempty"`
	Status          JobStatus `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
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

type FileInfoRepository interface {
	Create(ctx context.Context, fileInfo *FileInfo) error
	Get(ctx context.Context, fileID string) (*FileInfo, error)
	Update(ctx context.Context, fileInfo *FileInfo) error
	Delete(ctx context.Context, fileID string) error
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
	CanUploadFile(userID, fileType, linkedResourceType, linkedResourceID string) (bool, error)
	CanReadFile(userID, fileID string) (bool, error)
	CanDeleteFile(userID, fileID string) (bool, error)

	CreateFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error
	RemoveFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error
}
