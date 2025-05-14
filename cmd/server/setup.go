package server

import (
	"github.com/benjamin/file-storage-go/pkg/adapters/http"
	"github.com/benjamin/file-storage-go/pkg/domain" // For domain.FileStorage and domain.UploadJobRepository
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter initializes and returns a new Gin engine with all application routes.
// It takes the file storage service and job repository as dependencies.
// The metrics collector is expected to have registered its metrics globally.
func SetupRouter(
	fileStorage domain.FileStorage,
	jobRepo domain.UploadJobRepository,
) *gin.Engine {
	handlers := http.NewHandlers(fileStorage, jobRepo)

	r := gin.Default() // gin.Default() includes logger and recovery middleware

	// Metrics endpoint
	// promhttp.Handler() uses the default global registry, which NewPrometheusMetrics populates.
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	{
		api.POST("/upload-jobs", handlers.CreateUploadJob)
		api.GET("/upload-jobs/:jobId", handlers.GetUploadJobStatus)
		api.POST("/upload-jobs/:jobId", handlers.UploadFile)
		api.GET("/files/:fileId", handlers.DownloadFile)
		// api.DELETE("/files/:fileId", handlers.DeleteFile) // Uncomment when implemented
	}

	return r
}
