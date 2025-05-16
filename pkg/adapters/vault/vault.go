package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *api.Client
}

func NewVaultClient(address, token string) (*VaultClient, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultClient{
		client: client,
	}, nil
}

// StoreSecret stores a secret in Vault
func (v *VaultClient) StoreSecret(path string, data map[string]interface{}) error {
	_, err := v.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}
	return nil
}

// GetSecret retrieves a secret from Vault
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
