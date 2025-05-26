package secrets

import (
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
