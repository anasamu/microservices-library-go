package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AuthConfig holds comprehensive authentication configuration
type AuthConfig struct {
	JWT        *JWTConfig        `json:"jwt"`
	OAuth2     *OAuth2Config     `json:"oauth2"`
	RBAC       *RBACConfig       `json:"rbac"`
	ACL        *ACLConfig        `json:"acl"`
	ABAC       *ABACConfig       `json:"abac"`
	Middleware *MiddlewareConfig `json:"middleware"`
	Security   *SecurityConfig   `json:"security"`
	Audit      *AuditConfig      `json:"audit"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string             `json:"secret_key"`
	AccessExpiry  time.Duration      `json:"access_expiry"`
	RefreshExpiry time.Duration      `json:"refresh_expiry"`
	Issuer        string             `json:"issuer"`
	Audience      string             `json:"audience"`
	Algorithm     string             `json:"algorithm"`
	KeyID         string             `json:"key_id"`
	KeyRotation   *KeyRotationConfig `json:"key_rotation,omitempty"`
}

// KeyRotationConfig holds key rotation configuration
type KeyRotationConfig struct {
	Enabled          bool          `json:"enabled"`
	RotationInterval time.Duration `json:"rotation_interval"`
	KeyHistorySize   int           `json:"key_history_size"`
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	Providers     map[string]*OAuth2ProviderConfig `json:"providers"`
	DefaultScopes []string                         `json:"default_scopes"`
	StateTimeout  time.Duration                    `json:"state_timeout"`
}

// OAuth2ProviderConfig holds OAuth2 provider configuration
type OAuth2ProviderConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	Enabled      bool     `json:"enabled"`
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	Enabled            bool                `json:"enabled"`
	DefaultRoles       []string            `json:"default_roles"`
	DefaultPermissions []string            `json:"default_permissions"`
	RoleHierarchy      map[string][]string `json:"role_hierarchy"`
	PermissionCache    *CacheConfig        `json:"permission_cache,omitempty"`
}

// ACLConfig holds ACL configuration
type ACLConfig struct {
	Enabled       bool         `json:"enabled"`
	DefaultEffect string       `json:"default_effect"`
	MaxEntries    int          `json:"max_entries"`
	EntryCache    *CacheConfig `json:"entry_cache,omitempty"`
}

// ABACConfig holds ABAC configuration
type ABACConfig struct {
	Enabled        bool         `json:"enabled"`
	DefaultEffect  string       `json:"default_effect"`
	MaxPolicies    int          `json:"max_policies"`
	PolicyCache    *CacheConfig `json:"policy_cache,omitempty"`
	AttributeCache *CacheConfig `json:"attribute_cache,omitempty"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled         bool          `json:"enabled"`
	TTL             time.Duration `json:"ttl"`
	MaxSize         int           `json:"max_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	PasswordPolicy       *PasswordPolicyConfig       `json:"password_policy,omitempty"`
	SessionSecurity      *SessionSecurityConfig      `json:"session_security,omitempty"`
	RateLimiting         *RateLimitingConfig         `json:"rate_limiting,omitempty"`
	BruteForceProtection *BruteForceProtectionConfig `json:"brute_force_protection,omitempty"`
}

// PasswordPolicyConfig holds password policy configuration
type PasswordPolicyConfig struct {
	MinLength        int           `json:"min_length"`
	RequireUppercase bool          `json:"require_uppercase"`
	RequireLowercase bool          `json:"require_lowercase"`
	RequireNumbers   bool          `json:"require_numbers"`
	RequireSymbols   bool          `json:"require_symbols"`
	MaxAge           time.Duration `json:"max_age"`
	HistorySize      int           `json:"history_size"`
}

// SessionSecurityConfig holds session security configuration
type SessionSecurityConfig struct {
	SecureCookies         bool          `json:"secure_cookies"`
	HttpOnlyCookies       bool          `json:"http_only_cookies"`
	SameSitePolicy        string        `json:"same_site_policy"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool          `json:"enabled"`
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// BruteForceProtectionConfig holds brute force protection configuration
type BruteForceProtectionConfig struct {
	Enabled           bool          `json:"enabled"`
	MaxFailedAttempts int           `json:"max_failed_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
	ResetWindow       time.Duration `json:"reset_window"`
}

// AuditConfig holds audit configuration
type AuditConfig struct {
	Enabled       bool                   `json:"enabled"`
	LogLevel      string                 `json:"log_level"`
	LogEvents     []string               `json:"log_events"`
	RetentionDays int                    `json:"retention_days"`
	StorageType   string                 `json:"storage_type"` // file, database, external
	StorageConfig map[string]interface{} `json:"storage_config,omitempty"`
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWT: &JWTConfig{
			SecretKey:     "default-secret-key-change-in-production",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
			Issuer:        "microservices-library",
			Audience:      "microservices-clients",
			Algorithm:     "HS256",
			KeyID:         "default-key-id",
			KeyRotation: &KeyRotationConfig{
				Enabled:          false,
				RotationInterval: 24 * time.Hour,
				KeyHistorySize:   5,
			},
		},
		OAuth2: &OAuth2Config{
			Providers:     make(map[string]*OAuth2ProviderConfig),
			DefaultScopes: []string{"openid", "profile", "email"},
			StateTimeout:  10 * time.Minute,
		},
		RBAC: &RBACConfig{
			Enabled:            true,
			DefaultRoles:       []string{"user"},
			DefaultPermissions: []string{"read:own"},
			RoleHierarchy:      make(map[string][]string),
			PermissionCache: &CacheConfig{
				Enabled:         true,
				TTL:             5 * time.Minute,
				MaxSize:         1000,
				CleanupInterval: 1 * time.Hour,
			},
		},
		ACL: &ACLConfig{
			Enabled:       true,
			DefaultEffect: "deny",
			MaxEntries:    10000,
			EntryCache: &CacheConfig{
				Enabled:         true,
				TTL:             5 * time.Minute,
				MaxSize:         1000,
				CleanupInterval: 1 * time.Hour,
			},
		},
		ABAC: &ABACConfig{
			Enabled:       true,
			DefaultEffect: "deny",
			MaxPolicies:   1000,
			PolicyCache: &CacheConfig{
				Enabled:         true,
				TTL:             5 * time.Minute,
				MaxSize:         500,
				CleanupInterval: 1 * time.Hour,
			},
			AttributeCache: &CacheConfig{
				Enabled:         true,
				TTL:             10 * time.Minute,
				MaxSize:         200,
				CleanupInterval: 1 * time.Hour,
			},
		},
		Middleware: DefaultMiddlewareConfig(),
		Security: &SecurityConfig{
			PasswordPolicy: &PasswordPolicyConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
				MaxAge:           90 * 24 * time.Hour,
				HistorySize:      5,
			},
			SessionSecurity: &SessionSecurityConfig{
				SecureCookies:         true,
				HttpOnlyCookies:       true,
				SameSitePolicy:        "strict",
				SessionTimeout:        30 * time.Minute,
				MaxConcurrentSessions: 5,
			},
			RateLimiting: &RateLimitingConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				BurstSize:         10,
				WindowSize:        1 * time.Minute,
			},
			BruteForceProtection: &BruteForceProtectionConfig{
				Enabled:           true,
				MaxFailedAttempts: 5,
				LockoutDuration:   15 * time.Minute,
				ResetWindow:       1 * time.Hour,
			},
		},
		Audit: &AuditConfig{
			Enabled:       true,
			LogLevel:      "info",
			LogEvents:     []string{"login", "logout", "token_generation", "token_validation", "access_denied"},
			RetentionDays: 30,
			StorageType:   "file",
			StorageConfig: map[string]interface{}{
				"file_path": "/var/log/auth/audit.log",
			},
		},
	}
}

// LoadAuthConfigFromFile loads authentication configuration from file
func LoadAuthConfigFromFile(filePath string) (*AuthConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadAuthConfigFromEnv loads authentication configuration from environment variables
func LoadAuthConfigFromEnv() *AuthConfig {
	config := DefaultAuthConfig()

	// JWT Configuration
	if val := os.Getenv("AUTH_JWT_SECRET_KEY"); val != "" {
		config.JWT.SecretKey = val
	}
	if val := os.Getenv("AUTH_JWT_ACCESS_EXPIRY"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.JWT.AccessExpiry = duration
		}
	}
	if val := os.Getenv("AUTH_JWT_REFRESH_EXPIRY"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.JWT.RefreshExpiry = duration
		}
	}
	if val := os.Getenv("AUTH_JWT_ISSUER"); val != "" {
		config.JWT.Issuer = val
	}
	if val := os.Getenv("AUTH_JWT_AUDIENCE"); val != "" {
		config.JWT.Audience = val
	}
	if val := os.Getenv("AUTH_JWT_ALGORITHM"); val != "" {
		config.JWT.Algorithm = val
	}
	if val := os.Getenv("AUTH_JWT_KEY_ID"); val != "" {
		config.JWT.KeyID = val
	}

	// OAuth2 Configuration
	if val := os.Getenv("AUTH_OAUTH2_DEFAULT_SCOPES"); val != "" {
		config.OAuth2.DefaultScopes = strings.Split(val, ",")
	}
	if val := os.Getenv("AUTH_OAUTH2_STATE_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.OAuth2.StateTimeout = duration
		}
	}

	// RBAC Configuration
	if val := os.Getenv("AUTH_RBAC_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.RBAC.Enabled = enabled
		}
	}
	if val := os.Getenv("AUTH_RBAC_DEFAULT_ROLES"); val != "" {
		config.RBAC.DefaultRoles = strings.Split(val, ",")
	}
	if val := os.Getenv("AUTH_RBAC_DEFAULT_PERMISSIONS"); val != "" {
		config.RBAC.DefaultPermissions = strings.Split(val, ",")
	}

	// ACL Configuration
	if val := os.Getenv("AUTH_ACL_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.ACL.Enabled = enabled
		}
	}
	if val := os.Getenv("AUTH_ACL_DEFAULT_EFFECT"); val != "" {
		config.ACL.DefaultEffect = val
	}
	if val := os.Getenv("AUTH_ACL_MAX_ENTRIES"); val != "" {
		if maxEntries, err := strconv.Atoi(val); err == nil {
			config.ACL.MaxEntries = maxEntries
		}
	}

	// ABAC Configuration
	if val := os.Getenv("AUTH_ABAC_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.ABAC.Enabled = enabled
		}
	}
	if val := os.Getenv("AUTH_ABAC_DEFAULT_EFFECT"); val != "" {
		config.ABAC.DefaultEffect = val
	}
	if val := os.Getenv("AUTH_ABAC_MAX_POLICIES"); val != "" {
		if maxPolicies, err := strconv.Atoi(val); err == nil {
			config.ABAC.MaxPolicies = maxPolicies
		}
	}

	// Security Configuration
	if val := os.Getenv("AUTH_SECURITY_PASSWORD_MIN_LENGTH"); val != "" {
		if minLength, err := strconv.Atoi(val); err == nil {
			config.Security.PasswordPolicy.MinLength = minLength
		}
	}
	if val := os.Getenv("AUTH_SECURITY_SESSION_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.Security.SessionSecurity.SessionTimeout = duration
		}
	}
	if val := os.Getenv("AUTH_SECURITY_RATE_LIMIT_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Security.RateLimiting.Enabled = enabled
		}
	}
	if val := os.Getenv("AUTH_SECURITY_RATE_LIMIT_REQUESTS_PER_MINUTE"); val != "" {
		if requests, err := strconv.Atoi(val); err == nil {
			config.Security.RateLimiting.RequestsPerMinute = requests
		}
	}

	// Audit Configuration
	if val := os.Getenv("AUTH_AUDIT_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Audit.Enabled = enabled
		}
	}
	if val := os.Getenv("AUTH_AUDIT_LOG_LEVEL"); val != "" {
		config.Audit.LogLevel = val
	}
	if val := os.Getenv("AUTH_AUDIT_LOG_EVENTS"); val != "" {
		config.Audit.LogEvents = strings.Split(val, ",")
	}
	if val := os.Getenv("AUTH_AUDIT_RETENTION_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			config.Audit.RetentionDays = days
		}
	}

	return config
}

// SaveAuthConfigToFile saves authentication configuration to file
func SaveAuthConfigToFile(config *AuthConfig, filePath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateAuthConfig validates authentication configuration
func ValidateAuthConfig(config *AuthConfig) error {
	var errors []string

	// Validate JWT config
	if config.JWT == nil {
		errors = append(errors, "JWT configuration is required")
	} else {
		if config.JWT.SecretKey == "" {
			errors = append(errors, "JWT secret key is required")
		}
		if config.JWT.AccessExpiry <= 0 {
			errors = append(errors, "JWT access expiry must be positive")
		}
		if config.JWT.RefreshExpiry <= 0 {
			errors = append(errors, "JWT refresh expiry must be positive")
		}
		if config.JWT.Issuer == "" {
			errors = append(errors, "JWT issuer is required")
		}
		if config.JWT.Audience == "" {
			errors = append(errors, "JWT audience is required")
		}
	}

	// Validate security config
	if config.Security != nil {
		if config.Security.PasswordPolicy != nil {
			if config.Security.PasswordPolicy.MinLength < 6 {
				errors = append(errors, "password minimum length must be at least 6")
			}
		}
		if config.Security.SessionSecurity != nil {
			if config.Security.SessionSecurity.SessionTimeout <= 0 {
				errors = append(errors, "session timeout must be positive")
			}
		}
		if config.Security.RateLimiting != nil {
			if config.Security.RateLimiting.RequestsPerMinute <= 0 {
				errors = append(errors, "rate limit requests per minute must be positive")
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// MergeAuthConfig merges two authentication configurations
func MergeAuthConfig(base, override *AuthConfig) *AuthConfig {
	// This is a simplified merge - in a real implementation, you'd want more sophisticated merging
	merged := *base

	if override.JWT != nil {
		merged.JWT = override.JWT
	}
	if override.OAuth2 != nil {
		merged.OAuth2 = override.OAuth2
	}
	if override.RBAC != nil {
		merged.RBAC = override.RBAC
	}
	if override.ACL != nil {
		merged.ACL = override.ACL
	}
	if override.ABAC != nil {
		merged.ABAC = override.ABAC
	}
	if override.Middleware != nil {
		merged.Middleware = override.Middleware
	}
	if override.Security != nil {
		merged.Security = override.Security
	}
	if override.Audit != nil {
		merged.Audit = override.Audit
	}

	return &merged
}
