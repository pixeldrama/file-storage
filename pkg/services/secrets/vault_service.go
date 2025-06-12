package secrets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benjamin/file-storage-go/pkg/adapters/vault"
)

const (
	StorageCredsPath = "secret/data/storage"
)

type StorageCredentials struct {
	AccountName   string `json:"account_name"`
	StorageKey    string `json:"storage_key"`
	StorageURL    string `json:"storage_url"`
	ContainerName string `json:"container_name"`
}

type VaultService struct {
	client *vault.VaultClient
}

func NewVaultService(address, roleID, secretID string) (*VaultService, error) {
	token, err := getAppRoleToken(address, roleID, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app role token: %w", err)
	}

	client, err := vault.NewVaultClient(address, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultService{
		client: client,
	}, nil
}

func getAppRoleToken(address, roleID, secretID string) (string, error) {
	data := map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth data: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/auth/approle/login", strings.TrimRight(address, "/")),
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with app role: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate with app role: status code %d", resp.StatusCode)
	}

	var result struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	return result.Auth.ClientToken, nil
}

func (v *VaultService) GetStorageCredentials() (*StorageCredentials, error) {
	secret, err := v.client.GetSecret(StorageCredsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage credentials: %w", err)
	}

	data, ok := secret["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret data format")
	}

	creds := StorageCredentials{}
	if v, ok := data["account_name"].(string); ok {
		creds.AccountName = v
	} else {
		return nil, fmt.Errorf("invalid account_name format")
	}
	if v, ok := data["storage_key"].(string); ok {
		creds.StorageKey = v
	} else {
		return nil, fmt.Errorf("invalid storage_key format")
	}
	if v, ok := data["storage_url"].(string); ok {
		creds.StorageURL = v
	} else {
		return nil, fmt.Errorf("invalid storage_url format")
	}
	if v, ok := data["container_name"].(string); ok {
		creds.ContainerName = v
	} else {
		return nil, fmt.Errorf("invalid container_name format")
	}

	return &creds, nil
}
