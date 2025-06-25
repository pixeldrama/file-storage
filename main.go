package main

import (
	"context"
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
	"file-storage-go/pkg/loginit"
)

func main() {
	// Initialize the ECS logger
	logger := loginit.InitLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		if os.Getenv("USE_MOCK_STORAGE") != "true" {
			logger.Error("Failed to load config", "error", err)
			os.Exit(1)
		}
		if cfg == nil {
			logger.Warn("Config loading failed, but USE_MOCK_STORAGE is true. Proceeding with potentially incomplete config for mocks.")
			cfg = &config.Config{ServerPort: "8080"}
		}
	}

	metricsCollector := metrics.NewPrometheusMetrics()
	var fileStorage domain.FileStorage

	if os.Getenv("USE_MOCK_STORAGE") == "true" {
		logger.Info("Using MockStorage because USE_MOCK_STORAGE is set to true.")
		fileStorage = storage.NewMockStorage()
	} else {
		logger.Info("Using AzureBlobStorage.")
		if cfg.BlobStorageURL == "" {
			logger.Error("BLOB_STORAGE_URL is required when not using mock storage.")
			os.Exit(1)
		}

		var azureStorageErr error
		accountNameForCreds := cfg.BlobAccountName
		if accountNameForCreds == "" {
			logger.Warn("BlobAccountName is not set. This is fine for Azurite if BlobStorageURL is the Azurite URL. For real Azure, ensure BlobAccountName is configured.")
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
			logger.Error("Failed to initialize AzureBlobStorage client", "error", azureStorageErr)
			os.Exit(1)
		}
	}

	var jobRepo domain.UploadJobRepository
	if cfg.UseInMemoryRepo {
		logger.Info("Using InMemoryRepository because USE_IN_MEMORY_REPO is set to true.")
		jobRepo = repository.NewInMemoryRepository()
	} else {
		logger.Info("Using PostgresRepository.")
		jobRepo, err = repository.NewPostgresRepository(cfg.GetDBConnString())
		if err != nil {
			logger.Error("Failed to create postgres repository", "error", err)
			os.Exit(1)
		}
	}

	var virusChecker domain.VirusChecker
	if cfg.UseMockVirusChecker {
		logger.Info("Using MockVirusChecker because USE_MOCK_VIRUS_CHECKER is set to true.")
		virusChecker = viruschecker.NewMockVirusChecker()
	} else {
		logger.Info("Using HTTPVirusChecker.")
		if cfg.VirusCheckerURL == "" {
			logger.Error("VIRUS_CHECKER_URL is required when not using mock virus checker")
			os.Exit(1)
		}
		virusChecker = viruschecker.NewHTTPVirusChecker(cfg.VirusCheckerURL)
	}

	virusCheckTimeout, err := time.ParseDuration(cfg.VirusCheckTimeout)
	if err != nil {
		logger.Error("Invalid VIRUS_CHECK_TIMEOUT format", "error", err)
		os.Exit(1)
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
		Logger:           logger,
	}

	r := server.SetupRouter(serverConfig)

	logger.Info("Starting server", "port", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
