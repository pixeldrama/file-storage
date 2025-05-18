package server

import (
	"net/http"

	handlers "github.com/benjamin/file-storage-go/pkg/adapters/http"
	"github.com/benjamin/file-storage-go/pkg/domain"
	"github.com/benjamin/file-storage-go/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(
	fileStorage domain.FileStorage,
	jobRepo domain.UploadJobRepository,
) *gin.Engine {
	h := handlers.NewHandlers(fileStorage, jobRepo)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Apply auth middleware to all routes except health and metrics
	r.Use(middleware.AuthMiddleware())

	r.POST("/upload-jobs", h.CreateUploadJob)
	r.GET("/upload-jobs/:jobId", h.GetUploadJobStatus)
	r.POST("/upload-jobs/:jobId", h.UploadFile)
	r.GET("/files/:fileId", h.DownloadFile)
	r.DELETE("/files/:fileId", h.DeleteFile)

	return r
}
