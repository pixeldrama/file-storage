package auth

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type MockJWTVerifier struct{}

func NewMockJWTVerifier() *MockJWTVerifier {
	return &MockJWTVerifier{}
}

func (m *MockJWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	claims := &Claims{
		UserId: "mock-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "mock-user-id",
		},
	}

	token := &jwt.Token{
		Claims: claims,
		Valid:  true,
	}

	return token, nil
}

func (m *MockJWTVerifier) ExtractTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("empty authorization header")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
