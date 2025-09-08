package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/sirupsen/logrus"
)

// AuthManager manages multiple authentication and authorization providers
type AuthManager struct {
	providers map[string]AuthProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds auth manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	Metadata        map[string]string `json:"metadata"`
}

// AuthProvider interface for authentication and authorization backends
type AuthProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.AuthFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Authentication operations
	Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error)
	ValidateToken(ctx context.Context, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error)
	RefreshToken(ctx context.Context, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error)
	RevokeToken(ctx context.Context, request *types.TokenRevocationRequest) error

	// Authorization operations
	Authorize(ctx context.Context, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error)
	CheckPermission(ctx context.Context, request *types.PermissionRequest) (*types.PermissionResponse, error)

	// User management
	CreateUser(ctx context.Context, request *types.CreateUserRequest) (*types.CreateUserResponse, error)
	GetUser(ctx context.Context, request *types.GetUserRequest) (*types.GetUserResponse, error)
	UpdateUser(ctx context.Context, request *types.UpdateUserRequest) (*types.UpdateUserResponse, error)
	DeleteUser(ctx context.Context, request *types.DeleteUserRequest) error

	// Role and permission management
	AssignRole(ctx context.Context, request *types.AssignRoleRequest) error
	RemoveRole(ctx context.Context, request *types.RemoveRoleRequest) error
	GrantPermission(ctx context.Context, request *types.GrantPermissionRequest) error
	RevokePermission(ctx context.Context, request *types.RevokePermissionRequest) error

	// Health and monitoring
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*types.AuthStats, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// DefaultManagerConfig returns default auth manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "jwt",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		Metadata:        make(map[string]string),
	}
}

// NewAuthManager creates a new auth manager
func NewAuthManager(config *ManagerConfig, logger *logrus.Logger) *AuthManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &AuthManager{
		providers: make(map[string]AuthProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers an auth provider
func (am *AuthManager) RegisterProvider(provider AuthProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	am.providers[name] = provider
	am.logger.WithField("provider", name).Info("Auth provider registered")

	return nil
}

// GetProvider returns an auth provider by name
func (am *AuthManager) GetProvider(name string) (AuthProvider, error) {
	provider, exists := am.providers[name]
	if !exists {
		return nil, fmt.Errorf("auth provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default auth provider
func (am *AuthManager) GetDefaultProvider() (AuthProvider, error) {
	return am.GetProvider(am.config.DefaultProvider)
}

// Authenticate authenticates a user using the specified provider
func (am *AuthManager) Authenticate(ctx context.Context, providerName string, request *types.AuthRequest) (*types.AuthResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := am.validateAuthRequest(request); err != nil {
		return nil, fmt.Errorf("invalid auth request: %w", err)
	}

	response, err := provider.Authenticate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  response.UserID,
		"success":  response.Success,
	}).Debug("Authentication completed")

	return response, nil
}

// ValidateToken validates a token using the specified provider
func (am *AuthManager) ValidateToken(ctx context.Context, providerName string, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.ValidateToken(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"valid":    response.Valid,
		"user_id":  response.UserID,
	}).Debug("Token validation completed")

	return response, nil
}

// RefreshToken refreshes a token using the specified provider
func (am *AuthManager) RefreshToken(ctx context.Context, providerName string, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.RefreshToken(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"expires_at": response.ExpiresAt,
	}).Debug("Token refresh completed")

	return response, nil
}

// RevokeToken revokes a token using the specified provider
func (am *AuthManager) RevokeToken(ctx context.Context, providerName string, request *types.TokenRevocationRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.RevokeToken(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"token":    request.Token,
	}).Debug("Token revocation completed")

	return nil
}

// Authorize authorizes a user using the specified provider
func (am *AuthManager) Authorize(ctx context.Context, providerName string, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.Authorize(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  request.UserID,
		"resource": request.Resource,
		"action":   request.Action,
		"allowed":  response.Allowed,
	}).Debug("Authorization completed")

	return response, nil
}

// CheckPermission checks if a user has a specific permission
func (am *AuthManager) CheckPermission(ctx context.Context, providerName string, request *types.PermissionRequest) (*types.PermissionResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.CheckPermission(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"user_id":    request.UserID,
		"permission": request.Permission,
		"granted":    response.Granted,
	}).Debug("Permission check completed")

	return response, nil
}

// CreateUser creates a new user
func (am *AuthManager) CreateUser(ctx context.Context, providerName string, request *types.CreateUserRequest) (*types.CreateUserResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.CreateUser(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  response.UserID,
		"username": response.Username,
		"email":    response.Email,
	}).Info("User created successfully")

	return response, nil
}

// GetUser retrieves a user
func (am *AuthManager) GetUser(ctx context.Context, providerName string, request *types.GetUserRequest) (*types.GetUserResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.GetUser(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  response.UserID,
		"username": response.Username,
	}).Debug("User retrieved successfully")

	return response, nil
}

// UpdateUser updates an existing user
func (am *AuthManager) UpdateUser(ctx context.Context, providerName string, request *types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.UpdateUser(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  response.UserID,
	}).Info("User updated successfully")

	return response, nil
}

// DeleteUser deletes a user
func (am *AuthManager) DeleteUser(ctx context.Context, providerName string, request *types.DeleteUserRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.DeleteUser(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  request.UserID,
	}).Info("User deleted successfully")

	return nil
}

// AssignRole assigns a role to a user
func (am *AuthManager) AssignRole(ctx context.Context, providerName string, request *types.AssignRoleRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.AssignRole(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  request.UserID,
		"role":     request.Role,
	}).Info("Role assigned successfully")

	return nil
}

// RemoveRole removes a role from a user
func (am *AuthManager) RemoveRole(ctx context.Context, providerName string, request *types.RemoveRoleRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.RemoveRole(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  request.UserID,
		"role":     request.Role,
	}).Info("Role removed successfully")

	return nil
}

// GrantPermission grants a permission to a user
func (am *AuthManager) GrantPermission(ctx context.Context, providerName string, request *types.GrantPermissionRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.GrantPermission(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("Permission granted successfully")

	return nil
}

// RevokePermission revokes a permission from a user
func (am *AuthManager) RevokePermission(ctx context.Context, providerName string, request *types.RevokePermissionRequest) error {
	provider, err := am.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.RevokePermission(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	am.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("Permission revoked successfully")

	return nil
}

// validateAuthRequest validates an authentication request
func (am *AuthManager) validateAuthRequest(request *types.AuthRequest) error {
	if request == nil {
		return fmt.Errorf("auth request cannot be nil")
	}

	if request.Username == "" && request.Email == "" {
		return fmt.Errorf("username or email is required")
	}

	if request.Password == "" && request.Token == "" {
		return fmt.Errorf("password or token is required")
	}

	return nil
}

// HealthCheck performs health check on all providers
func (am *AuthManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range am.providers {
		results[name] = provider.HealthCheck(ctx)
	}

	return results
}

// GetStats returns statistics for all providers
func (am *AuthManager) GetStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	for name, provider := range am.providers {
		if providerStats, err := provider.GetStats(ctx); err == nil {
			stats[name] = providerStats
		}
	}

	return stats
}

// Close closes all providers
func (am *AuthManager) Close() error {
	var errors []error

	for name, provider := range am.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	am.logger.Info("All auth providers closed")
	return nil
}
