package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ABACManager handles Attribute-Based Access Control operations
type ABACManager struct {
	policies   map[string]*ABACPolicy
	attributes map[string]*Attribute
	logger     *logrus.Logger
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

// NewABACManager creates a new ABAC manager
func NewABACManager(logger *logrus.Logger) *ABACManager {
	return &ABACManager{
		policies:   make(map[string]*ABACPolicy),
		attributes: make(map[string]*Attribute),
		logger:     logger,
	}
}

// CreateAttribute creates a new attribute
func (am *ABACManager) CreateAttribute(ctx context.Context, name string, attrType AttributeType, category AttributeCategory, description string, values []string, defaultValue interface{}, required bool, createdBy string, metadata map[string]interface{}) (*Attribute, error) {
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

	am.attributes[name] = attribute

	am.logger.WithFields(logrus.Fields{
		"attribute_id": attrID,
		"name":         name,
		"type":         attrType,
		"category":     category,
		"required":     required,
	}).Info("Attribute created successfully")

	return attribute, nil
}

// GetAttribute retrieves an attribute by name
func (am *ABACManager) GetAttribute(ctx context.Context, name string) (*Attribute, error) {
	attribute, exists := am.attributes[name]
	if !exists {
		return nil, fmt.Errorf("attribute not found: %s", name)
	}
	return attribute, nil
}

// CreatePolicy creates a new ABAC policy
func (am *ABACManager) CreatePolicy(ctx context.Context, name, description string, effect ABACEffect, rules []ABACRule, target ABACTarget, condition map[string]interface{}, priority int, createdBy string, metadata map[string]interface{}) (*ABACPolicy, error) {
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

	am.policies[name] = policy

	am.logger.WithFields(logrus.Fields{
		"policy_id": policyID,
		"name":      name,
		"effect":    effect,
		"priority":  priority,
		"rules":     len(rules),
	}).Info("ABAC policy created successfully")

	return policy, nil
}

// GetPolicy retrieves an ABAC policy by name
func (am *ABACManager) GetPolicy(ctx context.Context, name string) (*ABACPolicy, error) {
	policy, exists := am.policies[name]
	if !exists {
		return nil, fmt.Errorf("ABAC policy not found: %s", name)
	}
	return policy, nil
}

// CheckAccess checks if a request is allowed based on ABAC policies
func (am *ABACManager) CheckAccess(ctx context.Context, request *ABACRequest) (*ABACDecision, error) {
	decision := &ABACDecision{
		Decision: ABACEffectDeny,
		Reason:   "No matching ABAC policies found",
		Context:  request.Context,
	}

	// Get all applicable policies
	applicablePolicies := am.getApplicablePolicies(request)

	// Sort by priority (higher priority first)
	am.sortPoliciesByPriority(applicablePolicies)

	// Evaluate policies in priority order
	for _, policy := range applicablePolicies {
		if am.evaluatePolicy(policy, request) {
			decision.Decision = policy.Effect
			decision.Reason = fmt.Sprintf("ABAC policy matched: %s", policy.Description)
			decision.PolicyID = policy.ID
			break
		}
	}

	am.logger.WithFields(logrus.Fields{
		"decision": decision.Decision,
		"reason":   decision.Reason,
	}).Debug("ABAC access check completed")

	return decision, nil
}

// getApplicablePolicies returns all policies that could apply to the request
func (am *ABACManager) getApplicablePolicies(request *ABACRequest) []*ABACPolicy {
	var applicable []*ABACPolicy

	for _, policy := range am.policies {
		if am.policyMatches(policy, request) {
			applicable = append(applicable, policy)
		}
	}

	return applicable
}

// policyMatches checks if a policy matches the request
func (am *ABACManager) policyMatches(policy *ABACPolicy, request *ABACRequest) bool {
	// Check target matching
	if !am.matchesTarget(policy.Target, request) {
		return false
	}

	// Check policy-level conditions
	if !am.evaluateConditions(policy.Condition, request) {
		return false
	}

	return true
}

// matchesTarget checks if the policy target matches the request
func (am *ABACManager) matchesTarget(target ABACTarget, request *ABACRequest) bool {
	// Check subjects
	if len(target.Subjects) > 0 {
		subjectID, exists := request.Subject["id"]
		if !exists || !am.contains(target.Subjects, fmt.Sprintf("%v", subjectID)) {
			return false
		}
	}

	// Check resources
	if len(target.Resources) > 0 {
		resourceID, exists := request.Resource["id"]
		if !exists || !am.matchesAnyPattern(target.Resources, fmt.Sprintf("%v", resourceID)) {
			return false
		}
	}

	// Check actions
	if len(target.Actions) > 0 {
		actionName, exists := request.Action["name"]
		if !exists || !am.matchesAnyPattern(target.Actions, fmt.Sprintf("%v", actionName)) {
			return false
		}
	}

	// Check environments
	if len(target.Environments) > 0 {
		environment, exists := request.Environment["name"]
		if !exists || !am.contains(target.Environments, fmt.Sprintf("%v", environment)) {
			return false
		}
	}

	return true
}

// evaluatePolicy evaluates an ABAC policy against the request
func (am *ABACManager) evaluatePolicy(policy *ABACPolicy, request *ABACRequest) bool {
	// Sort rules by priority
	am.sortRulesByPriority(policy.Rules)

	// Evaluate rules in priority order
	for _, rule := range policy.Rules {
		if am.evaluateRule(rule, request) {
			return rule.Effect == ABACEffectAllow
		}
	}

	// If no rules match, use policy default effect
	return policy.Effect == ABACEffectAllow
}

// evaluateRule evaluates an ABAC rule against the request
func (am *ABACManager) evaluateRule(rule ABACRule, request *ABACRequest) bool {
	// Check rule conditions
	if !am.evaluateConditions(rule.Condition, request) {
		return false
	}

	return true
}

// evaluateConditions evaluates conditions against the request
func (am *ABACManager) evaluateConditions(conditions map[string]interface{}, request *ABACRequest) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, expectedValue := range conditions {
		actualValue := am.getAttributeValue(key, request)
		if actualValue == nil {
			return false
		}

		if !am.compareValues(actualValue, expectedValue) {
			return false
		}
	}

	return true
}

// getAttributeValue retrieves an attribute value from the request
func (am *ABACManager) getAttributeValue(key string, request *ABACRequest) interface{} {
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
func (am *ABACManager) compareValues(actual, expected interface{}) bool {
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
		return am.compareArrays(actualArray, expectedValue)
	default:
		return actual == expected
	}
}

// compareArrays compares two arrays
func (am *ABACManager) compareArrays(actual, expected []interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, expectedItem := range expected {
		if !am.compareValues(actual[i], expectedItem) {
			return false
		}
	}

	return true
}

// contains checks if a slice contains a value
func (am *ABACManager) contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// matchesAnyPattern checks if a value matches any pattern in the slice
func (am *ABACManager) matchesAnyPattern(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if am.matchesPattern(pattern, value) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a value matches a pattern
func (am *ABACManager) matchesPattern(pattern, value string) bool {
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
func (am *ABACManager) sortPoliciesByPriority(policies []*ABACPolicy) {
	for i := 0; i < len(policies)-1; i++ {
		for j := 0; j < len(policies)-i-1; j++ {
			if policies[j].Priority < policies[j+1].Priority {
				policies[j], policies[j+1] = policies[j+1], policies[j]
			}
		}
	}
}

// sortRulesByPriority sorts rules by priority (higher priority first)
func (am *ABACManager) sortRulesByPriority(rules []ABACRule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := 0; j < len(rules)-i-1; j++ {
			if rules[j].Priority < rules[j+1].Priority {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}
}

// ListAttributes returns all attributes
func (am *ABACManager) ListAttributes(ctx context.Context) ([]*Attribute, error) {
	var attributes []*Attribute
	for _, attribute := range am.attributes {
		attributes = append(attributes, attribute)
	}
	return attributes, nil
}

// ListPolicies returns all ABAC policies
func (am *ABACManager) ListPolicies(ctx context.Context) ([]*ABACPolicy, error) {
	var policies []*ABACPolicy
	for _, policy := range am.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}
