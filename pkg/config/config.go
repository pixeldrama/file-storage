package config

import (
	"fmt"
	"log"
	"os"

	"github.com/benjamin/file-storage-go/pkg/services/secrets"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort          string `mapstructure:"SERVER_PORT"`
	BlobStorageURL      string `mapstructure:"BLOB_STORAGE_URL"`
	BlobAccountName     string `mapstructure:"BLOB_ACCOUNT_NAME"`
	ContainerName       string `mapstructure:"CONTAINER_NAME"`
	StorageKey          string `mapstructure:"STORAGE_KEY"`
	VaultAddress        string `mapstructure:"VAULT_ADDRESS"`
	VaultToken          string `mapstructure:"VAULT_TOKEN"`
	DBHost              string `mapstructure:"DB_HOST"`
	DBPort              string `mapstructure:"DB_PORT"`
	DBName              string `mapstructure:"DB_NAME"`
	DBUser              string `mapstructure:"DB_USER"`
	DBPassword          string `mapstructure:"DB_PASSWORD"`
	KeycloakURL         string `mapstructure:"KEYCLOAK_URL"`
	KeycloakClientID    string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	VirusCheckTimeout   string `mapstructure:"VIRUS_CHECK_TIMEOUT"`
	UseMockVirusChecker bool   `mapstructure:"USE_MOCK_VIRUS_CHECKER"`
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
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
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "file_storage")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("KEYCLOAK_URL", "http://localhost:8081/realms/file-storage")
	viper.SetDefault("KEYCLOAK_CLIENT_ID", "file-storage")
	viper.SetDefault("VIRUS_CHECK_TIMEOUT", "5s")
	viper.SetDefault("USE_MOCK_VIRUS_CHECKER", false)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		ServerPort:          viper.GetString("SERVER_PORT"),
		BlobStorageURL:      viper.GetString("BLOB_STORAGE_URL"),
		BlobAccountName:     viper.GetString("BLOB_ACCOUNT_NAME"),
		ContainerName:       viper.GetString("CONTAINER_NAME"),
		VaultAddress:        viper.GetString("VAULT_ADDRESS"),
		VaultToken:          viper.GetString("VAULT_TOKEN"),
		DBHost:              viper.GetString("DB_HOST"),
		DBPort:              viper.GetString("DB_PORT"),
		DBName:              viper.GetString("DB_NAME"),
		DBUser:              viper.GetString("DB_USER"),
		DBPassword:          viper.GetString("DB_PASSWORD"),
		KeycloakURL:         viper.GetString("KEYCLOAK_URL"),
		KeycloakClientID:    viper.GetString("KEYCLOAK_CLIENT_ID"),
		VirusCheckTimeout:   viper.GetString("VIRUS_CHECK_TIMEOUT"),
		UseMockVirusChecker: viper.GetBool("USE_MOCK_VIRUS_CHECKER"),
	}

	if os.Getenv("SKIP_STORAGE_VALIDATION") == "true" {
		return config, nil
	}

	storageKey := os.Getenv("STORAGE_KEY")
	if storageKey == "" {
		log.Printf("Creating vault service with address: %s", config.VaultAddress)
		vaultService, err := secrets.NewVaultService(config.VaultAddress, config.VaultToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault service: %w", err)
		}

		log.Printf("Retrieving storage credentials from vault")
		creds, err := vaultService.GetStorageCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to get storage credentials from vault: %w", err)
		}
		log.Printf("Successfully retrieved storage credentials from vault")

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

	if os.Getenv("USE_MOCK_STORAGE") != "true" && config.BlobStorageURL == "" {
		return nil, fmt.Errorf("BLOB_STORAGE_URL is required when not using mock storage (USE_MOCK_STORAGE is not 'true')")
	}

	return config, nil
}
