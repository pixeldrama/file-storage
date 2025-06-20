package repository

import (
	"context"
	"testing"
	"time"

	"file-storage-go/pkg/domain"
)

func TestInMemoryRepository_Create(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	job := &domain.UploadJob{
		ID:              "test-job",
		CreatedByUserId: "test-user",
		Filename:        "test.txt",
		Status:          domain.JobStatusUploading,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := repo.Create(ctx, job)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if len(repo.jobs) != 1 {
		t.Errorf("Expected 1 job in repository, got %d", len(repo.jobs))
	}

	stored := repo.jobs[job.ID]
	if stored.ID != job.ID || stored.Filename != job.Filename {
		t.Errorf("Stored job does not match input job")
	}
}

func TestInMemoryRepository_Get(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	job := &domain.UploadJob{
		ID:              "test-job",
		CreatedByUserId: "test-user",
		Filename:        "test.txt",
		Status:          domain.JobStatusUploading,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	repo.jobs[job.ID] = job

	retrieved, err := repo.Get(ctx, job.ID)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected job to be retrieved, got nil")
	}
	if retrieved.ID != job.ID || retrieved.Filename != job.Filename {
		t.Errorf("Retrieved job does not match stored job")
	}

	nonExistent, err := repo.Get(ctx, "non-existent")
	if err != nil {
		t.Errorf("Get for non-existent job failed: %v", err)
	}
	if nonExistent != nil {
		t.Error("Expected nil for non-existent job")
	}
}

func TestInMemoryRepository_Update(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	job := &domain.UploadJob{
		ID:              "test-job",
		CreatedByUserId: "test-user",
		Filename:        "test.txt",
		Status:          domain.JobStatusUploading,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	repo.jobs[job.ID] = job

	updatedJob := &domain.UploadJob{
		ID:              job.ID,
		CreatedByUserId: job.CreatedByUserId,
		Filename:        job.Filename,
		Status:          domain.JobStatusCompleted,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       time.Now(),
	}

	err := repo.Update(ctx, updatedJob)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	stored := repo.jobs[job.ID]
	if stored.Status != domain.JobStatusCompleted {
		t.Errorf("Expected status to be 'completed', got %q", stored.Status)
	}

	err = repo.Update(ctx, &domain.UploadJob{ID: "non-existent"})
	if err != nil {
		t.Errorf("Update for non-existent job should not return error: %v", err)
	}
}

func TestInMemoryRepository_GetByFileID(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	job := &domain.UploadJob{
		ID:              "test-job",
		CreatedByUserId: "test-user",
		Filename:        "test.txt",
		Status:          domain.JobStatusCompleted,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FileID:          "test-file-id",
	}

	repo.jobs[job.ID] = job

	retrieved, err := repo.GetByFileID(ctx, job.FileID)
	if err != nil {
		t.Errorf("GetByFileID failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected job to be retrieved, got nil")
	}
	if retrieved.FileID != job.FileID {
		t.Errorf("Retrieved job does not match stored job")
	}

	nonExistent, err := repo.GetByFileID(ctx, "non-existent")
	if err != nil {
		t.Errorf("GetByFileID for non-existent file failed: %v", err)
	}
	if nonExistent != nil {
		t.Error("Expected nil for non-existent file ID")
	}
}

func TestInMemoryRepository_ConcurrentOperations(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			job := &domain.UploadJob{
				ID:              "job1",
				CreatedByUserId: "test-user",
				Status:          domain.JobStatusUploading,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			repo.Create(ctx, job)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			repo.Get(ctx, "job1")
		}
		done <- true
	}()

	<-done
	<-done

	if len(repo.jobs) != 1 {
		t.Errorf("Expected 1 job after concurrent operations, got %d", len(repo.jobs))
	}
}
