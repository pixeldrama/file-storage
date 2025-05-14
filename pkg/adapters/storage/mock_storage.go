package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/benjamin/file-storage-go/pkg/domain"
)

// MockStorage is an in-memory implementation of domain.FileStorage for testing.
type MockStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewMockStorage creates a new MockStorage.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

// Upload stores the file content in memory.
func (ms *MockStorage) Upload(ctx context.Context, fileID string, reader io.Reader) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("mockstorage: failed to read data for upload: %w", err)
	}

	ms.files[fileID] = data
	return nil
}

// Download retrieves the file content from memory.
func (ms *MockStorage) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data, ok := ms.files[fileID]
	if !ok {
		return nil, fmt.Errorf("mockstorage: file with ID '%s' not found", fileID)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Delete removes the file from memory.
func (ms *MockStorage) Delete(ctx context.Context, fileID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.files[fileID]; !ok {
		return fmt.Errorf("mockstorage: file with ID '%s' not found for delete", fileID)
	}
	delete(ms.files, fileID)
	return nil
}

// Compile-time check to ensure MockStorage implements domain.FileStorage
var _ domain.FileStorage = (*MockStorage)(nil)
