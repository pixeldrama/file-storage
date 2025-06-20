package middleware

import (
	"log"
	"net/http"

	"file-storage-go/pkg/auth"

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
			log.Println("Authorization header missing")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString, err := config.JWTVerifier.ExtractTokenFromHeader(authHeader)
		if err != nil {
			log.Printf("Failed to extract token from header: %v\n", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := config.JWTVerifier.VerifyToken(tokenString)
		if err != nil {
			log.Printf("Failed to verify token: %v\n", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok {
			log.Printf("Invalid token claims format")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func RequireUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		claimsInterface, exists := c.Get("claims")
		if !exists {
			log.Printf("No claims found in context")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := claimsInterface.(*auth.Claims)
		if !ok {
			log.Printf("Invalid claims format in context")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.UserId == "" {
			log.Printf("Missing userId in token claims")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userId", claims.UserId)
		c.Next()
	}
}
