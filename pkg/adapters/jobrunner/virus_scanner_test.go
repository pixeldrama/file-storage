package jobrunner

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/benjamin/file-storage-go/pkg/domain"
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
				ID:        "test-job",
				FileID:    "test-file",
				Status:    domain.JobStatusVirusCheckPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			checkResult:    true,
			expectedStatus: domain.JobStatusCompleted,
		},
		{
			name: "virus check failed - malware detected",
			job: &domain.UploadJob{
				ID:        "test-job",
				FileID:    "test-file",
				Status:    domain.JobStatusVirusCheckPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			checkResult:    false,
			expectedStatus: domain.JobStatusFailed,
			expectedError:  "file contains malware",
		},
		{
			name: "download error",
			job: &domain.UploadJob{
				ID:        "test-job",
				FileID:    "test-file",
				Status:    domain.JobStatusVirusCheckPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			downloadErr:    errors.New("download failed"),
			expectedStatus: domain.JobStatusFailed,
			expectedError:  "failed to download file: download failed",
		},
		{
			name: "virus check error",
			job: &domain.UploadJob{
				ID:        "test-job",
				FileID:    "test-file",
				Status:    domain.JobStatusVirusCheckPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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

			runner := NewVirusScannerJobRunner(
				repo,
				fileStorage,
				virusChecker,
				5*time.Second,
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
		ID:        "stuck-job",
		FileID:    "stuck-file",
		Status:    domain.JobStatusVirusChecking,
		CreatedAt: now.Add(-10 * time.Minute),
		UpdatedAt: now.Add(-6 * time.Second),
	}

	pendingJob := &domain.UploadJob{
		ID:        "pending-job",
		FileID:    "pending-file",
		Status:    domain.JobStatusVirusCheckPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	completedJob := &domain.UploadJob{
		ID:        "completed-job",
		FileID:    "completed-file",
		Status:    domain.JobStatusCompleted,
		CreatedAt: now,
		UpdatedAt: now,
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

	runner := NewVirusScannerJobRunner(
		repo,
		fileStorage,
		virusChecker,
		5*time.Second,
	)

	jobsChan := make(chan *domain.UploadJob, 10)
	err := runner.processJobs(context.Background(), jobsChan)
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
