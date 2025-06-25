package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTVerifierInterface interface {
	VerifyToken(tokenString string) (*jwt.Token, error)
	ExtractTokenFromHeader(header string) (string, error)
}
