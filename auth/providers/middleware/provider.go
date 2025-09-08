package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AuthMiddleware provides HTTP middleware for authentication and authorization
type AuthMiddleware struct {
	authManager interface{} // Will be *gateway.AuthManager
	logger      *logrus.Logger
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	SkipPaths       []string          `json:"skip_paths"`
	TokenHeader     string            `json:"token_header"`
	TokenQueryParam string            `json:"token_query_param"`
	TokenCookie     string            `json:"token_cookie"`
	DefaultProvider string            `json:"default_provider"`
	RequireAuth     bool              `json:"require_auth"`
	RequireRoles    []string          `json:"require_roles"`
	RequirePerms    []string          `json:"require_permissions"`
	Metadata        map[string]string `json:"metadata"`
}

// AuthContextKey represents the key for auth context
type AuthContextKey string

const (
	UserIDKey    AuthContextKey = "user_id"
	UserEmailKey AuthContextKey = "user_email"
	UserRolesKey AuthContextKey = "user_roles"
	UserPermsKey AuthContextKey = "user_permissions"
	TokenKey     AuthContextKey = "token"
	ProviderKey  AuthContextKey = "provider"
	ServiceIDKey AuthContextKey = "service_id"
	ContextKey   AuthContextKey = "context"
)

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths:       []string{"/health", "/metrics", "/docs"},
		TokenHeader:     "Authorization",
		TokenQueryParam: "token",
		TokenCookie:     "auth_token",
		DefaultProvider: "jwt",
		RequireAuth:     true,
		RequireRoles:    []string{},
		RequirePerms:    []string{},
		Metadata:        make(map[string]string),
	}
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authManager interface{}, config *MiddlewareConfig, logger *logrus.Logger) *AuthMiddleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &AuthMiddleware{
		authManager: authManager,
		logger:      logger,
	}
}

// AuthenticationMiddleware provides authentication middleware
func (am *AuthMiddleware) AuthenticationMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for specified paths
			if am.shouldSkipPath(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from request
			token := am.extractToken(r, config)
			if token == "" {
				if config.RequireAuth {
					am.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Validate token (this would use the auth manager)
			// For now, we'll create a mock validation
			userInfo := am.validateToken(token, config.DefaultProvider)
			if userInfo == nil && config.RequireAuth {
				am.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add user info to context
			ctx := am.addUserToContext(r.Context(), userInfo, token, config.DefaultProvider)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorizationMiddleware provides authorization middleware
func (am *AuthMiddleware) AuthorizationMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user info from context
			userID := r.Context().Value(UserIDKey)
			userRoles := r.Context().Value(UserRolesKey)
			userPerms := r.Context().Value(UserPermsKey)

			if userID == nil {
				am.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			// Check required roles
			if len(config.RequireRoles) > 0 {
				if !am.hasRequiredRoles(userRoles, config.RequireRoles) {
					am.writeErrorResponse(w, http.StatusForbidden, "Insufficient roles")
					return
				}
			}

			// Check required permissions
			if len(config.RequirePerms) > 0 {
				if !am.hasRequiredPermissions(userPerms, config.RequirePerms) {
					am.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkipPath checks if the path should be skipped
func (am *AuthMiddleware) shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractToken extracts token from request
func (am *AuthMiddleware) extractToken(r *http.Request, config *MiddlewareConfig) string {
	// Try header first
	if token := r.Header.Get(config.TokenHeader); token != "" {
		// Remove "Bearer " prefix if present
		if strings.HasPrefix(token, "Bearer ") {
			return strings.TrimPrefix(token, "Bearer ")
		}
		return token
	}

	// Try query parameter
	if token := r.URL.Query().Get(config.TokenQueryParam); token != "" {
		return token
	}

	// Try cookie
	if cookie, err := r.Cookie(config.TokenCookie); err == nil {
		return cookie.Value
	}

	return ""
}

// validateToken validates a token (mock implementation)
func (am *AuthMiddleware) validateToken(token, provider string) map[string]interface{} {
	// This is a mock implementation
	// In real implementation, you would use the auth manager to validate the token
	return map[string]interface{}{
		"user_id":     "mock-user-id",
		"email":       "user@example.com",
		"roles":       []string{"user"},
		"permissions": []string{"read", "write"},
		"service_id":  "default-service",
		"context":     map[string]interface{}{},
	}
}

// addUserToContext adds user information to request context
func (am *AuthMiddleware) addUserToContext(ctx context.Context, userInfo map[string]interface{}, token, provider string) context.Context {
	if userInfo == nil {
		return ctx
	}

	// Add user information to context
	ctx = context.WithValue(ctx, UserIDKey, userInfo["user_id"])
	ctx = context.WithValue(ctx, UserEmailKey, userInfo["email"])
	ctx = context.WithValue(ctx, UserRolesKey, userInfo["roles"])
	ctx = context.WithValue(ctx, UserPermsKey, userInfo["permissions"])
	ctx = context.WithValue(ctx, TokenKey, token)
	ctx = context.WithValue(ctx, ProviderKey, provider)
	ctx = context.WithValue(ctx, ServiceIDKey, userInfo["service_id"])
	ctx = context.WithValue(ctx, ContextKey, userInfo["context"])

	return ctx
}

// hasRequiredRoles checks if user has required roles
func (am *AuthMiddleware) hasRequiredRoles(userRoles interface{}, requiredRoles []string) bool {
	if userRoles == nil {
		return false
	}

	roles, ok := userRoles.([]string)
	if !ok {
		return false
	}

	for _, requiredRole := range requiredRoles {
		found := false
		for _, role := range roles {
			if role == requiredRole {
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

// hasRequiredPermissions checks if user has required permissions
func (am *AuthMiddleware) hasRequiredPermissions(userPerms interface{}, requiredPerms []string) bool {
	if userPerms == nil {
		return false
	}

	perms, ok := userPerms.([]string)
	if !ok {
		return false
	}

	for _, requiredPerm := range requiredPerms {
		found := false
		for _, perm := range perms {
			if perm == requiredPerm {
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

// writeErrorResponse writes an error response
func (am *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s", "timestamp": "%s"}`, message, time.Now().Format(time.RFC3339))
}

// GetUserFromContext extracts user information from context
func GetUserFromContext(ctx context.Context) map[string]interface{} {
	userInfo := make(map[string]interface{})

	if userID := ctx.Value(UserIDKey); userID != nil {
		userInfo["user_id"] = userID
	}
	if email := ctx.Value(UserEmailKey); email != nil {
		userInfo["email"] = email
	}
	if roles := ctx.Value(UserRolesKey); roles != nil {
		userInfo["roles"] = roles
	}
	if perms := ctx.Value(UserPermsKey); perms != nil {
		userInfo["permissions"] = perms
	}
	if token := ctx.Value(TokenKey); token != nil {
		userInfo["token"] = token
	}
	if provider := ctx.Value(ProviderKey); provider != nil {
		userInfo["provider"] = provider
	}
	if serviceID := ctx.Value(ServiceIDKey); serviceID != nil {
		userInfo["service_id"] = serviceID
	}
	if context := ctx.Value(ContextKey); context != nil {
		userInfo["context"] = context
	}

	return userInfo
}
