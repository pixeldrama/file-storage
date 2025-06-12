package config

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	Address    string
	RoleID     string
	VaultToken string
	Client     *api.Client
}

func NewVaultConfig(address, roleID, vaultToken string) (*VaultConfig, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultConfig{
		Address:    address,
		RoleID:     roleID,
		VaultToken: vaultToken,
		Client:     client,
	}, nil
}

func (v *VaultConfig) Login() error {
	data := map[string]interface{}{
		"role_id":   v.RoleID,
		"secret_id": v.VaultToken,
	}

	secret, err := v.Client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return fmt.Errorf("failed to login to vault: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth info in response")
	}

	v.Client.SetToken(secret.Auth.ClientToken)
	return nil
}

func (v *VaultConfig) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := v.Client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	return secret.Data, nil
}
