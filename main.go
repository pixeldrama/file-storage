package main

import (
	"context"
	"log"
	"os"
	"time"

	"file-storage-go/cmd/server"
	"file-storage-go/pkg/adapters/jobrunner"
	"file-storage-go/pkg/adapters/metrics"
	"file-storage-go/pkg/adapters/repository"
	"file-storage-go/pkg/adapters/storage"
	"file-storage-go/pkg/adapters/viruschecker"
	"file-storage-go/pkg/config"
	"file-storage-go/pkg/domain"
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
			log.Println("Warning: BlobAccountName is not set. This is fine for Azurite if BlobStorageURL is the Azurite URL. For real Azure, ensure BlobAccountName is configured.")
			accountNameForCreds = cfg.BlobStorageURL
		}

		fileStorage, azureStorageErr = storage.NewAzureBlobStorage(
			accountNameForCreds,
			cfg.BlobStorageURL,
			cfg.StorageKey,
			cfg.ContainerName,
			metricsCollector,
		)
		if azureStorageErr != nil {
			log.Fatalf("Failed to initialize AzureBlobStorage client: %v", azureStorageErr)
		}
	}

	var jobRepo domain.UploadJobRepository
	if cfg.UseInMemoryRepo {
		log.Println("INFO: Using InMemoryRepository because USE_IN_MEMORY_REPO is set to true.")
		jobRepo = repository.NewInMemoryRepository()
	} else {
		log.Println("INFO: Using PostgresRepository.")
		jobRepo, err = repository.NewPostgresRepository(cfg.GetDBConnString())
		if err != nil {
			log.Fatalf("Failed to create postgres repository: %v", err)
		}
	}

	var virusChecker domain.VirusChecker
	if cfg.UseMockVirusChecker {
		log.Println("INFO: Using MockVirusChecker because USE_MOCK_VIRUS_CHECKER is set to true.")
		virusChecker = viruschecker.NewMockVirusChecker()
	} else {
		log.Println("INFO: Using HTTPVirusChecker.")
		if cfg.VirusCheckerURL == "" {
			log.Fatalf("VIRUS_CHECKER_URL is required when not using mock virus checker")
		}
		virusChecker = viruschecker.NewHTTPVirusChecker(cfg.VirusCheckerURL)
	}

	virusCheckTimeout, err := time.ParseDuration(cfg.VirusCheckTimeout)
	if err != nil {
		log.Fatalf("Invalid VIRUS_CHECK_TIMEOUT format: %v", err)
	}

	virusScanner := jobrunner.NewVirusScannerJobRunner(
		jobRepo,
		fileStorage,
		virusChecker,
		virusCheckTimeout,
		metricsCollector,
	)

	go virusScanner.Start(context.Background())

	serverConfig := server.ServerConfig{
		FileStorage:      fileStorage,
		JobRepo:          jobRepo,
		KeycloakURL:      cfg.KeycloakURL,
		KeycloakClientID: cfg.KeycloakClientID,
	}

	r := server.SetupRouter(serverConfig)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
