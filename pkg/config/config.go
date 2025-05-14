package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort      string
	BlobStorageURL  string
	BlobAccountName string
	ContainerName   string
	StorageKey      string
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
	viper.SetDefault("VAULT_URL", "")
	viper.SetDefault("USE_AZURITE", "false")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Get storage key from vault (mocked for now)
	storageKey := getStorageKeyFromVault(viper.GetBool("USE_AZURITE"))

	config := &Config{
		ServerPort:      viper.GetString("SERVER_PORT"),
		BlobStorageURL:  viper.GetString("BLOB_STORAGE_URL"),
		BlobAccountName: "",
		ContainerName:   viper.GetString("CONTAINER_NAME"),
		StorageKey:      storageKey,
	}

	if viper.GetBool("USE_AZURITE") {
		config.BlobStorageURL = "http://127.0.0.1:10000/devstoreaccount1"
		config.BlobAccountName = "devstoreaccount1"
		// The container name is part of the BlobStorageURL for Azurite,
		// or it can be created if it doesn't exist.
		// We'll keep ContainerName for potential separate use, but Azurite's default setup includes it in the URL.
	}

	if os.Getenv("USE_MOCK_STORAGE") != "true" && !viper.GetBool("USE_AZURITE") && config.BlobStorageURL == "" {
		return nil, fmt.Errorf("BLOB_STORAGE_URL is required when not using mock storage (USE_MOCK_STORAGE is not 'true') and not using Azurite (USE_AZURITE is not 'true')")
	}

	return config, nil
}

// Mocked vault integration
func getStorageKeyFromVault(useAzurite bool) string {
	if useAzurite {
		// Default Azurite account key
		return "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	}
	// In a real implementation, this would fetch the key from Azure Key Vault
	// For now, we'll use an environment variable
	key := os.Getenv("STORAGE_KEY")
	if key == "" {
		// For development purposes, use a mock key
		return "mock-storage-key"
	}
	return key
}
