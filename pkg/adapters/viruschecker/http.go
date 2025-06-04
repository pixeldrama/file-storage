package viruschecker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type httpResponse struct {
	Success bool   `json:"success"`
	Clean   bool   `json:"clean"`
	Message string `json:"message"`
}

type HTTPVirusChecker struct {
	client  *http.Client
	baseURL string
}

func NewHTTPVirusChecker() (*HTTPVirusChecker, error) {
	baseURL := os.Getenv("VIRUS_CHECKER_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("VIRUS_CHECKER_URL environment variable is required")
	}

	return &HTTPVirusChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

func (c *HTTPVirusChecker) CheckFile(ctx context.Context, reader io.Reader) (bool, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result httpResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return false, fmt.Errorf("virus check failed: %s", result.Message)
	}

	return result.Clean, nil
}
