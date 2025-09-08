package interfaces

import (
	"context"
	"time"
)

// AuthManager interface defines authentication operations
type AuthManager interface {
	// Token operations
	GenerateToken(ctx context.Context, claims TokenClaims) (string, error)
	GenerateTokenPair(ctx context.Context, claims TokenClaims) (TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (TokenPair, error)
	RevokeToken(ctx context.Context, token string) error

	// User authentication
	Authenticate(ctx context.Context, credentials Credentials) (*User, error)
	Logout(ctx context.Context, userID string) error

	// Session management
	CreateSession(ctx context.Context, userID string, metadata map[string]interface{}) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	RefreshSession(ctx context.Context, sessionID string) (*Session, error)
}

// RBACManager interface defines role-based access control operations
type RBACManager interface {
	// Role operations
	CreateRole(ctx context.Context, role Role) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	UpdateRole(ctx context.Context, role Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context) ([]Role, error)

	// Permission operations
	CreatePermission(ctx context.Context, permission Permission) error
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	UpdatePermission(ctx context.Context, permission Permission) error
	DeletePermission(ctx context.Context, permissionID string) error
	ListPermissions(ctx context.Context) ([]Permission, error)

	// User-Role assignments
	AssignRole(ctx context.Context, userID, roleID string) error
	UnassignRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]User, error)

	// Permission checks
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
	HasRole(ctx context.Context, userID, roleName string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]Permission, error)

	// Role-Permission assignments
	GrantPermission(ctx context.Context, roleID, permissionID string) error
	RevokePermission(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID      string                 `json:"user_id"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	IssuedAt    time.Time              `json:"iat"`
	ExpiresAt   time.Time              `json:"exp"`
	Issuer      string                 `json:"iss"`
	Audience    string                 `json:"aud"`
	Subject     string                 `json:"sub"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// Credentials represents user login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id,omitempty"`
}

// User represents a user in the system
type User struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expires_at"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	IsActive  bool                   `json:"is_active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Role represents a role in the system
type Role struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OAuth2Provider interface defines OAuth2 operations
type OAuth2Provider interface {
	// Provider information
	GetProviderName() string
	GetAuthURL(ctx context.Context, state string) (string, error)
	GetToken(ctx context.Context, code string) (*OAuth2Token, error)
	GetUserInfo(ctx context.Context, token *OAuth2Token) (*OAuth2User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OAuth2Token, error)
	RevokeToken(ctx context.Context, token string) error
}

// OAuth2Token represents OAuth2 token response
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// OAuth2User represents OAuth2 user information
type OAuth2User struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Picture       string                 `json:"picture"`
	Locale        string                 `json:"locale"`
	VerifiedEmail bool                   `json:"verified_email"`
	Provider      string                 `json:"provider"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
