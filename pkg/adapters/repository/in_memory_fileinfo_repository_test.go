package repository

import (
	"context"
	"testing"
	"time"

	"file-storage-go/pkg/domain"
)

func TestInMemoryFileInfoRepo_Create(t *testing.T) {
	repo := NewInMemoryFileInfoRepo()
	ctx := context.Background()
	fileInfo := &domain.FileInfo{
		ID:                 "test-file",
		Filename:           "test.txt",
		FileType:           "document",
		LinkedResourceType: "project",
		LinkedResourceID:   "project-123",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, fileInfo)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if len(repo.fileInfos) != 1 {
		t.Errorf("Expected 1 file info in repository, got %d", len(repo.fileInfos))
	}

	stored := repo.fileInfos[fileInfo.ID]
	if stored.ID != fileInfo.ID {
		t.Errorf("Stored file info does not match input file info")
	}
}

func TestInMemoryFileInfoRepo_Get(t *testing.T) {
	repo := NewInMemoryFileInfoRepo()
	ctx := context.Background()
	fileInfo := &domain.FileInfo{
		ID:                 "test-file",
		Filename:           "test.txt",
		FileType:           "document",
		LinkedResourceType: "project",
		LinkedResourceID:   "project-123",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	repo.fileInfos[fileInfo.ID] = fileInfo

	retrieved, err := repo.Get(ctx, fileInfo.ID)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected file info to be retrieved, got nil")
	}
	if retrieved.ID != fileInfo.ID {
		t.Errorf("Retrieved file info does not match stored file info")
	}

	nonExistent, err := repo.Get(ctx, "non-existent")
	if err != nil {
		t.Errorf("Get for non-existent file info failed: %v", err)
	}
	if nonExistent != nil {
		t.Error("Expected nil for non-existent file info")
	}
}

func TestInMemoryFileInfoRepo_Update(t *testing.T) {
	repo := NewInMemoryFileInfoRepo()
	ctx := context.Background()
	fileInfo := &domain.FileInfo{
		ID:                 "test-file",
		Filename:           "test.txt",
		FileType:           "document",
		LinkedResourceType: "project",
		LinkedResourceID:   "project-123",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	repo.fileInfos[fileInfo.ID] = fileInfo

	updatedFileInfo := &domain.FileInfo{
		ID:                 fileInfo.ID,
		Filename:           "updated.txt",
		FileType:           fileInfo.FileType,
		LinkedResourceType: fileInfo.LinkedResourceType,
		LinkedResourceID:   fileInfo.LinkedResourceID,
		CreatedAt:          fileInfo.CreatedAt,
		UpdatedAt:          time.Now(),
	}

	err := repo.Update(ctx, updatedFileInfo)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	stored := repo.fileInfos[fileInfo.ID]
	if stored.Filename != "updated.txt" {
		t.Errorf("Expected filename to be 'updated.txt', got %q", stored.Filename)
	}

	err = repo.Update(ctx, &domain.FileInfo{ID: "non-existent"})
	if err != nil {
		t.Errorf("Update for non-existent file info should not return error: %v", err)
	}
}

func TestInMemoryFileInfoRepo_Delete(t *testing.T) {
	repo := NewInMemoryFileInfoRepo()
	ctx := context.Background()
	fileInfo := &domain.FileInfo{
		ID:                 "test-file",
		Filename:           "test.txt",
		FileType:           "document",
		LinkedResourceType: "project",
		LinkedResourceID:   "project-123",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	repo.fileInfos[fileInfo.ID] = fileInfo

	err := repo.Delete(ctx, fileInfo.ID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if len(repo.fileInfos) != 0 {
		t.Errorf("Expected 0 file infos after delete, got %d", len(repo.fileInfos))
	}

	err = repo.Delete(ctx, "non-existent")
	if err != nil {
		t.Errorf("Delete for non-existent file info should not return error: %v", err)
	}
}
