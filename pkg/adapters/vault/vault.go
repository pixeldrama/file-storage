package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *api.Client
}

func NewVaultClient(address, token string) (*VaultClient, error) {
	config := api.DefaultConfig()
	config.Address = address
	config.Timeout = 10 * time.Second

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultClient{
		client: client,
	}, nil
}

func (v *VaultClient) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	return secret.Data, nil
}

func (v *VaultClient) ListSecrets(path string) ([]string, error) {
	secret, err := v.client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("no secrets found at path: %s", path)
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from vault")
	}

	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.(string)
	}

	return result, nil
}

func (v *VaultClient) StoreSecret(path string, data map[string]interface{}) error {
	_, err := v.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}
	return nil
}
