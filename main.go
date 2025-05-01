package main

import (
	"fmt"

	"github.com/benjamin/file-storage-go/pkg/adapters/http"
	"github.com/benjamin/file-storage-go/pkg/adapters/metrics"
	"github.com/benjamin/file-storage-go/pkg/adapters/repository"
	"github.com/benjamin/file-storage-go/pkg/adapters/storage"
	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize metrics collector
	metricsCollector := metrics.NewPrometheusMetrics()

	// Initialize storage client
	fileStorage, err := storage.NewAzureBlobStorage(
		cfg.BlobStorageURL,
		cfg.StorageKey,
		cfg.ContainerName,
		metricsCollector,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize storage client: %v", err))
	}

	// Initialize repository
	jobRepo := repository.NewInMemoryRepository()

	// Initialize HTTP handlers
	handlers := http.NewHandlers(fileStorage, jobRepo)

	// Initialize Gin router
	r := gin.Default()

	// Add Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	{
		api.POST("/upload-jobs", handlers.CreateUploadJob)
		api.GET("/upload-jobs/:jobId", handlers.GetUploadJobStatus)
		api.POST("/upload-jobs/:jobId", handlers.UploadFile)
		api.GET("/files/:fileId", handlers.DownloadFile)
	}

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
