package middleware

import (
	"net/http"

	"github.com/benjamin/file-storage-go/pkg/auth"
	"github.com/gin-gonic/gin"
)

type AuthMiddlewareConfig struct {
	JWTVerifier *auth.JWTVerifier
}

func NewAuthMiddleware(config AuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString, err := config.JWTVerifier.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := config.JWTVerifier.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Store the token claims in the context for later use
		c.Set("token", token)
		c.Next()
	}
}
