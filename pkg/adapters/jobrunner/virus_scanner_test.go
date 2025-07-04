package jobrunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"file-storage-go/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFileStorage struct {
	downloadFunc func(ctx context.Context, fileID string) (io.ReadCloser, error)
}

func (m *mockFileStorage) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	return m.downloadFunc(ctx, fileID)
}

func (m *mockFileStorage) Upload(ctx context.Context, fileID string, reader io.Reader) error {
	return nil
}

func (m *mockFileStorage) Delete(ctx context.Context, fileID string) error {
	return nil
}

type mockVirusChecker struct {
	checkFunc func(ctx context.Context, reader io.Reader) (bool, error)
}

func (m *mockVirusChecker) CheckFile(ctx context.Context, reader io.Reader) (bool, error) {
	return m.checkFunc(ctx, reader)
}

type mockJobRepository struct {
	jobs map[string]*domain.UploadJob
}

type mockFileInfoRepository struct {
	fileInfos map[string]*domain.FileInfo
}

func newMockFileInfoRepository() *mockFileInfoRepository {
	return &mockFileInfoRepository{
		fileInfos: make(map[string]*domain.FileInfo),
	}
}

func (m *mockFileInfoRepository) Create(ctx context.Context, fileInfo *domain.FileInfo) error {
	m.fileInfos[fileInfo.ID] = fileInfo
	return nil
}

func (m *mockFileInfoRepository) Get(ctx context.Context, fileID string) (*domain.FileInfo, error) {
	if fileInfo, exists := m.fileInfos[fileID]; exists {
		return fileInfo, nil
	}
	return nil, fmt.Errorf("file info not found")
}

func (m *mockFileInfoRepository) Update(ctx context.Context, fileInfo *domain.FileInfo) error {
	m.fileInfos[fileInfo.ID] = fileInfo
	return nil
}

func (m *mockFileInfoRepository) Delete(ctx context.Context, fileID string) error {
	delete(m.fileInfos, fileID)
	return nil
}

type mockFileAuthorization struct{}

func (m *mockFileAuthorization) CanUploadFile(userID, fileType, linkedResourceType, linkedResourceID string) (bool, error) {
	return true, nil
}

func (m *mockFileAuthorization) CanReadFile(userID, fileID string) (bool, error) {
	return true, nil
}

func (m *mockFileAuthorization) CanDeleteFile(userID, fileID string) (bool, error) {
	return true, nil
}

func (m *mockFileAuthorization) CreateFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error {
	return nil
}

func (m *mockFileAuthorization) RemoveFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error {
	return nil
}

func newMockJobRepository() *mockJobRepository {
	return &mockJobRepository{
		jobs: make(map[string]*domain.UploadJob),
	}
}

func (m *mockJobRepository) Create(ctx context.Context, job *domain.UploadJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepository) Get(ctx context.Context, jobID string) (*domain.UploadJob, error) {
	return m.jobs[jobID], nil
}

func (m *mockJobRepository) Update(ctx context.Context, job *domain.UploadJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepository) GetByFileID(ctx context.Context, fileID string) (*domain.UploadJob, error) {
	for _, job := range m.jobs {
		if job.FileID == fileID {
			return job, nil
		}
	}
	return nil, nil
}

func (m *mockJobRepository) GetByStatus(ctx context.Context, status domain.JobStatus) ([]*domain.UploadJob, error) {
	var jobs []*domain.UploadJob
	for _, job := range m.jobs {
		if job.Status == status {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

type mockMetrics struct{}

func (m *mockMetrics) RecordUploadDuration(status string, duration time.Duration)     {}
func (m *mockMetrics) RecordUploadSize(size int64)                                    {}
func (m *mockMetrics) RecordVirusCheckDuration(status string, duration time.Duration) {}

func TestVirusScannerJobRunner_ProcessJob(t *testing.T) {
	tests := []struct {
		name           string
		job            *domain.UploadJob
		downloadErr    error
		checkResult    bool
		checkErr       error
		expectedStatus domain.JobStatus
		expectedError  string
	}{
		{
			name: "successful virus check",
			job: &domain.UploadJob{
				ID:              "test-job",
				CreatedByUserId: "test-user",
				FileID:          "test-file",
				Status:          domain.JobStatusVirusCheckPending,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			checkResult:    true,
			expectedStatus: domain.JobStatusCompleted,
		},
		{
			name: "virus check failed - malware detected",
			job: &domain.UploadJob{
				ID:              "test-job",
				CreatedByUserId: "test-user",
				FileID:          "test-file",
				Status:          domain.JobStatusVirusCheckPending,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			checkResult:    false,
			expectedStatus: domain.JobStatusFailed,
			expectedError:  "file contains malware",
		},
		{
			name: "download error",
			job: &domain.UploadJob{
				ID:              "test-job",
				CreatedByUserId: "test-user",
				FileID:          "test-file",
				Status:          domain.JobStatusVirusCheckPending,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			downloadErr:    errors.New("download failed"),
			expectedStatus: domain.JobStatusFailed,
			expectedError:  "failed to download file: download failed",
		},
		{
			name: "virus check error",
			job: &domain.UploadJob{
				ID:              "test-job",
				CreatedByUserId: "test-user",
				FileID:          "test-file",
				Status:          domain.JobStatusVirusCheckPending,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			checkErr:       errors.New("check failed"),
			expectedStatus: domain.JobStatusFailed,
			expectedError:  "virus check failed: check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockJobRepository()
			require.NoError(t, repo.Create(context.Background(), tt.job))

			fileInfoRepo := newMockFileInfoRepository()
			fileInfo := &domain.FileInfo{
				ID:                 tt.job.FileID,
				Filename:           "test-file.txt",
				FileType:           "text/plain",
				LinkedResourceType: "test",
				LinkedResourceID:   "test-id",
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}
			require.NoError(t, fileInfoRepo.Create(context.Background(), fileInfo))

			fileStorage := &mockFileStorage{
				downloadFunc: func(ctx context.Context, fileID string) (io.ReadCloser, error) {
					if tt.downloadErr != nil {
						return nil, tt.downloadErr
					}
					return io.NopCloser(io.Reader(nil)), nil
				},
			}

			virusChecker := &mockVirusChecker{
				checkFunc: func(ctx context.Context, reader io.Reader) (bool, error) {
					if tt.checkErr != nil {
						return false, tt.checkErr
					}
					return tt.checkResult, nil
				},
			}

			metrics := &mockMetrics{}

			fileAuthorization := &mockFileAuthorization{}

			runner := NewVirusScannerJobRunner(
				repo,
				fileInfoRepo,
				fileAuthorization,
				fileStorage,
				virusChecker,
				5*time.Second,
				metrics,
			)

			err := runner.processJob(context.Background(), tt.job)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			updatedJob, err := repo.Get(context.Background(), tt.job.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, updatedJob.Status)
			if tt.expectedError != "" {
				assert.Contains(t, updatedJob.Error, tt.expectedError)
			}
		})
	}
}

func TestVirusScannerJobRunner_ProcessStuckJobs(t *testing.T) {
	now := time.Now()
	stuckJob := &domain.UploadJob{
		ID:              "stuck-job",
		CreatedByUserId: "test-user",
		FileID:          "stuck-file",
		Status:          domain.JobStatusVirusChecking,
		CreatedAt:       now.Add(-10 * time.Minute),
		UpdatedAt:       now.Add(-6 * time.Second),
	}

	pendingJob := &domain.UploadJob{
		ID:              "pending-job",
		CreatedByUserId: "test-user",
		FileID:          "pending-file",
		Status:          domain.JobStatusVirusCheckPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	completedJob := &domain.UploadJob{
		ID:              "completed-job",
		CreatedByUserId: "test-user",
		FileID:          "completed-file",
		Status:          domain.JobStatusCompleted,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	repo := newMockJobRepository()
	require.NoError(t, repo.Create(context.Background(), stuckJob))
	require.NoError(t, repo.Create(context.Background(), pendingJob))
	require.NoError(t, repo.Create(context.Background(), completedJob))

	fileStorage := &mockFileStorage{
		downloadFunc: func(ctx context.Context, fileID string) (io.ReadCloser, error) {
			return io.NopCloser(io.Reader(nil)), nil
		},
	}

	virusChecker := &mockVirusChecker{
		checkFunc: func(ctx context.Context, reader io.Reader) (bool, error) {
			return true, nil
		},
	}

	metrics := &mockMetrics{}
	fileInfoRepo := newMockFileInfoRepository()
	fileAuthorization := &mockFileAuthorization{}

	runner := NewVirusScannerJobRunner(
		repo,
		fileInfoRepo,
		fileAuthorization,
		fileStorage,
		virusChecker,
		5*time.Second,
		metrics,
	)

	jobsChan := make(chan *domain.UploadJob, 10)
	err := runner.queuePendingAndStuckJobs(context.Background(), jobsChan)
	require.NoError(t, err)

	// Both jobs should be sent to the channel
	receivedJobs := make([]*domain.UploadJob, 0)
	for i := 0; i < 2; i++ {
		select {
		case job := <-jobsChan:
			receivedJobs = append(receivedJobs, job)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for jobs")
		}
	}

	assert.Len(t, receivedJobs, 2)
	jobIDs := make(map[string]bool)
	for _, job := range receivedJobs {
		jobIDs[job.ID] = true
	}
	assert.True(t, jobIDs["stuck-job"])
	assert.True(t, jobIDs["pending-job"])
	assert.False(t, jobIDs["completed-job"], "completed job should not be processed")
}
