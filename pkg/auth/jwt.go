package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type KeycloakConfig struct {
	RealmURL string
	ClientID string
}

type KeycloakPublicKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type KeycloakKeys struct {
	Keys []KeycloakPublicKey `json:"keys"`
}

type JWTVerifier struct {
	config     KeycloakConfig
	httpClient *http.Client
	keys       map[string]KeycloakPublicKey
	mu         sync.RWMutex
}

func NewJWTVerifier(config KeycloakConfig) *JWTVerifier {
	return &JWTVerifier{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       make(map[string]KeycloakPublicKey),
	}
}

func (v *JWTVerifier) fetchPublicKeys() error {
	url := fmt.Sprintf("%s/protocol/openid-connect/certs", v.config.RealmURL)
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch public keys: %w", err)
	}
	defer resp.Body.Close()

	var keys KeycloakKeys
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return fmt.Errorf("failed to decode public keys: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	for _, key := range keys.Keys {
		v.keys[key.Kid] = key
	}

	return nil
}

func (v *JWTVerifier) getPublicKey(kid string) (KeycloakPublicKey, error) {
	v.mu.RLock()
	key, exists := v.keys[kid]
	v.mu.RUnlock()

	if !exists {
		if err := v.fetchPublicKeys(); err != nil {
			return KeycloakPublicKey{}, err
		}

		v.mu.RLock()
		key, exists = v.keys[kid]
		v.mu.RUnlock()

		if !exists {
			return KeycloakPublicKey{}, fmt.Errorf("public key not found for kid: %s", kid)
		}
	}

	return key, nil
}

func (v *JWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token header: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid not found in token header")
	}

	key, err := v.getPublicKey(kid)
	if err != nil {
		return nil, err
	}

	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, fmt.Errorf("failed to decode modulus: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, fmt.Errorf("failed to decode exponent: %w", err)
		}

		n := new(big.Int).SetBytes(nBytes)
		e := new(big.Int).SetBytes(eBytes)

		publicKey := &rsa.PublicKey{
			N: n,
			E: int(e.Int64()),
		}

		return publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (v *JWTVerifier) ExtractTokenFromHeader(header string) (string, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}
