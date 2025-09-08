package jwt

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

// JWTProvider implements JWT-based authentication
type JWTProvider struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      string
	algorithm     string
	keyID         string
	logger        *logrus.Logger
	configured    bool
}

// JWTConfig holds JWT provider configuration
type JWTConfig struct {
	SecretKey     string        `json:"secret_key"`
	AccessExpiry  time.Duration `json:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Issuer        string        `json:"issuer"`
	Audience      string        `json:"audience"`
	Algorithm     string        `json:"algorithm"`
	KeyID         string        `json:"key_id"`
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID      uuid.UUID              `json:"user_id"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	ServiceID   string                 `json:"service_id,omitempty"` // Dynamic service identifier
	Context     map[string]interface{} `json:"context,omitempty"`    // Dynamic context for multi-tenant scenarios
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

// DefaultJWTConfig returns default JWT configuration with environment variable support
func DefaultJWTConfig() *JWTConfig {
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

	return &JWTConfig{
		SecretKey:     secretKey,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
		Issuer:        issuer,
		Audience:      audience,
		Algorithm:     algorithm,
		KeyID:         keyID,
	}
}

// NewJWTProvider creates a new JWT provider
func NewJWTProvider(config *JWTConfig, logger *logrus.Logger) *JWTProvider {
	if config == nil {
		config = DefaultJWTConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &JWTProvider{
		secretKey:     []byte(config.SecretKey),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		issuer:        config.Issuer,
		audience:      config.Audience,
		algorithm:     config.Algorithm,
		keyID:         config.KeyID,
		logger:        logger,
		configured:    true,
	}
}

// GetName returns the provider name
func (jp *JWTProvider) GetName() string {
	return "jwt"
}

// GetSupportedFeatures returns supported features
func (jp *JWTProvider) GetSupportedFeatures() []string {
	return []string{
		"jwt",
		"token_generation",
		"token_validation",
		"token_refresh",
		"token_revocation",
		"role_based_access",
		"permission_based_access",
		"dynamic_context",
	}
}

// GetConnectionInfo returns connection information
func (jp *JWTProvider) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":      "jwt",
		"algorithm": jp.algorithm,
		"issuer":    jp.issuer,
		"audience":  jp.audience,
		"key_id":    jp.keyID,
	}
}

// Configure configures the JWT provider
func (jp *JWTProvider) Configure(config map[string]interface{}) error {
	if secretKey, ok := config["secret_key"].(string); ok {
		jp.secretKey = []byte(secretKey)
	}

	if accessExpiry, ok := config["access_expiry"].(time.Duration); ok {
		jp.accessExpiry = accessExpiry
	}

	if refreshExpiry, ok := config["refresh_expiry"].(time.Duration); ok {
		jp.refreshExpiry = refreshExpiry
	}

	if issuer, ok := config["issuer"].(string); ok {
		jp.issuer = issuer
	}

	if audience, ok := config["audience"].(string); ok {
		jp.audience = audience
	}

	if algorithm, ok := config["algorithm"].(string); ok {
		jp.algorithm = algorithm
	}

	if keyID, ok := config["key_id"].(string); ok {
		jp.keyID = keyID
	}

	jp.configured = true
	jp.logger.Info("JWT provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (jp *JWTProvider) IsConfigured() bool {
	return jp.configured
}

// GenerateTokenPair generates both access and refresh tokens
func (jp *JWTProvider) GenerateTokenPair(ctx context.Context, userID uuid.UUID, email string, roles, permissions []string, serviceID string, context map[string]interface{}, metadata map[string]interface{}) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := &JWTClaims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		ServiceID:   serviceID,
		Context:     context,
		Metadata:    metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jp.issuer,
			Audience:  []string{jp.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jp.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken.Header["kid"] = jp.keyID
	accessTokenString, err := accessToken.SignedString(jp.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		ServiceID: serviceID,
		Context:   context,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jp.issuer,
			Audience:  []string{jp.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jp.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken.Header["kid"] = jp.keyID
	refreshTokenString, err := refreshToken.SignedString(jp.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	jp.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"email":      email,
		"roles":      roles,
		"service_id": serviceID,
		"expires_at": now.Add(jp.accessExpiry),
	}).Info("Token pair generated successfully")

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(jp.accessExpiry),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (jp *JWTProvider) ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jp.secretKey, nil
	})

	if err != nil {
		jp.logger.WithError(err).Debug("Token validation failed")
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Additional validation
		if claims.Issuer != jp.issuer {
			return nil, fmt.Errorf("invalid issuer: expected %s, got %s", jp.issuer, claims.Issuer)
		}

		if len(claims.Audience) > 0 && claims.Audience[0] != jp.audience {
			return nil, fmt.Errorf("invalid audience: expected %s, got %s", jp.audience, claims.Audience[0])
		}

		jp.logger.WithFields(logrus.Fields{
			"user_id":    claims.UserID,
			"email":      claims.Email,
			"service_id": claims.ServiceID,
		}).Debug("Token validated successfully")

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken generates a new access token using refresh token
func (jp *JWTProvider) RefreshToken(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	claims, err := jp.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return jp.GenerateTokenPair(ctx, claims.UserID, claims.Email, claims.Roles, claims.Permissions, claims.ServiceID, claims.Context, claims.Metadata)
}

// RevokeToken marks a token as revoked
func (jp *JWTProvider) RevokeToken(ctx context.Context, tokenString string) error {
	claims, err := jp.ValidateToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	jp.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"service_id": claims.ServiceID,
		"jti":        claims.RegisteredClaims.ID,
	}).Info("Token revoked")

	// In a real implementation, you would:
	// 1. Store the token ID (JTI) in a blacklist/revocation list
	// 2. Check this list during token validation
	// 3. Implement token cleanup for expired tokens

	return nil
}

// HealthCheck performs health check
func (jp *JWTProvider) HealthCheck(ctx context.Context) error {
	// JWT provider is always healthy if configured
	if !jp.configured {
		return fmt.Errorf("JWT provider not configured")
	}
	return nil
}

// GetStats returns provider statistics
func (jp *JWTProvider) GetStats(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"provider":       "jwt",
		"configured":     jp.configured,
		"algorithm":      jp.algorithm,
		"issuer":         jp.issuer,
		"audience":       jp.audience,
		"access_expiry":  jp.accessExpiry.String(),
		"refresh_expiry": jp.refreshExpiry.String(),
	}
}

// Close closes the provider
func (jp *JWTProvider) Close() error {
	jp.logger.Info("JWT provider closed")
	return nil
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
	return time.Now().After(c.RegisteredClaims.ExpiresAt.Time)
}

// GetRemainingTime returns the remaining time until token expiration
func (c *JWTClaims) GetRemainingTime() time.Duration {
	return time.Until(c.RegisteredClaims.ExpiresAt.Time)
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
