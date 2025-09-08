package auth

import (
	"time"
)

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
