package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/test":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"key": "value",
						},
						"metadata": map[string]interface{}{
							"created_time": "2024-01-01T00:00:00Z",
							"version":      1,
						},
					},
				})
			}
		case "/v1/secret/metadata":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"keys": []string{"test"},
					},
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewVaultClient(server.URL, "test-token")
	require.NoError(t, err)

	path := "secret/data/test"

	t.Run("GetSecret", func(t *testing.T) {
		data, err := client.GetSecret(path)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "value", data["data"].(map[string]interface{})["key"])
	})

	t.Run("ListSecrets", func(t *testing.T) {
		secrets, err := client.ListSecrets("secret/metadata")
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		assert.Contains(t, secrets, "test")
	})
}
