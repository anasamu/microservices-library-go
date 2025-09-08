package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthManager is the main authentication manager that coordinates all auth components
type AuthManager struct {
	config        *AuthConfig
	jwtManager    *JWTManager
	oauth2Manager *OAuth2Manager
	rbacManager   *RBACManager
	aclManager    *ACLManager
	abacManager   *ABACManager
	policyEngine  *PolicyEngine
	auditManager  *AuditManager
	middleware    *AuthMiddleware
	logger        *logrus.Logger
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig, logger *logrus.Logger) (*AuthManager, error) {
	if config == nil {
		config = DefaultAuthConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Validate configuration
	if err := ValidateAuthConfig(config); err != nil {
		return nil, fmt.Errorf("invalid auth configuration: %w", err)
	}

	am := &AuthManager{
		config: config,
		logger: logger,
	}

	// Initialize JWT manager
	if config.JWT != nil {
		jwtConfig := &JWTManagerConfig{
			SecretKey:     config.JWT.SecretKey,
			AccessExpiry:  config.JWT.AccessExpiry,
			RefreshExpiry: config.JWT.RefreshExpiry,
			Issuer:        config.JWT.Issuer,
			Audience:      config.JWT.Audience,
			Algorithm:     config.JWT.Algorithm,
			KeyID:         config.JWT.KeyID,
		}
		am.jwtManager = NewJWTManager(jwtConfig, logger)
	}

	// Initialize OAuth2 manager
	if config.OAuth2 != nil {
		am.oauth2Manager = NewOAuth2Manager(logger)

		// Register OAuth2 providers
		for name, providerConfig := range config.OAuth2.Providers {
			if providerConfig.Enabled {
				oauth2Config := &OAuth2ManagerConfig{
					ClientID:     providerConfig.ClientID,
					ClientSecret: providerConfig.ClientSecret,
					RedirectURL:  providerConfig.RedirectURL,
					Scopes:       providerConfig.Scopes,
					AuthURL:      providerConfig.AuthURL,
					TokenURL:     providerConfig.TokenURL,
				}
				if err := am.oauth2Manager.RegisterProvider(name, oauth2Config); err != nil {
					logger.WithError(err).WithField("provider", name).Warn("Failed to register OAuth2 provider")
				}
			}
		}
	}

	// Initialize RBAC manager
	if config.RBAC != nil && config.RBAC.Enabled {
		am.rbacManager = NewRBACManager(logger)

		// Create default roles and permissions
		if err := am.initializeDefaultRBAC(context.Background()); err != nil {
			logger.WithError(err).Warn("Failed to initialize default RBAC")
		}
	}

	// Initialize ACL manager
	if config.ACL != nil && config.ACL.Enabled {
		am.aclManager = NewACLManager(logger)
	}

	// Initialize ABAC manager
	if config.ABAC != nil && config.ABAC.Enabled {
		am.abacManager = NewABACManager(logger)

		// Create default attributes
		if err := am.initializeDefaultABAC(context.Background()); err != nil {
			logger.WithError(err).Warn("Failed to initialize default ABAC")
		}
	}

	// Initialize policy engine
	if am.rbacManager != nil || am.aclManager != nil || am.abacManager != nil {
		am.policyEngine = NewPolicyEngine(am.rbacManager, am.aclManager, am.abacManager, logger)
	}

	// Initialize audit manager
	if config.Audit != nil && config.Audit.Enabled {
		am.auditManager = NewAuditManager(config.Audit, logger)
	}

	// Initialize middleware
	if am.jwtManager != nil && am.policyEngine != nil {
		am.middleware = NewAuthMiddleware(am.jwtManager, am.policyEngine, config.Middleware, logger)
	}

	logger.Info("Authentication manager initialized successfully")
	return am, nil
}

// initializeDefaultRBAC initializes default RBAC roles and permissions
func (am *AuthManager) initializeDefaultRBAC(ctx context.Context) error {
	// Create default permissions
	defaultPermissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"read:own", "user", "read", "Read own user data"},
		{"write:own", "user", "write", "Write own user data"},
		{"read:all", "user", "read", "Read all user data"},
		{"write:all", "user", "write", "Write all user data"},
		{"admin:all", "*", "*", "Full administrative access"},
	}

	for _, perm := range defaultPermissions {
		_, err := am.rbacManager.CreatePermission(ctx, perm.name, perm.resource, perm.action, perm.description, nil)
		if err != nil {
			am.logger.WithError(err).WithField("permission", perm.name).Debug("Permission already exists or creation failed")
		}
	}

	// Create default roles
	defaultRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{"user", "Basic user role", []string{"read:own", "write:own"}},
		{"admin", "Administrator role", []string{"read:all", "write:all", "admin:all"}},
		{"moderator", "Moderator role", []string{"read:all", "write:own"}},
	}

	for _, role := range defaultRoles {
		_, err := am.rbacManager.CreateRole(ctx, role.name, role.description, role.permissions, nil)
		if err != nil {
			am.logger.WithError(err).WithField("role", role.name).Debug("Role already exists or creation failed")
		}
	}

	return nil
}

// initializeDefaultABAC initializes default ABAC attributes
func (am *AuthManager) initializeDefaultABAC(ctx context.Context) error {
	// Create default attributes
	defaultAttributes := []struct {
		name        string
		attrType    AttributeType
		category    AttributeCategory
		description string
		required    bool
	}{
		{"user_id", AttributeTypeString, AttributeCategorySubject, "User identifier", true},
		{"tenant_id", AttributeTypeString, AttributeCategorySubject, "Tenant identifier", false},
		{"role", AttributeTypeString, AttributeCategorySubject, "User role", false},
		{"department", AttributeTypeString, AttributeCategorySubject, "User department", false},
		{"resource_id", AttributeTypeString, AttributeCategoryResource, "Resource identifier", true},
		{"resource_type", AttributeTypeString, AttributeCategoryResource, "Resource type", true},
		{"resource_owner", AttributeTypeString, AttributeCategoryResource, "Resource owner", false},
		{"action_name", AttributeTypeString, AttributeCategoryAction, "Action name", true},
		{"environment", AttributeTypeString, AttributeCategoryEnvironment, "Environment name", false},
		{"ip_address", AttributeTypeString, AttributeCategoryEnvironment, "IP address", false},
		{"time_of_day", AttributeTypeNumber, AttributeCategoryEnvironment, "Time of day (hour)", false},
	}

	for _, attr := range defaultAttributes {
		_, err := am.abacManager.CreateAttribute(ctx, attr.name, attr.attrType, attr.category, attr.description, nil, nil, attr.required, "system", nil)
		if err != nil {
			am.logger.WithError(err).WithField("attribute", attr.name).Debug("Attribute already exists or creation failed")
		}
	}

	return nil
}

// GetJWTManager returns the JWT manager
func (am *AuthManager) GetJWTManager() *JWTManager {
	return am.jwtManager
}

// GetOAuth2Manager returns the OAuth2 manager
func (am *AuthManager) GetOAuth2Manager() *OAuth2Manager {
	return am.oauth2Manager
}

// GetRBACManager returns the RBAC manager
func (am *AuthManager) GetRBACManager() *RBACManager {
	return am.rbacManager
}

// GetACLManager returns the ACL manager
func (am *AuthManager) GetACLManager() *ACLManager {
	return am.aclManager
}

// GetABACManager returns the ABAC manager
func (am *AuthManager) GetABACManager() *ABACManager {
	return am.abacManager
}

// GetPolicyEngine returns the policy engine
func (am *AuthManager) GetPolicyEngine() *PolicyEngine {
	return am.policyEngine
}

// GetAuditManager returns the audit manager
func (am *AuthManager) GetAuditManager() *AuditManager {
	return am.auditManager
}

// GetMiddleware returns the auth middleware
func (am *AuthManager) GetMiddleware() *AuthMiddleware {
	return am.middleware
}

// AuthenticateUser authenticates a user and returns tokens
func (am *AuthManager) AuthenticateUser(ctx context.Context, userID, tenantID uuid.UUID, email string, roles, permissions []string, metadata map[string]interface{}) (*TokenPair, error) {
	// Generate token pair
	tokenPair, err := am.jwtManager.GenerateTokenPair(ctx, userID, tenantID, email, roles, permissions, metadata)
	if err != nil {
		am.logger.WithError(err).WithField("user_id", userID).Error("Failed to generate token pair")
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Log authentication event
	if am.auditManager != nil {
		am.auditManager.LogTokenGeneration(ctx, userID.String(), tenantID.String(), "", "", AuditResultSuccess, "Token generated successfully", map[string]interface{}{
			"roles":       roles,
			"permissions": permissions,
		})
	}

	return tokenPair, nil
}

// ValidateUserToken validates a user token and returns claims
func (am *AuthManager) ValidateUserToken(ctx context.Context, token string) (*JWTClaims, error) {
	claims, err := am.jwtManager.ValidateToken(ctx, token)
	if err != nil {
		am.logger.WithError(err).Debug("Token validation failed")

		// Log failed validation
		if am.auditManager != nil {
			am.auditManager.LogTokenValidation(ctx, "", "", "", "", AuditResultFailure, err.Error(), nil)
		}

		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Log successful validation
	if am.auditManager != nil {
		am.auditManager.LogTokenValidation(ctx, claims.UserID.String(), claims.TenantID.String(), "", "", AuditResultSuccess, "Token validated successfully", nil)
	}

	return claims, nil
}

// CheckUserAccess checks if a user has access to a resource
func (am *AuthManager) CheckUserAccess(ctx context.Context, userID, tenantID, resource, action string, context map[string]interface{}) (*PolicyDecision, error) {
	if am.policyEngine == nil {
		return nil, fmt.Errorf("policy engine not available")
	}

	policyRequest := &PolicyRequest{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resource,
		Action:   action,
		Context:  context,
	}

	decision, err := am.policyEngine.EvaluatePolicy(ctx, policyRequest, am.config.Middleware.PolicyEvaluationOrder)
	if err != nil {
		am.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"resource": resource,
			"action":   action,
		}).Error("Policy evaluation failed")
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	// Log access decision
	if am.auditManager != nil {
		result := AuditResultSuccess
		if decision.Decision == PolicyEngineEffectDeny {
			result = AuditResultFailure
		}

		am.auditManager.LogAccessDecision(ctx, userID, tenantID, resource, action, "", result, decision.Reason, map[string]interface{}{
			"source":    decision.Source,
			"policy_id": decision.PolicyID,
		})
	}

	return decision, nil
}

// AssignUserRole assigns a role to a user
func (am *AuthManager) AssignUserRole(ctx context.Context, userID, roleName, assignedBy string) error {
	if am.rbacManager == nil {
		return fmt.Errorf("RBAC manager not available")
	}

	// This would typically involve database operations
	// For now, we'll just log the event
	am.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"role_name":   roleName,
		"assigned_by": assignedBy,
	}).Info("User role assigned")

	// Log role assignment
	if am.auditManager != nil {
		am.auditManager.LogRoleAssignment(ctx, userID, "", roleName, assignedBy, "", AuditResultSuccess, "Role assigned successfully", nil)
	}

	return nil
}

// RevokeUserRole revokes a role from a user
func (am *AuthManager) RevokeUserRole(ctx context.Context, userID, roleName, revokedBy string) error {
	if am.rbacManager == nil {
		return fmt.Errorf("RBAC manager not available")
	}

	// This would typically involve database operations
	// For now, we'll just log the event
	am.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"role_name":  roleName,
		"revoked_by": revokedBy,
	}).Info("User role revoked")

	// Log role revocation
	if am.auditManager != nil {
		am.auditManager.LogRoleAssignment(ctx, userID, "", roleName, revokedBy, "", AuditResultFailure, "Role revoked", nil)
	}

	return nil
}

// GetUserRoles returns all roles for a user
func (am *AuthManager) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	if am.rbacManager == nil {
		return nil, fmt.Errorf("RBAC manager not available")
	}

	return am.rbacManager.GetUserRoles(ctx, userID)
}

// GetUserPermissions returns all permissions for a user
func (am *AuthManager) GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error) {
	if am.rbacManager == nil {
		return nil, fmt.Errorf("RBAC manager not available")
	}

	return am.rbacManager.GetUserPermissions(ctx, userID)
}

// GetConfiguration returns the current configuration
func (am *AuthManager) GetConfiguration() *AuthConfig {
	return am.config
}

// UpdateConfiguration updates the configuration
func (am *AuthManager) UpdateConfiguration(newConfig *AuthConfig) error {
	if err := ValidateAuthConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	am.config = newConfig
	am.logger.Info("Authentication configuration updated")
	return nil
}

// GetStatistics returns authentication statistics
func (am *AuthManager) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get policy engine statistics
	if am.policyEngine != nil {
		policyStats, err := am.policyEngine.GetPolicyStatistics(ctx)
		if err == nil {
			stats["policies"] = policyStats
		}
	}

	// Add configuration info
	stats["configuration"] = map[string]interface{}{
		"jwt_enabled":        am.jwtManager != nil,
		"oauth2_enabled":     am.oauth2Manager != nil,
		"rbac_enabled":       am.rbacManager != nil,
		"acl_enabled":        am.aclManager != nil,
		"abac_enabled":       am.abacManager != nil,
		"audit_enabled":      am.auditManager != nil,
		"middleware_enabled": am.middleware != nil,
	}

	return stats, nil
}

// Shutdown gracefully shuts down the authentication manager
func (am *AuthManager) Shutdown(ctx context.Context) error {
	am.logger.Info("Shutting down authentication manager")

	// Stop audit manager
	if am.auditManager != nil {
		am.auditManager.Stop()
	}

	am.logger.Info("Authentication manager shutdown complete")
	return nil
}
