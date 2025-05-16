package server

import (
	"github.com/benjamin/file-storage-go/pkg/adapters/http"
	"github.com/benjamin/file-storage-go/pkg/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(
	fileStorage domain.FileStorage,
	jobRepo domain.UploadJobRepository,
) *gin.Engine {
	handlers := http.NewHandlers(fileStorage, jobRepo)

	r := gin.Default()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/upload-jobs", handlers.CreateUploadJob)
	r.GET("/upload-jobs/:jobId", handlers.GetUploadJobStatus)
	r.POST("/upload-jobs/:jobId", handlers.UploadFile)
	r.GET("/files/:fileId", handlers.DownloadFile)

	return r
}
