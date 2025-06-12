package secrets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/approle/login":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"auth": map[string]interface{}{
					"client_token": "test-token",
				},
			})
		case "/v1/secret/data/storage":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"account_name":   "test-account",
							"storage_key":    "test-key",
							"storage_url":    "http://test-url",
							"container_name": "test-container",
						},
						"metadata": map[string]interface{}{
							"created_time": "2024-01-01T00:00:00Z",
							"version":      1,
						},
					},
				})
			} else if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"created_time": "2024-01-01T00:00:00Z",
						"version":      1,
					},
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service, err := NewVaultService(server.URL, "test-role-id", "test-secret-id")
	require.NoError(t, err)

	t.Run("GetStorageCredentials", func(t *testing.T) {
		creds, err := service.GetStorageCredentials()
		assert.NoError(t, err)
		assert.NotNil(t, creds)
		assert.Equal(t, "test-account", creds.AccountName)
		assert.Equal(t, "test-key", creds.StorageKey)
		assert.Equal(t, "http://test-url", creds.StorageURL)
		assert.Equal(t, "test-container", creds.ContainerName)
	})
}
