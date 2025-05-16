package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/benjamin/file-storage-go/pkg/domain"
)

type MockStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

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

func (ms *MockStorage) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data, ok := ms.files[fileID]
	if !ok {
		return nil, fmt.Errorf("mockstorage: file with ID '%s' not found", fileID)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (ms *MockStorage) Delete(ctx context.Context, fileID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.files, fileID)
	return nil
}

var _ domain.FileStorage = (*MockStorage)(nil)
