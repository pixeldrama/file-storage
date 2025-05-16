package secrets

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/api"
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
	client *api.Client
}

func NewVaultService(address, token string) (*VaultService, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultService{
		client: client,
	}, nil
}

func (v *VaultService) StoreStorageCredentials(creds StorageCredentials) error {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"credentials": creds,
		},
	}
	_, err := v.client.Logical().Write(StorageCredsPath, data)
	if err != nil {
		return fmt.Errorf("failed to store storage credentials: %w", err)
	}
	return nil
}

func (v *VaultService) GetStorageCredentials() (*StorageCredentials, error) {
	secret, err := v.client.Logical().Read(StorageCredsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage credentials: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("no storage credentials found in vault at path: %s", StorageCredsPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret data format")
	}

	credsData, ok := data["credentials"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid credentials format in vault")
	}

	credsBytes, err := json.Marshal(credsData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credentials: %w", err)
	}

	var creds StorageCredentials
	if err := json.Unmarshal(credsBytes, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}
