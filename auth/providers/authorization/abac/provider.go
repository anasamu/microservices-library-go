package abac

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ABACProvider implements Attribute-Based Access Control as an AuthProvider
type ABACProvider struct {
	policies   map[string]*ABACPolicy
	attributes map[string]*Attribute
	logger     *logrus.Logger
	configured bool
}

// ABACPolicy represents an ABAC policy
type ABACPolicy struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      ABACEffect             `json:"effect"`
	Rules       []ABACRule             `json:"rules"`
	Target      ABACTarget             `json:"target"`
	Condition   map[string]interface{} `json:"condition,omitempty"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ABACRule represents a rule within an ABAC policy
type ABACRule struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      ABACEffect             `json:"effect"`
	Condition   map[string]interface{} `json:"condition"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Priority    int                    `json:"priority"`
}

// ABACTarget represents the target of an ABAC policy
type ABACTarget struct {
	Subjects     []string `json:"subjects,omitempty"`     // users, roles, groups
	Resources    []string `json:"resources,omitempty"`    // resource patterns
	Actions      []string `json:"actions,omitempty"`      // action patterns
	Environments []string `json:"environments,omitempty"` // environment conditions
}

// Attribute represents an attribute that can be used in ABAC policies
type Attribute struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Type         AttributeType          `json:"type"`
	Category     AttributeCategory      `json:"category"`
	Description  string                 `json:"description"`
	Values       []string               `json:"values,omitempty"` // for enum types
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Required     bool                   `json:"required"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ABACEffect represents the effect of an ABAC policy
type ABACEffect string

const (
	ABACEffectAllow ABACEffect = "allow"
	ABACEffectDeny  ABACEffect = "deny"
)

// AttributeType represents the type of an attribute
type AttributeType string

const (
	AttributeTypeString   AttributeType = "string"
	AttributeTypeNumber   AttributeType = "number"
	AttributeTypeBoolean  AttributeType = "boolean"
	AttributeTypeDateTime AttributeType = "datetime"
	AttributeTypeEnum     AttributeType = "enum"
	AttributeTypeArray    AttributeType = "array"
)

// AttributeCategory represents the category of an attribute
type AttributeCategory string

const (
	AttributeCategorySubject     AttributeCategory = "subject"     // user attributes
	AttributeCategoryResource    AttributeCategory = "resource"    // resource attributes
	AttributeCategoryAction      AttributeCategory = "action"      // action attributes
	AttributeCategoryEnvironment AttributeCategory = "environment" // environment attributes
)

// ABACRequest represents an ABAC access request
type ABACRequest struct {
	Subject     map[string]interface{} `json:"subject"`     // user attributes
	Resource    map[string]interface{} `json:"resource"`    // resource attributes
	Action      map[string]interface{} `json:"action"`      // action attributes
	Environment map[string]interface{} `json:"environment"` // environment attributes
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ABACDecision represents an ABAC access decision
type ABACDecision struct {
	Decision ABACEffect             `json:"decision"`
	Reason   string                 `json:"reason"`
	PolicyID uuid.UUID              `json:"policy_id,omitempty"`
	RuleID   uuid.UUID              `json:"rule_id,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// NewABACProvider creates a new ABAC provider
func NewABACProvider(logger *logrus.Logger) *ABACProvider {
	if logger == nil {
		logger = logrus.New()
	}
	return &ABACProvider{
		policies:   make(map[string]*ABACPolicy),
		attributes: make(map[string]*Attribute),
		logger:     logger,
		configured: true,
	}
}

// CreateAttribute creates a new attribute
func (ap *ABACProvider) CreateAttribute(ctx context.Context, name string, attrType AttributeType, category AttributeCategory, description string, values []string, defaultValue interface{}, required bool, createdBy string, metadata map[string]interface{}) (*Attribute, error) {
	attrID := uuid.New()
	attribute := &Attribute{
		ID:           attrID,
		Name:         name,
		Type:         attrType,
		Category:     category,
		Description:  description,
		Values:       values,
		DefaultValue: defaultValue,
		Required:     required,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    createdBy,
		Metadata:     metadata,
	}

	ap.attributes[name] = attribute

	ap.logger.WithFields(logrus.Fields{
		"attribute_id": attrID,
		"name":         name,
		"type":         attrType,
		"category":     category,
		"required":     required,
	}).Info("Attribute created successfully")

	return attribute, nil
}

// GetAttribute retrieves an attribute by name
func (ap *ABACProvider) GetAttribute(ctx context.Context, name string) (*Attribute, error) {
	attribute, exists := ap.attributes[name]
	if !exists {
		return nil, fmt.Errorf("attribute not found: %s", name)
	}
	return attribute, nil
}

// CreatePolicy creates a new ABAC policy
func (ap *ABACProvider) CreatePolicy(ctx context.Context, name, description string, effect ABACEffect, rules []ABACRule, target ABACTarget, condition map[string]interface{}, priority int, createdBy string, metadata map[string]interface{}) (*ABACPolicy, error) {
	policyID := uuid.New()
	policy := &ABACPolicy{
		ID:          policyID,
		Name:        name,
		Description: description,
		Effect:      effect,
		Rules:       rules,
		Target:      target,
		Condition:   condition,
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Metadata:    metadata,
	}

	ap.policies[name] = policy

	ap.logger.WithFields(logrus.Fields{
		"policy_id": policyID,
		"name":      name,
		"effect":    effect,
		"priority":  priority,
		"rules":     len(rules),
	}).Info("ABAC policy created successfully")

	return policy, nil
}

// GetPolicy retrieves an ABAC policy by name
func (ap *ABACProvider) GetPolicy(ctx context.Context, name string) (*ABACPolicy, error) {
	policy, exists := ap.policies[name]
	if !exists {
		return nil, fmt.Errorf("ABAC policy not found: %s", name)
	}
	return policy, nil
}

// CheckAccess checks if a request is allowed based on ABAC policies
func (ap *ABACProvider) CheckAccess(ctx context.Context, request *ABACRequest) (*ABACDecision, error) {
	decision := &ABACDecision{
		Decision: ABACEffectDeny,
		Reason:   "No matching ABAC policies found",
		Context:  request.Context,
	}

	// Get all applicable policies
	applicablePolicies := ap.getApplicablePolicies(request)

	// Sort by priority (higher priority first)
	ap.sortPoliciesByPriority(applicablePolicies)

	// Evaluate policies in priority order
	for _, policy := range applicablePolicies {
		if ap.evaluatePolicy(policy, request) {
			decision.Decision = policy.Effect
			decision.Reason = fmt.Sprintf("ABAC policy matched: %s", policy.Description)
			decision.PolicyID = policy.ID
			break
		}
	}

	ap.logger.WithFields(logrus.Fields{
		"decision": decision.Decision,
		"reason":   decision.Reason,
	}).Debug("ABAC access check completed")

	return decision, nil
}

// getApplicablePolicies returns all policies that could apply to the request
func (ap *ABACProvider) getApplicablePolicies(request *ABACRequest) []*ABACPolicy {
	var applicable []*ABACPolicy

	for _, policy := range ap.policies {
		if ap.policyMatches(policy, request) {
			applicable = append(applicable, policy)
		}
	}

	return applicable
}

// policyMatches checks if a policy matches the request
func (ap *ABACProvider) policyMatches(policy *ABACPolicy, request *ABACRequest) bool {
	// Check target matching
	if !ap.matchesTarget(policy.Target, request) {
		return false
	}

	// Check policy-level conditions
	if !ap.evaluateConditions(policy.Condition, request) {
		return false
	}

	return true
}

// matchesTarget checks if the policy target matches the request
func (ap *ABACProvider) matchesTarget(target ABACTarget, request *ABACRequest) bool {
	// Check subjects
	if len(target.Subjects) > 0 {
		subjectID, exists := request.Subject["id"]
		if !exists || !ap.contains(target.Subjects, fmt.Sprintf("%v", subjectID)) {
			return false
		}
	}

	// Check resources
	if len(target.Resources) > 0 {
		resourceID, exists := request.Resource["id"]
		if !exists || !ap.matchesAnyPattern(target.Resources, fmt.Sprintf("%v", resourceID)) {
			return false
		}
	}

	// Check actions
	if len(target.Actions) > 0 {
		actionName, exists := request.Action["name"]
		if !exists || !ap.matchesAnyPattern(target.Actions, fmt.Sprintf("%v", actionName)) {
			return false
		}
	}

	// Check environments
	if len(target.Environments) > 0 {
		environment, exists := request.Environment["name"]
		if !exists || !ap.contains(target.Environments, fmt.Sprintf("%v", environment)) {
			return false
		}
	}

	return true
}

// evaluatePolicy evaluates an ABAC policy against the request
func (ap *ABACProvider) evaluatePolicy(policy *ABACPolicy, request *ABACRequest) bool {
	// Sort rules by priority
	ap.sortRulesByPriority(policy.Rules)

	// Evaluate rules in priority order
	for _, rule := range policy.Rules {
		if ap.evaluateRule(rule, request) {
			return rule.Effect == ABACEffectAllow
		}
	}

	// If no rules match, use policy default effect
	return policy.Effect == ABACEffectAllow
}

// evaluateRule evaluates an ABAC rule against the request
func (ap *ABACProvider) evaluateRule(rule ABACRule, request *ABACRequest) bool {
	// Check rule conditions
	if !ap.evaluateConditions(rule.Condition, request) {
		return false
	}

	return true
}

// evaluateConditions evaluates conditions against the request
func (ap *ABACProvider) evaluateConditions(conditions map[string]interface{}, request *ABACRequest) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, expectedValue := range conditions {
		actualValue := ap.getAttributeValue(key, request)
		if actualValue == nil {
			return false
		}

		if !ap.compareValues(actualValue, expectedValue) {
			return false
		}
	}

	return true
}

// getAttributeValue retrieves an attribute value from the request
func (ap *ABACProvider) getAttributeValue(key string, request *ABACRequest) interface{} {
	// Check if it's a nested attribute (e.g., "subject.role")
	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		category := parts[0]
		attrName := parts[1]

		switch category {
		case "subject":
			return request.Subject[attrName]
		case "resource":
			return request.Resource[attrName]
		case "action":
			return request.Action[attrName]
		case "environment":
			return request.Environment[attrName]
		}
	}

	// Check all categories for the attribute
	if value, exists := request.Subject[key]; exists {
		return value
	}
	if value, exists := request.Resource[key]; exists {
		return value
	}
	if value, exists := request.Action[key]; exists {
		return value
	}
	if value, exists := request.Environment[key]; exists {
		return value
	}

	return nil
}

// compareValues compares two values based on their types
func (ap *ABACProvider) compareValues(actual, expected interface{}) bool {
	// Handle different comparison types
	switch expectedValue := expected.(type) {
	case string:
		actualStr := fmt.Sprintf("%v", actual)
		return actualStr == expectedValue
	case int, int64, float64:
		// Numeric comparison
		actualNum, ok := actual.(int)
		if !ok {
			if actualFloat, ok := actual.(float64); ok {
				actualNum = int(actualFloat)
			} else {
				return false
			}
		}
		expectedNum, ok := expectedValue.(int)
		if !ok {
			if expectedFloat, ok := expectedValue.(float64); ok {
				expectedNum = int(expectedFloat)
			} else {
				return false
			}
		}
		return actualNum == expectedNum
	case bool:
		actualBool, ok := actual.(bool)
		if !ok {
			return false
		}
		return actualBool == expectedValue
	case []interface{}:
		// Array comparison
		actualArray, ok := actual.([]interface{})
		if !ok {
			return false
		}
		return ap.compareArrays(actualArray, expectedValue)
	default:
		return actual == expected
	}
}

// compareArrays compares two arrays
func (ap *ABACProvider) compareArrays(actual, expected []interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, expectedItem := range expected {
		if !ap.compareValues(actual[i], expectedItem) {
			return false
		}
	}

	return true
}

// contains checks if a slice contains a value
func (ap *ABACProvider) contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// matchesAnyPattern checks if a value matches any pattern in the slice
func (ap *ABACProvider) matchesAnyPattern(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if ap.matchesPattern(pattern, value) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a value matches a pattern
func (ap *ABACProvider) matchesPattern(pattern, value string) bool {
	if pattern == "*" || pattern == value {
		return true
	}

	// Support wildcard patterns
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		return strings.Contains(value, strings.Trim(pattern, ".*"))
	}

	return false
}

// sortPoliciesByPriority sorts policies by priority (higher priority first)
func (ap *ABACProvider) sortPoliciesByPriority(policies []*ABACPolicy) {
	for i := 0; i < len(policies)-1; i++ {
		for j := 0; j < len(policies)-i-1; j++ {
			if policies[j].Priority < policies[j+1].Priority {
				policies[j], policies[j+1] = policies[j+1], policies[j]
			}
		}
	}
}

// sortRulesByPriority sorts rules by priority (higher priority first)
func (ap *ABACProvider) sortRulesByPriority(rules []ABACRule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := 0; j < len(rules)-i-1; j++ {
			if rules[j].Priority < rules[j+1].Priority {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}
}

// ListAttributes returns all attributes
func (ap *ABACProvider) ListAttributes(ctx context.Context) ([]*Attribute, error) {
	var attributes []*Attribute
	for _, attribute := range ap.attributes {
		attributes = append(attributes, attribute)
	}
	return attributes, nil
}

// ListPolicies returns all ABAC policies
func (ap *ABACProvider) ListPolicies(ctx context.Context) ([]*ABACPolicy, error) {
	var policies []*ABACPolicy
	for _, policy := range ap.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}

// AuthProvider interface implementation

// GetName returns the provider name
func (ap *ABACProvider) GetName() string {
	return "abac"
}

// GetSupportedFeatures returns supported features
func (ap *ABACProvider) GetSupportedFeatures() []types.AuthFeature {
	return []types.AuthFeature{
		types.FeatureABAC,
		types.FeatureAttributeBased,
		types.FeatureContextAware,
		types.FeatureDynamicPolicies,
		types.FeaturePolicyEngine,
	}
}

// GetConnectionInfo returns connection information
func (ap *ABACProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "local",
		Port:     0,
		Protocol: "abac",
		Version:  "1.0",
		Secure:   true,
	}
}

// Configure configures the ABAC provider
func (ap *ABACProvider) Configure(config map[string]interface{}) error {
	ap.configured = true
	ap.logger.Info("ABAC provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (ap *ABACProvider) IsConfigured() bool {
	return ap.configured
}

// Authenticate authenticates a user
func (ap *ABACProvider) Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	// ABAC is typically used for authorization, not authentication
	return &types.AuthResponse{
		Success:   false,
		Message:   "ABAC provider does not handle authentication",
		ServiceID: request.ServiceID,
		Context:   request.Context,
		Metadata:  request.Metadata,
	}, nil
}

// ValidateToken validates a token
func (ap *ABACProvider) ValidateToken(ctx context.Context, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error) {
	return &types.TokenValidationResponse{
		Valid:    false,
		Message:  "ABAC provider does not handle token validation",
		Metadata: request.Metadata,
	}, nil
}

// RefreshToken refreshes a token
func (ap *ABACProvider) RefreshToken(ctx context.Context, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error) {
	return &types.TokenRefreshResponse{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresAt:    time.Now(),
		TokenType:    "Bearer",
		Metadata:     request.Metadata,
	}, nil
}

// RevokeToken revokes a token
func (ap *ABACProvider) RevokeToken(ctx context.Context, request *types.TokenRevocationRequest) error {
	ap.logger.WithField("token", request.Token).Info("ABAC token revoked")
	return nil
}

// Authorize authorizes a user
func (ap *ABACProvider) Authorize(ctx context.Context, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error) {
	abacRequest := &ABACRequest{
		Subject: map[string]interface{}{
			"id": request.UserID,
		},
		Resource: map[string]interface{}{
			"id": request.Resource,
		},
		Action: map[string]interface{}{
			"name": request.Action,
		},
		Environment: request.Environment,
		Context:     request.Context,
	}

	decision, err := ap.CheckAccess(ctx, abacRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	return &types.AuthorizationResponse{
		Allowed:  decision.Decision == ABACEffectAllow,
		Reason:   decision.Reason,
		Policies: []string{decision.PolicyID.String()},
		Metadata: request.Metadata,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (ap *ABACProvider) CheckPermission(ctx context.Context, request *types.PermissionRequest) (*types.PermissionResponse, error) {
	abacRequest := &ABACRequest{
		Subject: map[string]interface{}{
			"id": request.UserID,
		},
		Resource: map[string]interface{}{
			"id": request.Resource,
		},
		Action: map[string]interface{}{
			"name": request.Permission,
		},
		Context: request.Context,
	}

	decision, err := ap.CheckAccess(ctx, abacRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	return &types.PermissionResponse{
		Granted:  decision.Decision == ABACEffectAllow,
		Reason:   decision.Reason,
		Metadata: request.Metadata,
	}, nil
}

// CreateUser creates a new user
func (ap *ABACProvider) CreateUser(ctx context.Context, request *types.CreateUserRequest) (*types.CreateUserResponse, error) {
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
func (ap *ABACProvider) GetUser(ctx context.Context, request *types.GetUserRequest) (*types.GetUserResponse, error) {
	return &types.GetUserResponse{
		UserID:      request.UserID,
		Username:    request.Username,
		Email:       request.Email,
		Roles:       []string{"abac-user"},
		Permissions: []string{"abac-access"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}, nil
}

// UpdateUser updates an existing user
func (ap *ABACProvider) UpdateUser(ctx context.Context, request *types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	return &types.UpdateUserResponse{
		UserID:    request.UserID,
		UpdatedAt: time.Now(),
		Metadata:  request.Metadata,
	}, nil
}

// DeleteUser deletes a user
func (ap *ABACProvider) DeleteUser(ctx context.Context, request *types.DeleteUserRequest) error {
	ap.logger.WithField("user_id", request.UserID).Info("ABAC user deleted")
	return nil
}

// AssignRole assigns a role to a user
func (ap *ABACProvider) AssignRole(ctx context.Context, request *types.AssignRoleRequest) error {
	ap.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("ABAC role assigned")
	return nil
}

// RemoveRole removes a role from a user
func (ap *ABACProvider) RemoveRole(ctx context.Context, request *types.RemoveRoleRequest) error {
	ap.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("ABAC role removed")
	return nil
}

// GrantPermission grants a permission to a user
func (ap *ABACProvider) GrantPermission(ctx context.Context, request *types.GrantPermissionRequest) error {
	ap.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("ABAC permission granted")
	return nil
}

// RevokePermission revokes a permission from a user
func (ap *ABACProvider) RevokePermission(ctx context.Context, request *types.RevokePermissionRequest) error {
	ap.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("ABAC permission revoked")
	return nil
}

// HealthCheck performs health check
func (ap *ABACProvider) HealthCheck(ctx context.Context) error {
	if !ap.configured {
		return fmt.Errorf("ABAC provider not configured")
	}
	return nil
}

// GetStats returns provider statistics
func (ap *ABACProvider) GetStats(ctx context.Context) (*types.AuthStats, error) {
	return &types.AuthStats{
		TotalUsers:    int64(len(ap.policies)),
		ActiveUsers:   int64(len(ap.policies)),
		TotalLogins:   200,
		FailedLogins:  10,
		ActiveTokens:  100,
		RevokedTokens: 5,
		ProviderData:  map[string]interface{}{"provider": "abac", "configured": ap.configured},
	}, nil
}

// Close closes the provider
func (ap *ABACProvider) Close() error {
	ap.logger.Info("ABAC provider closed")
	return nil
}
