package server

import (
	"net/http"

	handlers "file-storage-go/pkg/adapters/http"
	"file-storage-go/pkg/auth"
	"file-storage-go/pkg/domain"
	"file-storage-go/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerConfig struct {
	FileStorage      domain.FileStorage
	JobRepo          domain.UploadJobRepository
	KeycloakURL      string
	KeycloakClientID string
}

func SetupRouter(config ServerConfig) *gin.Engine {
	h := handlers.NewHandlers(config.FileStorage, config.JobRepo)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize JWT verifier
	jwtVerifier := auth.NewJWTVerifier(auth.KeycloakConfig{
		RealmURL: config.KeycloakURL,
		ClientID: config.KeycloakClientID,
	})

	// Apply auth middleware to all routes except health and metrics
	r.Use(middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		JWTVerifier: jwtVerifier,
	}))

	r.POST("/upload-jobs", h.CreateUploadJob)
	r.GET("/upload-jobs/:jobId", h.GetUploadJobStatus)
	r.POST("/upload-jobs/:jobId", h.UploadFile)
	r.GET("/files/:fileId", h.DownloadFile)
	r.DELETE("/files/:fileId", h.DeleteFile)

	return r
}
