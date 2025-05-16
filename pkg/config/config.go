package config

import (
	"fmt"
	"os"

	"github.com/benjamin/file-storage-go/pkg/services/secrets"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort      string
	BlobStorageURL  string
	BlobAccountName string
	ContainerName   string
	StorageKey      string
	VaultAddress    string `mapstructure:"VAULT_ADDRESS"`
	VaultToken      string `mapstructure:"VAULT_TOKEN"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("BLOB_STORAGE_URL", "")
	viper.SetDefault("CONTAINER_NAME", "files")
	viper.SetDefault("VAULT_ADDRESS", "http://localhost:8200")
	viper.SetDefault("VAULT_TOKEN", "dev-token")
	viper.SetDefault("USE_AZURITE", "false")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		ServerPort:      viper.GetString("SERVER_PORT"),
		BlobStorageURL:  viper.GetString("BLOB_STORAGE_URL"),
		BlobAccountName: "",
		ContainerName:   viper.GetString("CONTAINER_NAME"),
		VaultAddress:    viper.GetString("VAULT_ADDRESS"),
		VaultToken:      viper.GetString("VAULT_TOKEN"),
	}

	if viper.GetBool("USE_AZURITE") {
		config.BlobStorageURL = "http://127.0.0.1:10000/devstoreaccount1"
		config.BlobAccountName = "devstoreaccount1"
		config.StorageKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	} else {
		// Try to get storage key from environment first
		storageKey := os.Getenv("STORAGE_KEY")
		if storageKey == "" {
			// If not in environment, try to get from Vault
			vaultService, err := secrets.NewVaultService(config.VaultAddress, config.VaultToken)
			if err != nil {
				return nil, fmt.Errorf("failed to create vault service: %w", err)
			}

			creds, err := vaultService.GetStorageCredentials()
			if err != nil {
				return nil, fmt.Errorf("failed to get storage credentials from vault: %w", err)
			}

			config.StorageKey = creds.StorageKey
			if config.BlobAccountName == "" {
				config.BlobAccountName = creds.AccountName
			}
			if config.BlobStorageURL == "" {
				config.BlobStorageURL = creds.StorageURL
			}
			if config.ContainerName == "" {
				config.ContainerName = creds.ContainerName
			}
		} else {
			config.StorageKey = storageKey
		}
	}

	if os.Getenv("USE_MOCK_STORAGE") != "true" && !viper.GetBool("USE_AZURITE") && config.BlobStorageURL == "" {
		return nil, fmt.Errorf("BLOB_STORAGE_URL is required when not using mock storage (USE_MOCK_STORAGE is not 'true') and not using Azurite (USE_AZURITE is not 'true')")
	}

	return config, nil
}
