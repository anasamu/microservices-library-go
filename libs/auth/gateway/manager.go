package gateway

import (
	"context"
	"fmt"
	"time"

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
	GetSupportedFeatures() []AuthFeature
	GetConnectionInfo() *ConnectionInfo

	// Authentication operations
	Authenticate(ctx context.Context, request *AuthRequest) (*AuthResponse, error)
	ValidateToken(ctx context.Context, request *TokenValidationRequest) (*TokenValidationResponse, error)
	RefreshToken(ctx context.Context, request *TokenRefreshRequest) (*TokenRefreshResponse, error)
	RevokeToken(ctx context.Context, request *TokenRevocationRequest) error

	// Authorization operations
	Authorize(ctx context.Context, request *AuthorizationRequest) (*AuthorizationResponse, error)
	CheckPermission(ctx context.Context, request *PermissionRequest) (*PermissionResponse, error)

	// User management
	CreateUser(ctx context.Context, request *CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx context.Context, request *GetUserRequest) (*GetUserResponse, error)
	UpdateUser(ctx context.Context, request *UpdateUserRequest) (*UpdateUserResponse, error)
	DeleteUser(ctx context.Context, request *DeleteUserRequest) error

	// Role and permission management
	AssignRole(ctx context.Context, request *AssignRoleRequest) error
	RemoveRole(ctx context.Context, request *RemoveRoleRequest) error
	GrantPermission(ctx context.Context, request *GrantPermissionRequest) error
	RevokePermission(ctx context.Context, request *RevokePermissionRequest) error

	// Health and monitoring
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*AuthStats, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// AuthFeature represents an authentication/authorization feature
type AuthFeature string

const (
	// Authentication features
	FeatureJWT               AuthFeature = "jwt"
	FeatureOAuth2            AuthFeature = "oauth2"
	FeatureTwoFactor         AuthFeature = "two_factor"
	FeaturePasswordReset     AuthFeature = "password_reset"
	FeatureAccountLockout    AuthFeature = "account_lockout"
	FeatureSessionManagement AuthFeature = "session_management"
	FeatureSSO               AuthFeature = "sso"
	FeatureLDAP              AuthFeature = "ldap"
	FeatureSAML              AuthFeature = "saml"
	FeatureOpenIDConnect     AuthFeature = "openid_connect"

	// Authorization features
	FeatureRBAC            AuthFeature = "rbac"
	FeatureABAC            AuthFeature = "abac"
	FeatureACL             AuthFeature = "acl"
	FeaturePolicyEngine    AuthFeature = "policy_engine"
	FeatureAttributeBased  AuthFeature = "attribute_based"
	FeatureContextAware    AuthFeature = "context_aware"
	FeatureDynamicPolicies AuthFeature = "dynamic_policies"
	FeatureAuditLogging    AuthFeature = "audit_logging"

	// Security features
	FeatureEncryption           AuthFeature = "encryption"
	FeatureTokenBlacklist       AuthFeature = "token_blacklist"
	FeatureRateLimiting         AuthFeature = "rate_limiting"
	FeatureBruteForceProtection AuthFeature = "brute_force_protection"
	FeatureDeviceManagement     AuthFeature = "device_management"
	FeatureGeolocation          AuthFeature = "geolocation"
	FeatureRiskAssessment       AuthFeature = "risk_assessment"
)

// ConnectionInfo represents auth provider connection information
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
	Secure   bool   `json:"secure"`
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Username      string                 `json:"username"`
	Password      string                 `json:"password"`
	Email         string                 `json:"email"`
	Token         string                 `json:"token,omitempty"`
	TwoFactorCode string                 `json:"two_factor_code,omitempty"`
	DeviceID      string                 `json:"device_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	ServiceID     string                 `json:"service_id,omitempty"` // Dynamic service identifier
	Context       map[string]interface{} `json:"context,omitempty"`    // Dynamic context for multi-tenant scenarios
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success      bool                   `json:"success"`
	UserID       string                 `json:"user_id"`
	AccessToken  string                 `json:"access_token,omitempty"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time              `json:"expires_at,omitempty"`
	TokenType    string                 `json:"token_type,omitempty"`
	Roles        []string               `json:"roles,omitempty"`
	Permissions  []string               `json:"permissions,omitempty"`
	Requires2FA  bool                   `json:"requires_2fa,omitempty"`
	ServiceID    string                 `json:"service_id,omitempty"` // Dynamic service identifier
	Context      map[string]interface{} `json:"context,omitempty"`    // Dynamic context for multi-tenant scenarios
	Message      string                 `json:"message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TokenValidationRequest represents a token validation request
type TokenValidationRequest struct {
	Token     string                 `json:"token"`
	TokenType string                 `json:"token_type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TokenValidationResponse represents a token validation response
type TokenValidationResponse struct {
	Valid     bool                   `json:"valid"`
	UserID    string                 `json:"user_id,omitempty"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
	ExpiresAt time.Time              `json:"expires_at,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TokenRefreshRequest represents a token refresh request
type TokenRefreshRequest struct {
	RefreshToken string                 `json:"refresh_token"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TokenRefreshResponse represents a token refresh response
type TokenRefreshResponse struct {
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time              `json:"expires_at"`
	TokenType    string                 `json:"token_type"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TokenRevocationRequest represents a token revocation request
type TokenRevocationRequest struct {
	Token    string                 `json:"token"`
	UserID   string                 `json:"user_id,omitempty"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuthorizationRequest represents an authorization request
type AuthorizationRequest struct {
	UserID      string                 `json:"user_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuthorizationResponse represents an authorization response
type AuthorizationResponse struct {
	Allowed  bool                   `json:"allowed"`
	Reason   string                 `json:"reason,omitempty"`
	Policies []string               `json:"policies,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionRequest represents a permission check request
type PermissionRequest struct {
	UserID     string                 `json:"user_id"`
	Permission string                 `json:"permission"`
	Resource   string                 `json:"resource,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionResponse represents a permission check response
type PermissionResponse struct {
	Granted  bool                   `json:"granted"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateUserRequest represents a create user request
type CreateUserRequest struct {
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Password    string                 `json:"password"`
	FirstName   string                 `json:"first_name,omitempty"`
	LastName    string                 `json:"last_name,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateUserResponse represents a create user response
type CreateUserResponse struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// GetUserRequest represents a get user request
type GetUserRequest struct {
	UserID   string                 `json:"user_id,omitempty"`
	Username string                 `json:"username,omitempty"`
	Email    string                 `json:"email,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetUserResponse represents a get user response
type GetUserResponse struct {
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	FirstName   string                 `json:"first_name,omitempty"`
	LastName    string                 `json:"last_name,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Password   string                 `json:"password,omitempty"`
	FirstName  string                 `json:"first_name,omitempty"`
	LastName   string                 `json:"last_name,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserResponse represents an update user response
type UpdateUserResponse struct {
	UserID    string                 `json:"user_id"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DeleteUserRequest represents a delete user request
type DeleteUserRequest struct {
	UserID   string                 `json:"user_id"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AssignRoleRequest represents an assign role request
type AssignRoleRequest struct {
	UserID   string                 `json:"user_id"`
	Role     string                 `json:"role"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RemoveRoleRequest represents a remove role request
type RemoveRoleRequest struct {
	UserID   string                 `json:"user_id"`
	Role     string                 `json:"role"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GrantPermissionRequest represents a grant permission request
type GrantPermissionRequest struct {
	UserID     string                 `json:"user_id"`
	Permission string                 `json:"permission"`
	Resource   string                 `json:"resource,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RevokePermissionRequest represents a revoke permission request
type RevokePermissionRequest struct {
	UserID     string                 `json:"user_id"`
	Permission string                 `json:"permission"`
	Resource   string                 `json:"resource,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuthStats represents authentication statistics
type AuthStats struct {
	TotalUsers    int64                  `json:"total_users"`
	ActiveUsers   int64                  `json:"active_users"`
	TotalLogins   int64                  `json:"total_logins"`
	FailedLogins  int64                  `json:"failed_logins"`
	ActiveTokens  int64                  `json:"active_tokens"`
	RevokedTokens int64                  `json:"revoked_tokens"`
	ProviderData  map[string]interface{} `json:"provider_data"`
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
func (am *AuthManager) Authenticate(ctx context.Context, providerName string, request *AuthRequest) (*AuthResponse, error) {
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
func (am *AuthManager) ValidateToken(ctx context.Context, providerName string, request *TokenValidationRequest) (*TokenValidationResponse, error) {
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
func (am *AuthManager) RefreshToken(ctx context.Context, providerName string, request *TokenRefreshRequest) (*TokenRefreshResponse, error) {
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
func (am *AuthManager) RevokeToken(ctx context.Context, providerName string, request *TokenRevocationRequest) error {
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
func (am *AuthManager) Authorize(ctx context.Context, providerName string, request *AuthorizationRequest) (*AuthorizationResponse, error) {
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
func (am *AuthManager) CheckPermission(ctx context.Context, providerName string, request *PermissionRequest) (*PermissionResponse, error) {
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
func (am *AuthManager) CreateUser(ctx context.Context, providerName string, request *CreateUserRequest) (*CreateUserResponse, error) {
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
func (am *AuthManager) GetUser(ctx context.Context, providerName string, request *GetUserRequest) (*GetUserResponse, error) {
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
func (am *AuthManager) UpdateUser(ctx context.Context, providerName string, request *UpdateUserRequest) (*UpdateUserResponse, error) {
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
func (am *AuthManager) DeleteUser(ctx context.Context, providerName string, request *DeleteUserRequest) error {
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
func (am *AuthManager) AssignRole(ctx context.Context, providerName string, request *AssignRoleRequest) error {
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
func (am *AuthManager) RemoveRole(ctx context.Context, providerName string, request *RemoveRoleRequest) error {
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
func (am *AuthManager) GrantPermission(ctx context.Context, providerName string, request *GrantPermissionRequest) error {
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
func (am *AuthManager) RevokePermission(ctx context.Context, providerName string, request *RevokePermissionRequest) error {
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
func (am *AuthManager) validateAuthRequest(request *AuthRequest) error {
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
