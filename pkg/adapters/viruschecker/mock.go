package viruschecker

import (
	"context"
	"io"
	"strings"
)

type MockVirusChecker struct{}

func NewMockVirusChecker() *MockVirusChecker {
	return &MockVirusChecker{}
}

func (c *MockVirusChecker) CheckFile(ctx context.Context, reader io.Reader) (bool, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return false, err
	}

	contentStr := strings.TrimSpace(string(content))
	return contentStr != "virus", nil
}
