package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// "crypto/rand" // Keep for future tests needing random data
	// "github.com/google/uuid" // Keep for future tests needing UUIDs

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Configuration for the API base URL and auth token
// These should be configurable, e.g., via environment variables for real test runs.
var (
	apiBaseURL = getEnv("API_BASE_URL", "http://localhost:8080")
	authToken  = os.Getenv("API_AUTH_TOKEN")
)

const (
	testfileName    = "test_file.txt"
	testfileContent = "This is a test file for acceptance testing."
	pollInterval    = 1 * time.Second
	pollTimeout     = 30 * time.Second
)

type UploadJobResponse struct {
	JobID     string    `json:"jobId"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"createdAt"`
}

type UploadJobStatusResponse struct {
	JobID     string    `json:"jobId"`
	Filename  string    `json:"filename"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error,omitempty"`
	FileID    string    `json:"fileId,omitempty"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func TestMain(m *testing.M) {
	// Setup: Potentially check if the API is running, or seed data.
	// For now, we'll assume the API is running at apiBaseURL.

	// Check and log API_BASE_URL
	if apiBaseURL == "" {
		fmt.Println("Error: API_BASE_URL is not set and no default was provided. Please set the API_BASE_URL environment variable.")
		os.Exit(1)
	}
	fmt.Printf("INFO: Using API Base URL: %s\n", apiBaseURL)
	if authToken == "" {
		fmt.Println("WARNING: API_AUTH_TOKEN is not set. Tests requiring auth may fail or be skipped.")
	}

	// Create a dummy test file
	err := os.WriteFile(testfileName, []byte(testfileContent), 0644)
	if err != nil {
		fmt.Printf("Failed to create test file: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	// Teardown: Clean up dummy test file
	os.Remove(testfileName)

	os.Exit(code)
}

// --- HTTP Client Helpers ---

func createAPIRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err, "Failed to create request")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return req
}

func executeAPIRequest(t *testing.T, client *http.Client, req *http.Request, expectedStatusCode int) *http.Response {
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute request")
	// Read body for logging even if not used by caller, then restore it
	var respBodyBytes []byte
	if resp.Body != nil {
		respBodyBytes, _ = io.ReadAll(resp.Body)
		resp.Body.Close()                                        // close original body
		resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes)) // restore body
	}

	require.Equal(t, expectedStatusCode, resp.StatusCode,
		fmt.Sprintf("Expected status %d, got %d. URL: %s %s. Response body: %s",
			expectedStatusCode, resp.StatusCode, req.Method, req.URL.String(), string(respBodyBytes)))
	return resp
}

func decodeJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err, "Failed to decode JSON response")
}

// --- API Specific Client Functions ---

func createUploadJobClient(t *testing.T, httpClient *http.Client) UploadJobResponse {
	url := fmt.Sprintf("%s/upload-jobs", apiBaseURL)
	// The API spec says POST /upload-jobs, but doesn't specify a body for creation.
	// Assuming it needs an empty body or a body indicating the filename (which is in the response).
	// For now, sending a minimal JSON with filename, adjust if API needs different.
	payload := strings.NewReader(fmt.Sprintf(`{"filename": "%s"}`, testfileName))
	req := createAPIRequest(t, http.MethodPost, url, payload)
	req.Header.Set("Content-Type", "application/json")

	resp := executeAPIRequest(t, httpClient, req, http.StatusCreated)

	var jobResponse UploadJobResponse
	decodeJSONResponse(t, resp, &jobResponse)
	require.NotEmpty(t, jobResponse.JobID, "JobID should not be empty")
	return jobResponse
}

func uploadFileForJobClient(t *testing.T, httpClient *http.Client, jobId string, filePath string) UploadJobStatusResponse {
	url := fmt.Sprintf("%s/upload-jobs/%s", apiBaseURL, jobId)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	require.NoError(t, err, "Failed to open test file")
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err, "Failed to create form file")

	_, err = io.Copy(part, file)
	require.NoError(t, err, "Failed to copy file to form")

	err = writer.Close()
	require.NoError(t, err, "Failed to close multipart writer")

	req := createAPIRequest(t, http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := executeAPIRequest(t, httpClient, req, http.StatusCreated) // API spec says 201 for successful upload

	var statusResponse UploadJobStatusResponse
	decodeJSONResponse(t, resp, &statusResponse)
	return statusResponse
}

func getJobStatusClient(t *testing.T, httpClient *http.Client, jobId string) UploadJobStatusResponse {
	url := fmt.Sprintf("%s/upload-jobs/%s", apiBaseURL, jobId)
	req := createAPIRequest(t, http.MethodGet, url, nil)

	resp := executeAPIRequest(t, httpClient, req, http.StatusOK)

	var statusResponse UploadJobStatusResponse
	decodeJSONResponse(t, resp, &statusResponse)
	return statusResponse
}

func downloadFileClient(t *testing.T, httpClient *http.Client, fileId string) ([]byte, string) {
	url := fmt.Sprintf("%s/files/%s", apiBaseURL, fileId)
	req := createAPIRequest(t, http.MethodGet, url, nil)

	resp := executeAPIRequest(t, httpClient, req, http.StatusOK)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read downloaded file content")

	contentDisposition := resp.Header.Get("Content-Disposition")
	return data, contentDisposition
}

func deleteFileClient(t *testing.T, httpClient *http.Client, fileId string) {
	url := fmt.Sprintf("%s/files/%s", apiBaseURL, fileId)
	req := createAPIRequest(t, http.MethodDelete, url, nil)
	executeAPIRequest(t, httpClient, req, http.StatusNoContent)
}

// --- Test Cases ---

func TestFileLifecycle_SuccessfulUploadDownloadDelete(t *testing.T) {
	// Initialize HTTP client for this test
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// 1. Create an upload job
	t.Log("Step 1: Creating upload job...")
	job := createUploadJobClient(t, httpClient)
	assert.NotEmpty(t, job.JobID, "Job ID should be returned")
	t.Logf("Upload job created with ID: %s", job.JobID)

	// 2. Upload a test file
	t.Logf("Step 2: Uploading file '%s' for job ID: %s...", testfileName, job.JobID)
	initialStatus := uploadFileForJobClient(t, httpClient, job.JobID, testfileName)
	assert.Contains(t, []string{"PENDING", "UPLOADING", "VIRUS_CHECKING", "COMPLETED"}, initialStatus.Status, "Initial status after upload")
	t.Logf("File upload initiated, initial status: %s", initialStatus.Status)

	// 3. Poll the job status until COMPLETED
	t.Log("Step 3: Polling job status...")
	var currentStatus UploadJobStatusResponse
	startTime := time.Now()
	for {
		currentStatus = getJobStatusClient(t, httpClient, job.JobID)
		t.Logf("Current job status: %s (FileID: %s)", currentStatus.Status, currentStatus.FileID)
		if currentStatus.Status == "COMPLETED" {
			assert.NotEmpty(t, currentStatus.FileID, "FileID should be present when status is COMPLETED")
			break
		}
		if currentStatus.Status == "FAILED" {
			t.Fatalf("Upload job failed: %s", currentStatus.Error)
		}
		if time.Since(startTime) > pollTimeout {
			t.Fatalf("Polling timed out after %v, last status: %s", pollTimeout, currentStatus.Status)
		}
		time.Sleep(pollInterval)
	}
	t.Logf("Job completed. FileID: %s", currentStatus.FileID)

	// 4. Verify the Location header and extract fileId (FileID is directly in status response)
	fileId := currentStatus.FileID
	require.NotEmpty(t, fileId, "FileID from completed job status should not be empty")
	// The OpenAPI spec mentions a Location header on the GET /upload-jobs/{jobId} endpoint
	// when status is COMPLETED. Let's verify that too.
	// We need to make the call again to get the headers specifically after completion.
	finalJobStatusReq := createAPIRequest(t, http.MethodGet, fmt.Sprintf("%s/upload-jobs/%s", apiBaseURL, job.JobID), nil)
	finalJobStatusResp := executeAPIRequest(t, httpClient, finalJobStatusReq, http.StatusOK)
	locationHeader := finalJobStatusResp.Header.Get("Location")
	assert.NotEmpty(t, locationHeader, "Location header should be present for completed job")
	expectedLocationSuffix := fmt.Sprintf("/files/%s", fileId)
	assert.True(t, strings.HasSuffix(locationHeader, expectedLocationSuffix),
		fmt.Sprintf("Location header '%s' should end with '%s'", locationHeader, expectedLocationSuffix))
	finalJobStatusResp.Body.Close()

	// 5. Download the file using fileId
	t.Logf("Step 5: Downloading file with FileID: %s...", fileId)
	downloadedData, contentDisposition := downloadFileClient(t, httpClient, fileId)
	assert.NotEmpty(t, downloadedData, "Downloaded data should not be empty")
	t.Logf("File downloaded successfully (%d bytes). Content-Disposition: %s", len(downloadedData), contentDisposition)

	// 6. Verify the downloaded file matches the uploaded file
	t.Log("Step 6: Verifying downloaded file content...")
	assert.Equal(t, testfileContent, string(downloadedData), "Downloaded file content should match original")
	// Also check Content-Disposition if it's expected to contain the original filename
	assert.Contains(t, contentDisposition, fmt.Sprintf("filename=\"%s\"", testfileName), "Content-Disposition should contain the correct filename")
	t.Log("File content verified.")

	// 7. Delete the file
	t.Logf("Step 7: Deleting file with FileID: %s...", fileId)
	deleteFileClient(t, httpClient, fileId)
	t.Log("File deleted successfully.")

	// 8. Attempt to download the deleted file and expect a 404
	t.Log("Step 8: Attempting to download deleted file (expecting 404)...")
	url := fmt.Sprintf("%s/files/%s", apiBaseURL, fileId)
	req := createAPIRequest(t, http.MethodGet, url, nil)
	_ = executeAPIRequest(t, httpClient, req, http.StatusNotFound) // Expect 404
	t.Log("Download attempt for deleted file correctly resulted in 404.")
}
