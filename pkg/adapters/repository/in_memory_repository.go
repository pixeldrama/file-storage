package repository

import (
	"context"
	"sync"

	"github.com/benjamin/file-storage-go/pkg/domain"
)

type InMemoryRepository struct {
	jobs map[string]*domain.UploadJob
	mu   sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		jobs: make(map[string]*domain.UploadJob),
	}
}

func (r *InMemoryRepository) Create(ctx context.Context, job *domain.UploadJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryRepository) Get(ctx context.Context, jobID string) (*domain.UploadJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.jobs[jobID]
	if !exists {
		return nil, nil
	}

	return job, nil
}

func (r *InMemoryRepository) Update(ctx context.Context, job *domain.UploadJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return nil
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryRepository) GetByFileID(ctx context.Context, fileID string) (*domain.UploadJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, job := range r.jobs {
		if job.FileID == fileID {
			return job, nil
		}
	}
	return nil, nil // Or an error indicating not found
}
