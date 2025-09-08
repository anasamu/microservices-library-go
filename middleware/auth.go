package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtSecret string
	logger    *logrus.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecret string, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := m.extractToken(c)
		if err != nil {
			m.logger.WithError(err).Warn("Failed to extract token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": err.Error()})
			c.Abort()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.WithError(err).Warn("Failed to validate token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		// Add to request context for logging
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "User role not found"})
			c.Abort()
			return
		}

		if userRole != requiredRole {
			m.logger.WithFields(logrus.Fields{
				"user_role":     userRole,
				"required_role": requiredRole,
				"user_id":       c.GetString("user_id"),
			}).Warn("Insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "User role not found"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range requiredRoles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.logger.WithFields(logrus.Fields{
				"user_role":      userRole,
				"required_roles": requiredRoles,
				"user_id":        c.GetString("user_id"),
			}).Warn("Insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenant middleware that requires tenant validation
func (m *AuthMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			// Try to get from query parameter
			tenantID = c.Query("tenant_id")
		}

		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request", "message": "Tenant ID is required"})
			c.Abort()
			return
		}

		// Validate tenant ID format
		if _, err := uuid.Parse(tenantID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request", "message": "Invalid tenant ID format"})
			c.Abort()
			return
		}

		// Check if user's tenant matches the requested tenant
		userTenantID := c.GetString("tenant_id")
		if userTenantID != tenantID {
			m.logger.WithFields(logrus.Fields{
				"user_tenant_id":      userTenantID,
				"requested_tenant_id": tenantID,
				"user_id":             c.GetString("user_id"),
			}).Warn("Tenant mismatch")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Tenant mismatch"})
			c.Abort()
			return
		}

		// Set tenant ID in context
		c.Set("requested_tenant_id", tenantID)

		c.Next()
	}
}

// OptionalAuth middleware that optionally validates authentication
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := m.extractToken(c)
		if err != nil {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			m.logger.WithError(err).Debug("Invalid token in optional auth")
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		// Add to request context for logging
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// extractToken extracts JWT token from request
func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// Try to get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// Try to get token from query parameter
	token := c.Query("token")
	if token != "" {
		return token, nil
	}

	// Try to get token from cookie
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie, nil
	}

	return "", &AuthError{Message: "No token provided"}
}

// validateToken validates JWT token and returns claims
func (m *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &AuthError{Message: "Unexpected signing method"}
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return nil, &AuthError{Message: "Invalid token", Cause: err}
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, &AuthError{Message: "Invalid token claims"}
}

// AuthError represents an authentication error
type AuthError struct {
	Message string
	Cause   error
}

func (e *AuthError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// GenerateToken generates a JWT token for a user
func (m *AuthMiddleware) GenerateToken(userID, tenantID, email, role string) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "siakad",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

// GetUserFromContext extracts user information from context
func GetUserFromContext(c *gin.Context) (userID, tenantID, email, role string) {
	userID = c.GetString("user_id")
	tenantID = c.GetString("tenant_id")
	email = c.GetString("user_email")
	role = c.GetString("user_role")
	return
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", &AuthError{Message: "User ID not found in context"}
	}
	return userID, nil
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(c *gin.Context) (string, error) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		return "", &AuthError{Message: "Tenant ID not found in context"}
	}
	return tenantID, nil
}
