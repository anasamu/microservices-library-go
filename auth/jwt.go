package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      string
	logger        *logrus.Logger
}

// JWTManagerConfig holds JWT manager configuration
type JWTManagerConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
	Audience      string
	Algorithm     string
	KeyID         string
}

// DefaultJWTConfig returns default JWT configuration with environment variable support
func DefaultJWTConfig() *JWTManagerConfig {
	accessExpiry := 15 * time.Minute
	if val := os.Getenv("JWT_ACCESS_EXPIRY"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			accessExpiry = duration
		}
	}

	refreshExpiry := 7 * 24 * time.Hour
	if val := os.Getenv("JWT_REFRESH_EXPIRY"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			refreshExpiry = duration
		}
	}

	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "microservices-library"
	}

	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = "microservices-clients"
	}

	algorithm := os.Getenv("JWT_ALGORITHM")
	if algorithm == "" {
		algorithm = "HS256"
	}

	keyID := os.Getenv("JWT_KEY_ID")
	if keyID == "" {
		keyID = "default-key-id"
	}

	return &JWTManagerConfig{
		SecretKey:     secretKey,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
		Issuer:        issuer,
		Audience:      audience,
		Algorithm:     algorithm,
		KeyID:         keyID,
	}
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID      uuid.UUID              `json:"user_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config *JWTManagerConfig, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(config.SecretKey),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		issuer:        config.Issuer,
		audience:      config.Audience,
		logger:        logger,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (jm *JWTManager) GenerateTokenPair(ctx context.Context, userID, tenantID uuid.UUID, email string, roles, permissions []string, metadata map[string]interface{}) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := &JWTClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		Metadata:    metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Audience:  []string{jm.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jm.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Audience:  []string{jm.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jm.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	jm.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"email":      email,
		"roles":      roles,
		"expires_at": now.Add(jm.accessExpiry),
	}).Info("Token pair generated successfully")

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(jm.accessExpiry),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (jm *JWTManager) ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	})

	if err != nil {
		jm.logger.WithError(err).Debug("Token validation failed")
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Additional validation
		if claims.Issuer != jm.issuer {
			return nil, fmt.Errorf("invalid issuer: expected %s, got %s", jm.issuer, claims.Issuer)
		}

		if len(claims.Audience) > 0 && claims.Audience[0] != jm.audience {
			return nil, fmt.Errorf("invalid audience: expected %s, got %s", jm.audience, claims.Audience[0])
		}

		jm.logger.WithFields(logrus.Fields{
			"user_id":   claims.UserID,
			"tenant_id": claims.TenantID,
			"email":     claims.Email,
		}).Debug("Token validated successfully")

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken generates a new access token using refresh token
func (jm *JWTManager) RefreshToken(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	claims, err := jm.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return jm.GenerateTokenPair(ctx, claims.UserID, claims.TenantID, claims.Email, claims.Roles, claims.Permissions, claims.Metadata)
}

// RevokeToken marks a token as revoked (in a real implementation, you'd store this in a database)
func (jm *JWTManager) RevokeToken(ctx context.Context, tokenString string) error {
	claims, err := jm.ValidateToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	jm.logger.WithFields(logrus.Fields{
		"user_id":   claims.UserID,
		"tenant_id": claims.TenantID,
		"jti":       claims.ID,
	}).Info("Token revoked")

	// In a real implementation, you would:
	// 1. Store the token ID (JTI) in a blacklist/revocation list
	// 2. Check this list during token validation
	// 3. Implement token cleanup for expired tokens

	return nil
}

// ExtractTokenFromContext extracts token from context
func (jm *JWTManager) ExtractTokenFromContext(ctx context.Context) (string, error) {
	// This would typically extract from gRPC metadata or HTTP headers
	// Implementation depends on your context structure
	return "", fmt.Errorf("not implemented")
}

// HasRole checks if the user has a specific role
func (c *JWTClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (c *JWTClaims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (c *JWTClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the user has any of the specified permissions
func (c *JWTClaims) HasAnyPermission(permissions ...string) bool {
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}
	return false
}

// IsExpired checks if the token is expired
func (c *JWTClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time)
}

// GetRemainingTime returns the remaining time until token expiration
func (c *JWTClaims) GetRemainingTime() time.Duration {
	return time.Until(c.ExpiresAt.Time)
}

// GenerateRSAKeyPair generates RSA key pair for JWT signing
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// EncodeRSAPrivateKeyToPEM encodes RSA private key to PEM format
func EncodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	return pem.EncodeToMemory(&privBlock)
}

// EncodeRSAPublicKeyToPEM encodes RSA public key to PEM format
func EncodeRSAPublicKeyToPEM(publicKey *rsa.PublicKey) []byte {
	pubDER := x509.MarshalPKCS1PublicKey(publicKey)
	pubBlock := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   pubDER,
	}
	return pem.EncodeToMemory(&pubBlock)
}
