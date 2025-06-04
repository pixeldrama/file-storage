package viruschecker

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
)

type MockVirusChecker struct{}

func NewMockVirusChecker() *MockVirusChecker {
	return &MockVirusChecker{}
}

func (c *MockVirusChecker) CheckFile(ctx context.Context, reader io.Reader) (bool, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(content)) == "virus", nil
}
