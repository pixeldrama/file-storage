package server

import (
	"log/slog"
	"net/http"

	handlers "file-storage-go/pkg/adapters/http"
	"file-storage-go/pkg/auth"
	"file-storage-go/pkg/domain"
	"file-storage-go/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerConfig struct {
	FileStorage          domain.FileStorage
	JobRepo              domain.UploadJobRepository
	FileInfoRepo         domain.FileInfoRepository
	FileAuthorization    domain.FileAuthorization
	KeycloakURL          string
	KeycloakClientID     string
	UseMockAuthorization bool
	Logger           *slog.Logger
}

func SetupRouter(config ServerConfig) *gin.Engine {
	h := handlers.NewHandlers(config.FileStorage, config.JobRepo, config.FileInfoRepo, config.FileAuthorization)

	// Create a new Gin engine without any default middleware
	r := gin.New()

	// Use our custom ECS logger middleware
	r.Use(middleware.GinLoggerMiddleware(config.Logger))

	// Use Gin's recovery middleware to handle panics
	r.Use(gin.Recovery())

	// Disable Gin's debug output
	gin.SetMode(gin.ReleaseMode)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	var jwtVerifier auth.JWTVerifierInterface
	if config.UseMockAuthorization {
		config.Logger.Info("Using MockJWTVerifier because UseMockAuthorization is set to true.")
		jwtVerifier = auth.NewMockJWTVerifier()
	} else {
		config.Logger.Info("Using KeycloakJWTVerifier")
		jwtVerifier = auth.NewJWTVerifier(auth.KeycloakConfig{
			RealmURL: config.KeycloakURL,
			ClientID: config.KeycloakClientID,
		})
	}

	// Apply auth middleware to all routes except health and metrics
	r.Use(middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		JWTVerifier: jwtVerifier,
	}))

	r.Use(middleware.RequireUserId())

	r.POST("/upload-jobs", h.CreateUploadJob)
	r.GET("/upload-jobs/:jobId", h.GetUploadJobStatus)
	r.POST("/upload-jobs/:jobId", h.UploadFile)
	r.GET("/files/:fileId", h.GetFileInfo)
	r.GET("/files/:fileId/download", h.DownloadFile)
	r.DELETE("/files/:fileId", h.DeleteFile)

	return r
}
