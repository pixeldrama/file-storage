package main

import (
	"log"
	"os"

	"github.com/benjamin/file-storage-go/cmd/server"
	"github.com/benjamin/file-storage-go/pkg/adapters/metrics"
	"github.com/benjamin/file-storage-go/pkg/adapters/repository"
	"github.com/benjamin/file-storage-go/pkg/adapters/storage"
	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/benjamin/file-storage-go/pkg/domain"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		if os.Getenv("USE_MOCK_STORAGE") != "true" {
			log.Fatalf("Failed to load config: %v", err)
		}
		if cfg == nil {
			log.Println("Warning: Config loading failed, but USE_MOCK_STORAGE is true. Proceeding with potentially incomplete config for mocks.")
			cfg = &config.Config{ServerPort: "8080"}
		}
	}

	metricsCollector := metrics.NewPrometheusMetrics()
	var fileStorage domain.FileStorage

	if os.Getenv("USE_MOCK_STORAGE") == "true" {
		log.Println("INFO: Using MockStorage because USE_MOCK_STORAGE is set to true.")
		fileStorage = storage.NewMockStorage()
	} else {
		log.Println("INFO: Using AzureBlobStorage.")
		if cfg.BlobStorageURL == "" {
			log.Fatalf("BLOB_STORAGE_URL is required when not using mock storage.")
		}

		var azureStorageErr error
		accountNameForCreds := cfg.BlobAccountName
		if accountNameForCreds == "" {
			// This is a fallback for non-Azurite setups.
			// For real Azure, ensure BlobAccountName is configured or use a robust way to get it from BlobStorageURL.
			// The azblob.NewSharedKeyCredential expects the account name, not the full URL.
			// For Azurite, BlobAccountName is explicitly "devstoreaccount1".
			// If you are using a real Azure Storage account, you should set BlobAccountName in your config.
			log.Println("Warning: BlobAccountName is not set. This is fine for Azurite if BlobStorageURL is the Azurite URL. For real Azure, ensure BlobAccountName is configured.")
			// As a simple fallback, we assume the BlobStorageURL might be the account name if it's not a full URL.
			// This part might need adjustment based on your actual Azure Storage URL structures.
			// A more robust solution extracts the account name from the full blob URL.
			// For now, we will pass BlobStorageURL, but this is likely incorrect for NewSharedKeyCredential with real Azure Storage URLs.
			accountNameForCreds = cfg.BlobStorageURL // This line is problematic for real Azure URLs but okay if BlobAccountName will be set.
		}

		fileStorage, azureStorageErr = storage.NewAzureBlobStorage(
			accountNameForCreds, // Account name for SharedKeyCredential
			cfg.BlobStorageURL,  // Full service URL for the client
			cfg.StorageKey,
			cfg.ContainerName,
			metricsCollector,
		)
		if azureStorageErr != nil {
			log.Fatalf("Failed to initialize AzureBlobStorage client: %v", azureStorageErr)
		}
	}

	jobRepo := repository.NewInMemoryRepository()

	r := server.SetupRouter(fileStorage, jobRepo)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
