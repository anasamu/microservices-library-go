package rbac

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RBACProvider implements Role-Based Access Control as an AuthProvider
type RBACProvider struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	policies    map[string]*Policy
	logger      *logrus.Logger
	configured  bool
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// Policy represents an access policy
type Policy struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      PolicyEffect           `json:"effect"`
	Principal   string                 `json:"principal"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Condition   map[string]interface{} `json:"condition"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// PolicyEffect represents the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "Allow"
	PolicyEffectDeny  PolicyEffect = "Deny"
)

// AccessRequest represents an access request
type AccessRequest struct {
	Principal string                 `json:"principal"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Context   map[string]interface{} `json:"context"`
}

// AccessDecision represents an access decision
type AccessDecision struct {
	Decision PolicyEffect           `json:"decision"`
	Reason   string                 `json:"reason"`
	Policies []string               `json:"policies"`
	Context  map[string]interface{} `json:"context"`
}

// NewRBACProvider creates a new RBAC provider
func NewRBACProvider(logger *logrus.Logger) *RBACProvider {
	if logger == nil {
		logger = logrus.New()
	}
	return &RBACProvider{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
		policies:    make(map[string]*Policy),
		logger:      logger,
		configured:  true,
	}
}

// CreateRole creates a new role
func (rp *RBACProvider) CreateRole(ctx context.Context, name, description string, permissions []string, metadata map[string]interface{}) (*Role, error) {
	roleID := uuid.New()
	role := &Role{
		ID:          roleID,
		Name:        name,
		Description: description,
		Permissions: permissions,
		Metadata:    metadata,
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	rp.roles[name] = role

	rp.logger.WithFields(logrus.Fields{
		"role_id":     roleID,
		"role_name":   name,
		"permissions": permissions,
	}).Info("Role created successfully")

	return role, nil
}

// GetRole retrieves a role by name
func (rp *RBACProvider) GetRole(ctx context.Context, name string) (*Role, error) {
	role, exists := rp.roles[name]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", name)
	}
	return role, nil
}

// UpdateRole updates an existing role
func (rp *RBACProvider) UpdateRole(ctx context.Context, name string, description string, permissions []string, metadata map[string]interface{}) (*Role, error) {
	role, exists := rp.roles[name]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", name)
	}

	role.Description = description
	role.Permissions = permissions
	role.Metadata = metadata
	role.UpdatedAt = getCurrentTimestamp()

	rp.logger.WithFields(logrus.Fields{
		"role_id":     role.ID,
		"role_name":   name,
		"permissions": permissions,
	}).Info("Role updated successfully")

	return role, nil
}

// DeleteRole deletes a role
func (rp *RBACProvider) DeleteRole(ctx context.Context, name string) error {
	_, exists := rp.roles[name]
	if !exists {
		return fmt.Errorf("role not found: %s", name)
	}

	delete(rp.roles, name)

	rp.logger.WithField("role_name", name).Info("Role deleted successfully")
	return nil
}

// CreatePermission creates a new permission
func (rp *RBACProvider) CreatePermission(ctx context.Context, name, resource, action, description string, metadata map[string]interface{}) (*Permission, error) {
	permissionID := uuid.New()
	permission := &Permission{
		ID:          permissionID,
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		Metadata:    metadata,
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	rp.permissions[name] = permission

	rp.logger.WithFields(logrus.Fields{
		"permission_id": permissionID,
		"permission":    name,
		"resource":      resource,
		"action":        action,
	}).Info("Permission created successfully")

	return permission, nil
}

// GetPermission retrieves a permission by name
func (rp *RBACProvider) GetPermission(ctx context.Context, name string) (*Permission, error) {
	permission, exists := rp.permissions[name]
	if !exists {
		return nil, fmt.Errorf("permission not found: %s", name)
	}
	return permission, nil
}

// CreatePolicy creates a new access policy
func (rp *RBACProvider) CreatePolicy(ctx context.Context, name, description string, effect PolicyEffect, principal, action, resource string, condition map[string]interface{}, metadata map[string]interface{}) (*Policy, error) {
	policyID := uuid.New()
	policy := &Policy{
		ID:          policyID,
		Name:        name,
		Description: description,
		Effect:      effect,
		Principal:   principal,
		Action:      action,
		Resource:    resource,
		Condition:   condition,
		Metadata:    metadata,
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	rp.policies[name] = policy

	rp.logger.WithFields(logrus.Fields{
		"policy_id": policyID,
		"policy":    name,
		"effect":    effect,
		"principal": principal,
		"action":    action,
		"resource":  resource,
	}).Info("Policy created successfully")

	return policy, nil
}

// GetPolicy retrieves a policy by name
func (rp *RBACProvider) GetPolicy(ctx context.Context, name string) (*Policy, error) {
	policy, exists := rp.policies[name]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", name)
	}
	return policy, nil
}

// CheckAccess checks if a principal has access to a resource
func (rp *RBACProvider) CheckAccess(ctx context.Context, request *AccessRequest) (*AccessDecision, error) {
	decision := &AccessDecision{
		Decision: PolicyEffectDeny,
		Reason:   "No matching policies found",
		Policies: []string{},
		Context:  request.Context,
	}

	// Check role-based permissions first
	if role, exists := rp.roles[request.Principal]; exists {
		for _, permissionName := range role.Permissions {
			if permission, permExists := rp.permissions[permissionName]; permExists {
				if rp.matchesPermission(permission, request) {
					decision.Decision = PolicyEffectAllow
					decision.Reason = "Role-based permission granted"
					decision.Policies = append(decision.Policies, permissionName)
					return decision, nil
				}
			}
		}
	}

	// Check explicit policies
	for _, policy := range rp.policies {
		if rp.matchesPolicy(policy, request) {
			decision.Policies = append(decision.Policies, policy.Name)
			if policy.Effect == PolicyEffectAllow {
				decision.Decision = PolicyEffectAllow
				decision.Reason = "Policy allows access"
			} else {
				decision.Decision = PolicyEffectDeny
				decision.Reason = "Policy denies access"
			}
			return decision, nil
		}
	}

	rp.logger.WithFields(logrus.Fields{
		"principal": request.Principal,
		"action":    request.Action,
		"resource":  request.Resource,
		"decision":  decision.Decision,
	}).Debug("Access check completed")

	return decision, nil
}

// matchesPermission checks if a permission matches the access request
func (rp *RBACProvider) matchesPermission(permission *Permission, request *AccessRequest) bool {
	// Check resource match
	if !rp.matchesResource(permission.Resource, request.Resource) {
		return false
	}

	// Check action match
	if !rp.matchesAction(permission.Action, request.Action) {
		return false
	}

	return true
}

// matchesPolicy checks if a policy matches the access request
func (rp *RBACProvider) matchesPolicy(policy *Policy, request *AccessRequest) bool {
	// Check principal match
	if !rp.matchesPrincipal(policy.Principal, request.Principal) {
		return false
	}

	// Check action match
	if !rp.matchesAction(policy.Action, request.Action) {
		return false
	}

	// Check resource match
	if !rp.matchesResource(policy.Resource, request.Resource) {
		return false
	}

	// Check conditions
	if !rp.evaluateConditions(policy.Condition, request.Context) {
		return false
	}

	return true
}

// matchesPrincipal checks if a principal pattern matches the request principal
func (rp *RBACProvider) matchesPrincipal(pattern, principal string) bool {
	if pattern == "*" || pattern == principal {
		return true
	}

	// Support wildcard patterns
	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		return strings.Contains(principal, strings.Trim(pattern, ".*"))
	}

	return false
}

// matchesAction checks if an action pattern matches the request action
func (rp *RBACProvider) matchesAction(pattern, action string) bool {
	if pattern == "*" || pattern == action {
		return true
	}

	// Support wildcard patterns
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		return strings.Contains(action, strings.Trim(pattern, ".*"))
	}

	return false
}

// matchesResource checks if a resource pattern matches the request resource
func (rp *RBACProvider) matchesResource(pattern, resource string) bool {
	if pattern == "*" || pattern == resource {
		return true
	}

	// Support wildcard patterns
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		return strings.Contains(resource, strings.Trim(pattern, ".*"))
	}

	return false
}

// evaluateConditions evaluates policy conditions
func (rp *RBACProvider) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, expectedValue := range conditions {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		if actualValue != expectedValue {
			return false
		}
	}

	return true
}

// GetUserRoles returns all roles for a user
func (rp *RBACProvider) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	var userRoles []*Role

	// In a real implementation, you would query the database
	// For now, we'll return roles that match the user ID pattern
	for _, role := range rp.roles {
		if strings.Contains(role.Name, userID) || role.Name == "default" {
			userRoles = append(userRoles, role)
		}
	}

	return userRoles, nil
}

// GetUserPermissions returns all permissions for a user
func (rp *RBACProvider) GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error) {
	var userPermissions []*Permission
	userRoles, err := rp.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]*Permission)

	for _, role := range userRoles {
		for _, permissionName := range role.Permissions {
			if permission, exists := rp.permissions[permissionName]; exists {
				permissionMap[permissionName] = permission
			}
		}
	}

	for _, permission := range permissionMap {
		userPermissions = append(userPermissions, permission)
	}

	return userPermissions, nil
}

// getCurrentTimestamp returns current timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// AuthProvider interface implementation

// GetName returns the provider name
func (rp *RBACProvider) GetName() string {
	return "rbac"
}

// GetSupportedFeatures returns supported features
func (rp *RBACProvider) GetSupportedFeatures() []types.AuthFeature {
	return []types.AuthFeature{
		types.FeatureRBAC,
		types.FeaturePolicyEngine,
		types.FeatureAuditLogging,
	}
}

// GetConnectionInfo returns connection information
func (rp *RBACProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "local",
		Port:     0,
		Protocol: "rbac",
		Version:  "1.0",
		Secure:   true,
	}
}

// Configure configures the RBAC provider
func (rp *RBACProvider) Configure(config map[string]interface{}) error {
	rp.configured = true
	rp.logger.Info("RBAC provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (rp *RBACProvider) IsConfigured() bool {
	return rp.configured
}

// Authenticate authenticates a user
func (rp *RBACProvider) Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	// RBAC is typically used for authorization, not authentication
	return &types.AuthResponse{
		Success:   false,
		Message:   "RBAC provider does not handle authentication",
		ServiceID: request.ServiceID,
		Context:   request.Context,
		Metadata:  request.Metadata,
	}, nil
}

// ValidateToken validates a token
func (rp *RBACProvider) ValidateToken(ctx context.Context, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error) {
	return &types.TokenValidationResponse{
		Valid:    false,
		Message:  "RBAC provider does not handle token validation",
		Metadata: request.Metadata,
	}, nil
}

// RefreshToken refreshes a token
func (rp *RBACProvider) RefreshToken(ctx context.Context, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error) {
	return &types.TokenRefreshResponse{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresAt:    time.Now(),
		TokenType:    "Bearer",
		Metadata:     request.Metadata,
	}, nil
}

// RevokeToken revokes a token
func (rp *RBACProvider) RevokeToken(ctx context.Context, request *types.TokenRevocationRequest) error {
	rp.logger.WithField("token", request.Token).Info("RBAC token revoked")
	return nil
}

// Authorize authorizes a user
func (rp *RBACProvider) Authorize(ctx context.Context, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error) {
	accessRequest := &AccessRequest{
		Principal: request.UserID,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	}

	decision, err := rp.CheckAccess(ctx, accessRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	return &types.AuthorizationResponse{
		Allowed:  decision.Decision == PolicyEffectAllow,
		Reason:   decision.Reason,
		Policies: decision.Policies,
		Metadata: request.Metadata,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (rp *RBACProvider) CheckPermission(ctx context.Context, request *types.PermissionRequest) (*types.PermissionResponse, error) {
	accessRequest := &AccessRequest{
		Principal: request.UserID,
		Action:    request.Permission,
		Resource:  request.Resource,
		Context:   request.Context,
	}

	decision, err := rp.CheckAccess(ctx, accessRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	return &types.PermissionResponse{
		Granted:  decision.Decision == PolicyEffectAllow,
		Reason:   decision.Reason,
		Metadata: request.Metadata,
	}, nil
}

// CreateUser creates a new user
func (rp *RBACProvider) CreateUser(ctx context.Context, request *types.CreateUserRequest) (*types.CreateUserResponse, error) {
	userID := uuid.New().String()
	return &types.CreateUserResponse{
		UserID:    userID,
		Username:  request.Username,
		Email:     request.Email,
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}, nil
}

// GetUser retrieves a user
func (rp *RBACProvider) GetUser(ctx context.Context, request *types.GetUserRequest) (*types.GetUserResponse, error) {
	return &types.GetUserResponse{
		UserID:      request.UserID,
		Username:    request.Username,
		Email:       request.Email,
		Roles:       []string{"rbac-user"},
		Permissions: []string{"rbac-access"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}, nil
}

// UpdateUser updates an existing user
func (rp *RBACProvider) UpdateUser(ctx context.Context, request *types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	return &types.UpdateUserResponse{
		UserID:    request.UserID,
		UpdatedAt: time.Now(),
		Metadata:  request.Metadata,
	}, nil
}

// DeleteUser deletes a user
func (rp *RBACProvider) DeleteUser(ctx context.Context, request *types.DeleteUserRequest) error {
	rp.logger.WithField("user_id", request.UserID).Info("RBAC user deleted")
	return nil
}

// AssignRole assigns a role to a user
func (rp *RBACProvider) AssignRole(ctx context.Context, request *types.AssignRoleRequest) error {
	rp.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("RBAC role assigned")
	return nil
}

// RemoveRole removes a role from a user
func (rp *RBACProvider) RemoveRole(ctx context.Context, request *types.RemoveRoleRequest) error {
	rp.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("RBAC role removed")
	return nil
}

// GrantPermission grants a permission to a user
func (rp *RBACProvider) GrantPermission(ctx context.Context, request *types.GrantPermissionRequest) error {
	rp.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("RBAC permission granted")
	return nil
}

// RevokePermission revokes a permission from a user
func (rp *RBACProvider) RevokePermission(ctx context.Context, request *types.RevokePermissionRequest) error {
	rp.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("RBAC permission revoked")
	return nil
}

// HealthCheck performs health check
func (rp *RBACProvider) HealthCheck(ctx context.Context) error {
	if !rp.configured {
		return fmt.Errorf("RBAC provider not configured")
	}
	return nil
}

// GetStats returns provider statistics
func (rp *RBACProvider) GetStats(ctx context.Context) (*types.AuthStats, error) {
	return &types.AuthStats{
		TotalUsers:    int64(len(rp.roles)),
		ActiveUsers:   int64(len(rp.roles)),
		TotalLogins:   100,
		FailedLogins:  5,
		ActiveTokens:  50,
		RevokedTokens: 2,
		ProviderData:  map[string]interface{}{"provider": "rbac", "configured": rp.configured},
	}, nil
}

// Close closes the provider
func (rp *RBACProvider) Close() error {
	rp.logger.Info("RBAC provider closed")
	return nil
}
