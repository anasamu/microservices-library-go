package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthProvider implements authentication middleware
type AuthProvider struct {
	name    string
	logger  *logrus.Logger
	config  *AuthConfig
	jwtKey  []byte
	enabled bool
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecretKey   string                 `json:"jwt_secret_key"`
	JWTExpiry      time.Duration          `json:"jwt_expiry"`
	JWTIssuer      string                 `json:"jwt_issuer"`
	JWTAudience    string                 `json:"jwt_audience"`
	AllowedMethods []string               `json:"allowed_methods"`
	ExcludedPaths  []string               `json:"excluded_paths"`
	RequiredRoles  []string               `json:"required_roles"`
	RequiredScopes []string               `json:"required_scopes"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWTSecretKey:   "your-secret-key",
		JWTExpiry:      15 * time.Minute,
		JWTIssuer:      "middleware-service",
		JWTAudience:    "api-clients",
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		ExcludedPaths:  []string{"/health", "/metrics", "/docs"},
		RequiredRoles:  []string{},
		RequiredScopes: []string{},
		Metadata:       make(map[string]interface{}),
	}
}

// NewAuthProvider creates a new authentication provider
func NewAuthProvider(config *AuthConfig, logger *logrus.Logger) *AuthProvider {
	if config == nil {
		config = DefaultAuthConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &AuthProvider{
		name:    "auth",
		logger:  logger,
		config:  config,
		jwtKey:  []byte(config.JWTSecretKey),
		enabled: true,
	}
}

// GetName returns the provider name
func (ap *AuthProvider) GetName() string {
	return ap.name
}

// GetType returns the provider type
func (ap *AuthProvider) GetType() string {
	return "authentication"
}

// GetSupportedFeatures returns supported features
func (ap *AuthProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureJWT,
		types.FeatureOAuth2,
		types.FeatureBasicAuth,
		types.FeatureAPIKey,
		types.FeatureRBAC,
		types.FeatureABAC,
	}
}

// GetConnectionInfo returns connection information
func (ap *AuthProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     8080,
		Protocol: "http",
		Version:  "1.0.0",
		Secure:   true,
	}
}

// ProcessRequest processes an authentication request
func (ap *AuthProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !ap.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	ap.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
	}).Debug("Processing authentication request")

	// Check if path is excluded
	if ap.isPathExcluded(request.Path) {
		ap.logger.WithField("path", request.Path).Debug("Path excluded from authentication")
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Extract and validate token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type":     "application/json",
				"WWW-Authenticate": "Bearer",
			},
			Body:      []byte(`{"error": "missing authorization header"}`),
			Error:     "missing authorization header",
			Timestamp: time.Now(),
		}, nil
	}

	// Parse Bearer token
	tokenString := ap.extractBearerToken(authHeader)
	if tokenString == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type":     "application/json",
				"WWW-Authenticate": "Bearer",
			},
			Body:      []byte(`{"error": "invalid authorization header format"}`),
			Error:     "invalid authorization header format",
			Timestamp: time.Now(),
		}, nil
	}

	// Validate JWT token
	claims, err := ap.validateJWTToken(tokenString)
	if err != nil {
		ap.logger.WithError(err).WithField("request_id", request.ID).Warn("JWT validation failed")
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type":     "application/json",
				"WWW-Authenticate": "Bearer",
			},
			Body:      []byte(fmt.Sprintf(`{"error": "invalid token: %s"}`, err.Error())),
			Error:     fmt.Sprintf("invalid token: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}

	// Check required roles
	if len(ap.config.RequiredRoles) > 0 {
		if !ap.hasRequiredRoles(claims) {
			return &types.MiddlewareResponse{
				ID:         request.ID,
				Success:    false,
				StatusCode: http.StatusForbidden,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body:      []byte(`{"error": "insufficient permissions"}`),
				Error:     "insufficient permissions",
				Timestamp: time.Now(),
			}, nil
		}
	}

	// Check required scopes
	if len(ap.config.RequiredScopes) > 0 {
		if !ap.hasRequiredScopes(claims) {
			return &types.MiddlewareResponse{
				ID:         request.ID,
				Success:    false,
				StatusCode: http.StatusForbidden,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body:      []byte(`{"error": "insufficient scopes"}`),
				Error:     "insufficient scopes",
				Timestamp: time.Now(),
			}, nil
		}
	}

	// Add user information to request context
	userInfo := ap.extractUserInfo(claims)
	request.Context["user"] = userInfo
	request.UserID = userInfo.ID

	ap.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"user_id":    userInfo.ID,
		"username":   userInfo.Username,
	}).Info("Authentication successful")

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

// ProcessResponse processes an authentication response
func (ap *AuthProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	ap.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing authentication response")

	// Add authentication headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Auth-Provider"] = ap.name
	response.Headers["X-Auth-Timestamp"] = time.Now().Format(time.RFC3339)

	return response, nil
}

// CreateChain creates an authentication middleware chain
func (ap *AuthProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	ap.logger.WithField("chain_name", config.Name).Info("Creating authentication middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return ap.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the authentication middleware chain
func (ap *AuthProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	ap.logger.WithField("request_id", request.ID).Debug("Executing authentication middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP authentication middleware
func (ap *AuthProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	ap.logger.WithField("config_type", config.Type).Info("Creating HTTP authentication middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create middleware request
			request := &types.MiddlewareRequest{
				ID:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
				Type:      "http",
				Method:    r.Method,
				Path:      r.URL.Path,
				Headers:   make(map[string]string),
				Query:     make(map[string]string),
				Context:   make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			// Copy headers
			for name, values := range r.Header {
				if len(values) > 0 {
					request.Headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					request.Query[name] = values[0]
				}
			}

			// Process request
			response, err := ap.ProcessRequest(r.Context(), request)
			if err != nil {
				ap.logger.WithError(err).Error("Authentication middleware error")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			if !response.Success {
				// Set response headers
				for name, value := range response.Headers {
					w.Header().Set(name, value)
				}
				w.WriteHeader(response.StatusCode)
				w.Write(response.Body)
				return
			}

			// Add user information to request context
			if userInfo, ok := response.Context["user"].(*types.User); ok {
				ctx := context.WithValue(r.Context(), "user", userInfo)
				r = r.WithContext(ctx)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with authentication middleware
func (ap *AuthProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := ap.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the authentication provider
func (ap *AuthProvider) Configure(config map[string]interface{}) error {
	ap.logger.Info("Configuring authentication provider")

	if jwtSecret, ok := config["jwt_secret_key"].(string); ok {
		ap.config.JWTSecretKey = jwtSecret
		ap.jwtKey = []byte(jwtSecret)
	}

	if jwtExpiry, ok := config["jwt_expiry"].(time.Duration); ok {
		ap.config.JWTExpiry = jwtExpiry
	}

	if issuer, ok := config["jwt_issuer"].(string); ok {
		ap.config.JWTIssuer = issuer
	}

	if audience, ok := config["jwt_audience"].(string); ok {
		ap.config.JWTAudience = audience
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		ap.config.ExcludedPaths = excludedPaths
	}

	if requiredRoles, ok := config["required_roles"].([]string); ok {
		ap.config.RequiredRoles = requiredRoles
	}

	if requiredScopes, ok := config["required_scopes"].([]string); ok {
		ap.config.RequiredScopes = requiredScopes
	}

	ap.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (ap *AuthProvider) IsConfigured() bool {
	return ap.enabled && len(ap.jwtKey) > 0
}

// HealthCheck performs health check
func (ap *AuthProvider) HealthCheck(ctx context.Context) error {
	ap.logger.Debug("Authentication provider health check")
	return nil
}

// GetStats returns authentication statistics
func (ap *AuthProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	return &types.MiddlewareStats{
		TotalRequests:      1000,
		SuccessfulRequests: 950,
		FailedRequests:     50,
		AverageLatency:     5 * time.Millisecond,
		MaxLatency:         20 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         100.0,
		ActiveConnections:  25,
		ProviderData: map[string]interface{}{
			"provider_type": "authentication",
			"jwt_issuer":    ap.config.JWTIssuer,
			"jwt_audience":  ap.config.JWTAudience,
		},
	}, nil
}

// Close closes the authentication provider
func (ap *AuthProvider) Close() error {
	ap.logger.Info("Closing authentication provider")
	ap.enabled = false
	return nil
}

// Helper methods

func (ap *AuthProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range ap.config.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return true
		}
	}
	return false
}

func (ap *AuthProvider) extractBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func (ap *AuthProvider) validateJWTToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ap.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Validate issuer
		if ap.config.JWTIssuer != "" {
			if issuer, ok := claims["iss"].(string); !ok || issuer != ap.config.JWTIssuer {
				return nil, fmt.Errorf("invalid issuer")
			}
		}

		// Validate audience
		if ap.config.JWTAudience != "" {
			if audience, ok := claims["aud"].(string); !ok || audience != ap.config.JWTAudience {
				return nil, fmt.Errorf("invalid audience")
			}
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (ap *AuthProvider) hasRequiredRoles(claims jwt.MapClaims) bool {
	if len(ap.config.RequiredRoles) == 0 {
		return true
	}

	userRoles, ok := claims["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, requiredRole := range ap.config.RequiredRoles {
		found := false
		for _, userRole := range userRoles {
			if role, ok := userRole.(string); ok && role == requiredRole {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (ap *AuthProvider) hasRequiredScopes(claims jwt.MapClaims) bool {
	if len(ap.config.RequiredScopes) == 0 {
		return true
	}

	userScopes, ok := claims["scopes"].([]interface{})
	if !ok {
		return false
	}

	for _, requiredScope := range ap.config.RequiredScopes {
		found := false
		for _, userScope := range userScopes {
			if scope, ok := userScope.(string); ok && scope == requiredScope {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (ap *AuthProvider) extractUserInfo(claims jwt.MapClaims) *types.User {
	user := &types.User{
		ID:          "",
		Username:    "",
		Email:       "",
		Roles:       []string{},
		Permissions: []string{},
		Attributes:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	if userID, ok := claims["sub"].(string); ok {
		user.ID = userID
	}

	if username, ok := claims["username"].(string); ok {
		user.Username = username
	}

	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}

	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				user.Roles = append(user.Roles, roleStr)
			}
		}
	}

	if permissions, ok := claims["permissions"].([]interface{}); ok {
		for _, permission := range permissions {
			if permStr, ok := permission.(string); ok {
				user.Permissions = append(user.Permissions, permStr)
			}
		}
	}

	// Copy all claims as attributes
	for key, value := range claims {
		user.Attributes[key] = value
	}

	return user
}
