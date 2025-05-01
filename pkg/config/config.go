package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort     string
	BlobStorageURL string
	ContainerName  string
	StorageKey     string
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

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Get storage key from vault (mocked for now)
	storageKey := getStorageKeyFromVault()

	config := &Config{
		ServerPort:     viper.GetString("SERVER_PORT"),
		BlobStorageURL: viper.GetString("BLOB_STORAGE_URL"),
		ContainerName:  viper.GetString("CONTAINER_NAME"),
		StorageKey:     storageKey,
	}

	// Validate required fields
	if config.BlobStorageURL == "" {
		return nil, fmt.Errorf("BLOB_STORAGE_URL is required")
	}

	return config, nil
}

// Mocked vault integration
func getStorageKeyFromVault() string {
	// In a real implementation, this would fetch the key from Azure Key Vault
	// For now, we'll use an environment variable
	key := os.Getenv("STORAGE_KEY")
	if key == "" {
		// For development purposes, use a mock key
		return "mock-storage-key"
	}
	return key
}
