package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RBACManager handles Role-Based Access Control
type RBACManager struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	policies    map[string]*Policy
	logger      *logrus.Logger
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

// NewRBACManager creates a new RBAC manager
func NewRBACManager(logger *logrus.Logger) *RBACManager {
	return &RBACManager{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
		policies:    make(map[string]*Policy),
		logger:      logger,
	}
}

// CreateRole creates a new role
func (rm *RBACManager) CreateRole(ctx context.Context, name, description string, permissions []string, metadata map[string]interface{}) (*Role, error) {
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

	rm.roles[name] = role

	rm.logger.WithFields(logrus.Fields{
		"role_id":     roleID,
		"role_name":   name,
		"permissions": permissions,
	}).Info("Role created successfully")

	return role, nil
}

// GetRole retrieves a role by name
func (rm *RBACManager) GetRole(ctx context.Context, name string) (*Role, error) {
	role, exists := rm.roles[name]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", name)
	}
	return role, nil
}

// UpdateRole updates an existing role
func (rm *RBACManager) UpdateRole(ctx context.Context, name string, description string, permissions []string, metadata map[string]interface{}) (*Role, error) {
	role, exists := rm.roles[name]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", name)
	}

	role.Description = description
	role.Permissions = permissions
	role.Metadata = metadata
	role.UpdatedAt = getCurrentTimestamp()

	rm.logger.WithFields(logrus.Fields{
		"role_id":     role.ID,
		"role_name":   name,
		"permissions": permissions,
	}).Info("Role updated successfully")

	return role, nil
}

// DeleteRole deletes a role
func (rm *RBACManager) DeleteRole(ctx context.Context, name string) error {
	_, exists := rm.roles[name]
	if !exists {
		return fmt.Errorf("role not found: %s", name)
	}

	delete(rm.roles, name)

	rm.logger.WithField("role_name", name).Info("Role deleted successfully")
	return nil
}

// CreatePermission creates a new permission
func (rm *RBACManager) CreatePermission(ctx context.Context, name, resource, action, description string, metadata map[string]interface{}) (*Permission, error) {
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

	rm.permissions[name] = permission

	rm.logger.WithFields(logrus.Fields{
		"permission_id": permissionID,
		"permission":    name,
		"resource":      resource,
		"action":        action,
	}).Info("Permission created successfully")

	return permission, nil
}

// GetPermission retrieves a permission by name
func (rm *RBACManager) GetPermission(ctx context.Context, name string) (*Permission, error) {
	permission, exists := rm.permissions[name]
	if !exists {
		return nil, fmt.Errorf("permission not found: %s", name)
	}
	return permission, nil
}

// CreatePolicy creates a new access policy
func (rm *RBACManager) CreatePolicy(ctx context.Context, name, description string, effect PolicyEffect, principal, action, resource string, condition map[string]interface{}, metadata map[string]interface{}) (*Policy, error) {
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

	rm.policies[name] = policy

	rm.logger.WithFields(logrus.Fields{
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
func (rm *RBACManager) GetPolicy(ctx context.Context, name string) (*Policy, error) {
	policy, exists := rm.policies[name]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", name)
	}
	return policy, nil
}

// CheckAccess checks if a principal has access to a resource
func (rm *RBACManager) CheckAccess(ctx context.Context, request *AccessRequest) (*AccessDecision, error) {
	decision := &AccessDecision{
		Decision: PolicyEffectDeny,
		Reason:   "No matching policies found",
		Policies: []string{},
		Context:  request.Context,
	}

	// Check role-based permissions first
	if role, exists := rm.roles[request.Principal]; exists {
		for _, permissionName := range role.Permissions {
			if permission, permExists := rm.permissions[permissionName]; permExists {
				if rm.matchesPermission(permission, request) {
					decision.Decision = PolicyEffectAllow
					decision.Reason = "Role-based permission granted"
					decision.Policies = append(decision.Policies, permissionName)
					return decision, nil
				}
			}
		}
	}

	// Check explicit policies
	for _, policy := range rm.policies {
		if rm.matchesPolicy(policy, request) {
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

	rm.logger.WithFields(logrus.Fields{
		"principal": request.Principal,
		"action":    request.Action,
		"resource":  request.Resource,
		"decision":  decision.Decision,
	}).Debug("Access check completed")

	return decision, nil
}

// matchesPermission checks if a permission matches the access request
func (rm *RBACManager) matchesPermission(permission *Permission, request *AccessRequest) bool {
	// Check resource match
	if !rm.matchesResource(permission.Resource, request.Resource) {
		return false
	}

	// Check action match
	if !rm.matchesAction(permission.Action, request.Action) {
		return false
	}

	return true
}

// matchesPolicy checks if a policy matches the access request
func (rm *RBACManager) matchesPolicy(policy *Policy, request *AccessRequest) bool {
	// Check principal match
	if !rm.matchesPrincipal(policy.Principal, request.Principal) {
		return false
	}

	// Check action match
	if !rm.matchesAction(policy.Action, request.Action) {
		return false
	}

	// Check resource match
	if !rm.matchesResource(policy.Resource, request.Resource) {
		return false
	}

	// Check conditions
	if !rm.evaluateConditions(policy.Condition, request.Context) {
		return false
	}

	return true
}

// matchesPrincipal checks if a principal pattern matches the request principal
func (rm *RBACManager) matchesPrincipal(pattern, principal string) bool {
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
func (rm *RBACManager) matchesAction(pattern, action string) bool {
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
func (rm *RBACManager) matchesResource(pattern, resource string) bool {
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
func (rm *RBACManager) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
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
func (rm *RBACManager) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	var userRoles []*Role

	// In a real implementation, you would query the database
	// For now, we'll return roles that match the user ID pattern
	for _, role := range rm.roles {
		if strings.Contains(role.Name, userID) || role.Name == "default" {
			userRoles = append(userRoles, role)
		}
	}

	return userRoles, nil
}

// GetUserPermissions returns all permissions for a user
func (rm *RBACManager) GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error) {
	var userPermissions []*Permission
	userRoles, err := rm.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]*Permission)

	for _, role := range userRoles {
		for _, permissionName := range role.Permissions {
			if permission, exists := rm.permissions[permissionName]; exists {
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
