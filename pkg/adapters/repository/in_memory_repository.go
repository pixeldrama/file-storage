package repository

import (
	"context"
	"sync"

	"file-storage-go/pkg/domain"
)

type InMemoryJobRepo struct {
	jobs map[string]*domain.UploadJob
	mu   sync.RWMutex
}

func NewInMemoryJobRepo() *InMemoryJobRepo {
	return &InMemoryJobRepo{
		jobs: make(map[string]*domain.UploadJob),
	}
}

func (r *InMemoryJobRepo) Create(ctx context.Context, job *domain.UploadJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryJobRepo) Get(ctx context.Context, jobID string) (*domain.UploadJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.jobs[jobID]
	if !exists {
		return nil, nil
	}

	return job, nil
}

func (r *InMemoryJobRepo) Update(ctx context.Context, job *domain.UploadJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return nil
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryJobRepo) GetByFileID(ctx context.Context, fileID string) (*domain.UploadJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, job := range r.jobs {
		if job.FileID == fileID {
			return job, nil
		}
	}
	return nil, nil
}

func (r *InMemoryJobRepo) GetByStatus(ctx context.Context, status domain.JobStatus) ([]*domain.UploadJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var jobs []*domain.UploadJob
	for _, job := range r.jobs {
		if job.Status == status {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

type InMemoryFileInfoRepo struct {
	fileInfos map[string]*domain.FileInfo
	mu        sync.RWMutex
}

func NewInMemoryFileInfoRepo() *InMemoryFileInfoRepo {
	return &InMemoryFileInfoRepo{
		fileInfos: make(map[string]*domain.FileInfo),
	}
}

func (r *InMemoryFileInfoRepo) Create(ctx context.Context, fileInfo *domain.FileInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fileInfos[fileInfo.ID] = fileInfo
	return nil
}

func (r *InMemoryFileInfoRepo) Get(ctx context.Context, fileID string) (*domain.FileInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fileInfo, exists := r.fileInfos[fileID]
	if !exists {
		return nil, nil
	}

	return fileInfo, nil
}

func (r *InMemoryFileInfoRepo) Update(ctx context.Context, fileInfo *domain.FileInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.fileInfos[fileInfo.ID]; !exists {
		return nil
	}

	r.fileInfos[fileInfo.ID] = fileInfo
	return nil
}

func (r *InMemoryFileInfoRepo) Delete(ctx context.Context, fileID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.fileInfos, fileID)
	return nil
}
