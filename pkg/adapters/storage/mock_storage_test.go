package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestMockStorage_Upload(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()
	fileID := "test-file"
	content := "test content"

	err := storage.Upload(ctx, fileID, strings.NewReader(content))
	if err != nil {
		t.Errorf("Upload failed: %v", err)
	}

	if len(storage.files) != 1 {
		t.Errorf("Expected 1 file in storage, got %d", len(storage.files))
	}

	if string(storage.files[fileID]) != content {
		t.Errorf("Expected content %q, got %q", content, string(storage.files[fileID]))
	}
}

func TestMockStorage_Download(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()
	fileID := "test-file"
	content := "test content"

	storage.files[fileID] = []byte(content)

	reader, err := storage.Download(ctx, fileID)
	if err != nil {
		t.Errorf("Download failed: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("Failed to read downloaded content: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}

	_, err = storage.Download(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestMockStorage_Delete(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()
	fileID := "test-file"
	content := "test content"

	storage.files[fileID] = []byte(content)

	err := storage.Delete(ctx, fileID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if len(storage.files) != 0 {
		t.Errorf("Expected 0 files after delete, got %d", len(storage.files))
	}
}

func TestMockStorage_ConcurrentOperations(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			storage.Upload(ctx, "file1", bytes.NewReader([]byte("content1")))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			storage.Upload(ctx, "file2", bytes.NewReader([]byte("content2")))
		}
		done <- true
	}()

	<-done
	<-done

	if len(storage.files) != 2 {
		t.Errorf("Expected 2 files after concurrent uploads, got %d", len(storage.files))
	}
}
