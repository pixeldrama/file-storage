package http

import (
	"time"

	"file-storage-go/pkg/domain"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusUploading JobStatus = "UPLOADING"
	JobStatusChecking  JobStatus = "VIRUS_CHECKING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
)

type UploadJob struct {
	JobID     string    `json:"jobId"`
	Filename  string    `json:"filename,omitempty"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	FileID    string    `json:"fileId,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func toAPIJobStatus(domainStatus domain.JobStatus) JobStatus {
	switch domainStatus {
	case domain.JobStatusPending:
		return JobStatusPending
	case domain.JobStatusUploading:
		return JobStatusUploading
	case domain.JobStatusVirusCheckPending, domain.JobStatusVirusChecking:
		return JobStatusChecking
	case domain.JobStatusCompleted:
		return JobStatusCompleted
	case domain.JobStatusFailed:
		return JobStatusFailed
	default:
		return JobStatusFailed
	}
}

func ToAPIJob(job *domain.UploadJob) *UploadJob {
	if job == nil {
		return nil
	}

	return &UploadJob{
		JobID:     job.ID,
		Filename:  job.Filename,
		Status:    toAPIJobStatus(job.Status),
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
		FileID:    job.FileID,
		Error:     job.Error,
	}
}
