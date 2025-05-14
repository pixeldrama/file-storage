package main

import (
	"fmt"

	"github.com/benjamin/file-storage-go/cmd/server"
	"github.com/benjamin/file-storage-go/pkg/adapters/metrics"
	"github.com/benjamin/file-storage-go/pkg/adapters/repository"
	"github.com/benjamin/file-storage-go/pkg/adapters/storage"
	"github.com/benjamin/file-storage-go/pkg/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	metricsCollector := metrics.NewPrometheusMetrics()

	fileStorage, err := storage.NewAzureBlobStorage(
		cfg.BlobStorageURL,
		cfg.StorageKey,
		cfg.ContainerName,
		metricsCollector,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize storage client: %v", err))
	}

	jobRepo := repository.NewInMemoryRepository()

	r := server.SetupRouter(fileStorage, jobRepo)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
