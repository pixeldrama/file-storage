package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointsRequireAuthentication(t *testing.T) {
	savedAuthToken := authToken
	authToken = ""
	defer func() {
		authToken = savedAuthToken
	}()

	t.Run("POST /upload-jobs requires authentication", func(t *testing.T) {
		reqBody := map[string]string{
			"filename": "test.txt",
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := createAPIRequest(t, "POST", fmt.Sprintf("%s/upload-jobs", apiBaseURL), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GET /upload-jobs/{jobId} requires authentication", func(t *testing.T) {
		req := createAPIRequest(t, "GET", fmt.Sprintf("%s/upload-jobs/%s", apiBaseURL, "some-job-id"), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("POST /upload-jobs/{jobId} requires authentication", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		require.NoError(t, err)
		_, err = part.Write([]byte("test content"))
		require.NoError(t, err)
		writer.Close()

		req := createAPIRequest(t, "POST", fmt.Sprintf("%s/upload-jobs/%s", apiBaseURL, "some-job-id"), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GET /files/{fileId} requires authentication", func(t *testing.T) {
		req := createAPIRequest(t, "GET", fmt.Sprintf("%s/files/%s", apiBaseURL, "some-file-id"), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("DELETE /files/{fileId} requires authentication", func(t *testing.T) {
		req := createAPIRequest(t, "DELETE", fmt.Sprintf("%s/files/%s", apiBaseURL, "some-file-id"), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Health endpoint does not require authentication", func(t *testing.T) {
		req := createAPIRequest(t, "GET", fmt.Sprintf("%s/health", apiBaseURL), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Metrics endpoint does not require authentication", func(t *testing.T) {
		req := createAPIRequest(t, "GET", fmt.Sprintf("%s/metrics", apiBaseURL), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
