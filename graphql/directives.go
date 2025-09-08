package graphql

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
)

// DirectiveRegistry manages GraphQL directives
type DirectiveRegistry struct {
	directives map[string]graphql.FieldMiddleware
	logger     *logrus.Logger
}

// NewDirectiveRegistry creates a new directive registry
func NewDirectiveRegistry(logger *logrus.Logger) *DirectiveRegistry {
	return &DirectiveRegistry{
		directives: make(map[string]graphql.FieldMiddleware),
		logger:     logger,
	}
}

// Register registers a directive
func (dr *DirectiveRegistry) Register(name string, directive graphql.FieldMiddleware) {
	dr.directives[name] = directive
	dr.logger.WithField("directive", name).Info("GraphQL directive registered")
}

// Get gets a directive by name
func (dr *DirectiveRegistry) Get(name string) (graphql.FieldMiddleware, bool) {
	directive, exists := dr.directives[name]
	return directive, exists
}

// List lists all registered directives
func (dr *DirectiveRegistry) List() []string {
	var names []string
	for name := range dr.directives {
		names = append(names, name)
	}
	return names
}

// AuthDirective provides authentication directive
func AuthDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Check if user is authenticated
		user, exists := GetUserFromContext(ctx)
		if !exists || user == nil {
			logger.Warn("Unauthenticated access attempt")
			return nil, &graphql.Error{
				Message: "Authentication required",
				Extensions: map[string]interface{}{
					"code": "UNAUTHENTICATED",
				},
			}
		}

		return next(ctx)
	}
}

// HasRoleDirective provides role-based access control directive
func HasRoleDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get directive arguments
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Get role from directive arguments
		roleArg, exists := rc.Field.Arguments.ForName("role")
		if !exists {
			return nil, fmt.Errorf("role argument required")
		}

		requiredRole, ok := roleArg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid role argument")
		}

		// Check user role
		user, exists := GetUserFromContext(ctx)
		if !exists || user == nil {
			logger.Warn("Unauthenticated access attempt for role-based directive")
			return nil, &graphql.Error{
				Message: "Authentication required",
				Extensions: map[string]interface{}{
					"code": "UNAUTHENTICATED",
				},
			}
		}

		// Check if user has required role
		// This is a simplified implementation - in real app, check actual user roles
		userRoles := getUserRoles(user) // Implement this function
		if !hasRole(userRoles, requiredRole) {
			logger.WithFields(logrus.Fields{
				"user":          user,
				"required_role": requiredRole,
				"user_roles":    userRoles,
			}).Warn("Insufficient role for access")

			return nil, &graphql.Error{
				Message: "Insufficient permissions",
				Extensions: map[string]interface{}{
					"code": "FORBIDDEN",
				},
			}
		}

		return next(ctx)
	}
}

// HasPermissionDirective provides permission-based access control directive
func HasPermissionDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get directive arguments
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Get permission from directive arguments
		permissionArg, exists := rc.Field.Arguments.ForName("permission")
		if !exists {
			return nil, fmt.Errorf("permission argument required")
		}

		requiredPermission, ok := permissionArg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid permission argument")
		}

		// Check user permissions
		user, exists := GetUserFromContext(ctx)
		if !exists || user == nil {
			logger.Warn("Unauthenticated access attempt for permission-based directive")
			return nil, &graphql.Error{
				Message: "Authentication required",
				Extensions: map[string]interface{}{
					"code": "UNAUTHENTICATED",
				},
			}
		}

		// Check if user has required permission
		userPermissions := getUserPermissions(user) // Implement this function
		if !hasPermission(userPermissions, requiredPermission) {
			logger.WithFields(logrus.Fields{
				"user":                user,
				"required_permission": requiredPermission,
				"user_permissions":    userPermissions,
			}).Warn("Insufficient permission for access")

			return nil, &graphql.Error{
				Message: "Insufficient permissions",
				Extensions: map[string]interface{}{
					"code": "FORBIDDEN",
				},
			}
		}

		return next(ctx)
	}
}

// RateLimitDirective provides rate limiting directive
func RateLimitDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	// Simple in-memory rate limiter (in production, use Redis)
	requests := make(map[string][]int64)

	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get directive arguments
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Get rate limit arguments
		limitArg, exists := rc.Field.Arguments.ForName("limit")
		if !exists {
			return nil, fmt.Errorf("limit argument required")
		}

		limit, ok := limitArg.Value.(int)
		if !ok {
			return nil, fmt.Errorf("invalid limit argument")
		}

		windowArg, exists := rc.Field.Arguments.ForName("window")
		if !exists {
			return nil, fmt.Errorf("window argument required")
		}

		window, ok := windowArg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid window argument")
		}

		// Get client identifier
		clientID := getClientID(ctx)

		// Check rate limit
		now := getCurrentTimestamp()
		windowDuration := parseWindow(window)

		// Clean old requests
		if clientRequests, exists := requests[clientID]; exists {
			var validRequests []int64
			for _, reqTime := range clientRequests {
				if now-reqTime < windowDuration {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientID] = validRequests
		}

		// Check if limit exceeded
		if len(requests[clientID]) >= limit {
			logger.WithFields(logrus.Fields{
				"client_id": clientID,
				"limit":     limit,
				"window":    window,
			}).Warn("Rate limit exceeded")

			return nil, &graphql.Error{
				Message: "Rate limit exceeded",
				Extensions: map[string]interface{}{
					"code": "RATE_LIMITED",
				},
			}
		}

		// Add current request
		requests[clientID] = append(requests[clientID], now)

		return next(ctx)
	}
}

// CacheDirective provides caching directive
func CacheDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	// Simple in-memory cache (in production, use Redis)
	cache := make(map[string]interface{})

	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get directive arguments
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Get cache arguments
		maxAgeArg, exists := rc.Field.Arguments.ForName("maxAge")
		if !exists {
			return nil, fmt.Errorf("maxAge argument required")
		}

		maxAge, ok := maxAgeArg.Value.(int)
		if !ok {
			return nil, fmt.Errorf("invalid maxAge argument")
		}

		// Generate cache key
		cacheKey := generateCacheKey(ctx, rc)

		// Check cache
		if cachedValue, exists := cache[cacheKey]; exists {
			logger.WithField("cache_key", cacheKey).Debug("Cache hit")
			return cachedValue, nil
		}

		// Execute resolver
		result, err := next(ctx)
		if err != nil {
			return nil, err
		}

		// Cache result
		cache[cacheKey] = result
		logger.WithField("cache_key", cacheKey).Debug("Cache miss, value cached")

		return result, nil
	}
}

// DeprecatedDirective provides deprecation directive
func DeprecatedDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get directive arguments
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Get deprecation reason
		reason := "This field is deprecated"
		if reasonArg, exists := rc.Field.Arguments.ForName("reason"); exists {
			if reasonStr, ok := reasonArg.Value.(string); ok {
				reason = reasonStr
			}
		}

		// Log deprecation warning
		logger.WithFields(logrus.Fields{
			"field":  rc.Field.Name,
			"reason": reason,
		}).Warn("Deprecated field accessed")

		return next(ctx)
	}
}

// TenantDirective provides tenant isolation directive
func TenantDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Check if tenant is present in context
		tenant, exists := GetTenantFromContext(ctx)
		if !exists || tenant == "" {
			logger.Warn("Tenant context missing")
			return nil, &graphql.Error{
				Message: "Tenant context required",
				Extensions: map[string]interface{}{
					"code": "TENANT_REQUIRED",
				},
			}
		}

		// Add tenant to field context
		rc := graphql.GetFieldContext(ctx)
		if rc != nil {
			rc.Field.Arguments = append(rc.Field.Arguments, &graphql.Argument{
				Name:  "tenant_id",
				Value: tenant,
			})
		}

		return next(ctx)
	}
}

// ValidationDirective provides input validation directive
func ValidationDirective(logger *logrus.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get field context
		rc := graphql.GetFieldContext(ctx)
		if rc == nil {
			return nil, fmt.Errorf("invalid field context")
		}

		// Validate arguments
		for _, arg := range rc.Field.Arguments {
			if err := validateArgument(arg); err != nil {
				logger.WithFields(logrus.Fields{
					"argument": arg.Name,
					"error":    err.Error(),
				}).Warn("Argument validation failed")

				return nil, &graphql.Error{
					Message: fmt.Sprintf("Invalid argument %s: %s", arg.Name, err.Error()),
					Extensions: map[string]interface{}{
						"code": "VALIDATION_ERROR",
					},
				}
			}
		}

		return next(ctx)
	}
}

// Helper functions

// getUserRoles gets user roles (implement based on your user model)
func getUserRoles(user interface{}) []string {
	// This is a placeholder implementation
	// In a real application, extract roles from user object
	return []string{"user"}
}

// hasRole checks if user has required role
func hasRole(userRoles []string, requiredRole string) bool {
	for _, role := range userRoles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// getUserPermissions gets user permissions (implement based on your user model)
func getUserPermissions(user interface{}) []string {
	// This is a placeholder implementation
	// In a real application, extract permissions from user object
	return []string{"read"}
}

// hasPermission checks if user has required permission
func hasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

// getClientID gets client identifier from context
func getClientID(ctx context.Context) string {
	// Try to get from headers
	rc := graphql.GetRequestContext(ctx)
	if rc != nil {
		if clientIP := rc.Header.Get("X-Forwarded-For"); clientIP != "" {
			return clientIP
		}
		if clientIP := rc.Header.Get("X-Real-IP"); clientIP != "" {
			return clientIP
		}
	}

	// Fallback to user ID
	if user, exists := GetUserFromContext(ctx); exists {
		return fmt.Sprintf("user_%v", user)
	}

	return "anonymous"
}

// getCurrentTimestamp gets current timestamp
func getCurrentTimestamp() int64 {
	// This is a simplified implementation
	// In production, use proper timestamp
	return 0 // Implement based on your needs
}

// parseWindow parses time window string
func parseWindow(window string) int64 {
	// This is a simplified implementation
	// In production, parse window properly (e.g., "1m", "1h", "1d")
	switch window {
	case "1m":
		return 60
	case "1h":
		return 3600
	case "1d":
		return 86400
	default:
		return 60
	}
}

// generateCacheKey generates cache key for field
func generateCacheKey(ctx context.Context, rc *graphql.FieldContext) string {
	var key strings.Builder

	// Add field name
	key.WriteString(rc.Field.Name)

	// Add arguments
	for _, arg := range rc.Field.Arguments {
		key.WriteString(fmt.Sprintf("_%s_%v", arg.Name, arg.Value))
	}

	// Add tenant
	if tenant, exists := GetTenantFromContext(ctx); exists {
		key.WriteString(fmt.Sprintf("_tenant_%s", tenant))
	}

	return key.String()
}

// validateArgument validates a GraphQL argument
func validateArgument(arg *graphql.Argument) error {
	// This is a simplified implementation
	// In production, implement proper validation based on field type

	if arg.Value == nil {
		return fmt.Errorf("argument cannot be null")
	}

	// Add more validation rules as needed
	return nil
}
