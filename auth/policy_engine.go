package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PolicyEngine handles dynamic policy evaluation
type PolicyEngine struct {
	rbacManager *RBACManager
	aclManager  *ACLManager
	abacManager *ABACManager
	logger      *logrus.Logger
}

// PolicyRequest represents a policy evaluation request
type PolicyRequest struct {
	UserID      string                 `json:"user_id"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Environment map[string]interface{} `json:"environment,omitempty"`
}

// PolicyDecision represents a policy evaluation decision
type PolicyDecision struct {
	Decision    PolicyEngineEffect     `json:"decision"`
	Reason      string                 `json:"reason"`
	Source      PolicySource           `json:"source"`
	PolicyID    uuid.UUID              `json:"policy_id,omitempty"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// PolicyEngineEffect represents the effect of a policy decision
type PolicyEngineEffect string

const (
	PolicyEngineEffectAllow PolicyEngineEffect = "allow"
	PolicyEngineEffectDeny  PolicyEngineEffect = "deny"
)

// PolicySource represents the source of a policy decision
type PolicySource string

const (
	PolicySourceRBAC    PolicySource = "rbac"
	PolicySourceACL     PolicySource = "acl"
	PolicySourceABAC    PolicySource = "abac"
	PolicySourceDefault PolicySource = "default"
)

// PolicyEvaluationOrder defines the order of policy evaluation
type PolicyEvaluationOrder struct {
	RBACEnabled bool `json:"rbac_enabled"`
	ACLEnabled  bool `json:"acl_enabled"`
	ABACEnabled bool `json:"abac_enabled"`
	RBACFirst   bool `json:"rbac_first"`
	ACLFirst    bool `json:"acl_first"`
	ABACFirst   bool `json:"abac_first"`
}

// DefaultPolicyEvaluationOrder returns the default evaluation order
func DefaultPolicyEvaluationOrder() *PolicyEvaluationOrder {
	return &PolicyEvaluationOrder{
		RBACEnabled: true,
		ACLEnabled:  true,
		ABACEnabled: true,
		RBACFirst:   true,
		ACLFirst:    false,
		ABACFirst:   false,
	}
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(rbacManager *RBACManager, aclManager *ACLManager, abacManager *ABACManager, logger *logrus.Logger) *PolicyEngine {
	return &PolicyEngine{
		rbacManager: rbacManager,
		aclManager:  aclManager,
		abacManager: abacManager,
		logger:      logger,
	}
}

// EvaluatePolicy evaluates a policy request using the configured policy engines
func (pe *PolicyEngine) EvaluatePolicy(ctx context.Context, request *PolicyRequest, order *PolicyEvaluationOrder) (*PolicyDecision, error) {
	if order == nil {
		order = DefaultPolicyEvaluationOrder()
	}

	decision := &PolicyDecision{
		Decision:    PolicyEngineEffectDeny,
		Reason:      "No policies matched",
		Source:      PolicySourceDefault,
		EvaluatedAt: time.Now(),
		Context:     request.Context,
	}

	// Determine evaluation order
	evaluationOrder := pe.determineEvaluationOrder(order)

	// Evaluate policies in the determined order
	for _, source := range evaluationOrder {
		switch source {
		case PolicySourceRBAC:
			if order.RBACEnabled {
				if rbacDecision, err := pe.evaluateRBAC(ctx, request); err == nil {
					if rbacDecision.Decision == PolicyEffectAllow {
						decision.Decision = PolicyEngineEffectAllow
						decision.Reason = rbacDecision.Reason
						decision.Source = PolicySourceRBAC
						// Note: RBAC policies don't have UUID IDs, so we'll use a placeholder
						decision.PolicyID = uuid.New()
						return decision, nil
					}
				}
			}
		case PolicySourceACL:
			if order.ACLEnabled {
				if aclDecision, err := pe.evaluateACL(ctx, request); err == nil {
					if aclDecision.Decision == ACLEffectAllow {
						decision.Decision = PolicyEngineEffectAllow
						decision.Reason = aclDecision.Reason
						decision.Source = PolicySourceACL
						decision.PolicyID = aclDecision.EntryID
						return decision, nil
					} else if aclDecision.Decision == ACLEffectDeny {
						decision.Decision = PolicyEngineEffectDeny
						decision.Reason = aclDecision.Reason
						decision.Source = PolicySourceACL
						decision.PolicyID = aclDecision.EntryID
						return decision, nil
					}
				}
			}
		case PolicySourceABAC:
			if order.ABACEnabled {
				if abacDecision, err := pe.evaluateABAC(ctx, request); err == nil {
					if abacDecision.Decision == ABACEffectAllow {
						decision.Decision = PolicyEngineEffectAllow
						decision.Reason = abacDecision.Reason
						decision.Source = PolicySourceABAC
						decision.PolicyID = abacDecision.PolicyID
						return decision, nil
					} else if abacDecision.Decision == ABACEffectDeny {
						decision.Decision = PolicyEngineEffectDeny
						decision.Reason = abacDecision.Reason
						decision.Source = PolicySourceABAC
						decision.PolicyID = abacDecision.PolicyID
						return decision, nil
					}
				}
			}
		}
	}

	pe.logger.WithFields(logrus.Fields{
		"user_id":   request.UserID,
		"tenant_id": request.TenantID,
		"resource":  request.Resource,
		"action":    request.Action,
		"decision":  decision.Decision,
		"reason":    decision.Reason,
		"source":    decision.Source,
	}).Info("Policy evaluation completed")

	return decision, nil
}

// determineEvaluationOrder determines the order of policy evaluation
func (pe *PolicyEngine) determineEvaluationOrder(order *PolicyEvaluationOrder) []PolicySource {
	var sources []PolicySource

	// Add sources based on priority
	if order.RBACFirst && order.RBACEnabled {
		sources = append(sources, PolicySourceRBAC)
	}
	if order.ACLFirst && order.ACLEnabled {
		sources = append(sources, PolicySourceACL)
	}
	if order.ABACFirst && order.ABACEnabled {
		sources = append(sources, PolicySourceABAC)
	}

	// Add remaining enabled sources
	if order.RBACEnabled && !order.RBACFirst {
		sources = append(sources, PolicySourceRBAC)
	}
	if order.ACLEnabled && !order.ACLFirst {
		sources = append(sources, PolicySourceACL)
	}
	if order.ABACEnabled && !order.ABACFirst {
		sources = append(sources, PolicySourceABAC)
	}

	return sources
}

// evaluateRBAC evaluates RBAC policies
func (pe *PolicyEngine) evaluateRBAC(ctx context.Context, request *PolicyRequest) (*AccessDecision, error) {
	accessRequest := &AccessRequest{
		Principal: request.UserID,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	}

	return pe.rbacManager.CheckAccess(ctx, accessRequest)
}

// evaluateACL evaluates ACL policies
func (pe *PolicyEngine) evaluateACL(ctx context.Context, request *PolicyRequest) (*ACLDecision, error) {
	aclRequest := &ACLRequest{
		Principal: request.UserID,
		Resource:  request.Resource,
		Action:    request.Action,
		Context:   request.Context,
	}

	return pe.aclManager.CheckAccess(ctx, aclRequest)
}

// evaluateABAC evaluates ABAC policies
func (pe *PolicyEngine) evaluateABAC(ctx context.Context, request *PolicyRequest) (*ABACDecision, error) {
	// Convert request to ABAC format
	abacRequest := &ABACRequest{
		Subject: map[string]interface{}{
			"id":        request.UserID,
			"tenant_id": request.TenantID,
		},
		Resource: map[string]interface{}{
			"id":   request.Resource,
			"type": pe.extractResourceType(request.Resource),
		},
		Action: map[string]interface{}{
			"name": request.Action,
		},
		Environment: request.Environment,
		Context:     request.Context,
	}

	// Add user attributes from context
	if userAttrs, exists := request.Context["user_attributes"]; exists {
		if attrs, ok := userAttrs.(map[string]interface{}); ok {
			for k, v := range attrs {
				abacRequest.Subject[k] = v
			}
		}
	}

	// Add resource attributes from context
	if resourceAttrs, exists := request.Context["resource_attributes"]; exists {
		if attrs, ok := resourceAttrs.(map[string]interface{}); ok {
			for k, v := range attrs {
				abacRequest.Resource[k] = v
			}
		}
	}

	return pe.abacManager.CheckAccess(ctx, abacRequest)
}

// extractResourceType extracts resource type from resource string
func (pe *PolicyEngine) extractResourceType(resource string) string {
	parts := strings.Split(resource, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return resource
}

// BatchEvaluatePolicy evaluates multiple policy requests
func (pe *PolicyEngine) BatchEvaluatePolicy(ctx context.Context, requests []*PolicyRequest, order *PolicyEvaluationOrder) ([]*PolicyDecision, error) {
	var decisions []*PolicyDecision

	for _, request := range requests {
		decision, err := pe.EvaluatePolicy(ctx, request, order)
		if err != nil {
			pe.logger.WithError(err).WithField("user_id", request.UserID).Error("Failed to evaluate policy")
			// Return deny decision for failed evaluations
			decision = &PolicyDecision{
				Decision:    PolicyEngineEffectDeny,
				Reason:      fmt.Sprintf("Policy evaluation failed: %v", err),
				Source:      PolicySourceDefault,
				EvaluatedAt: time.Now(),
				Context:     request.Context,
			}
		}
		decisions = append(decisions, decision)
	}

	return decisions, nil
}

// GetPolicyStatistics returns statistics about policy evaluation
func (pe *PolicyEngine) GetPolicyStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// RBAC statistics
	if pe.rbacManager != nil {
		rbacStats := map[string]interface{}{
			"roles_count":       len(pe.rbacManager.roles),
			"permissions_count": len(pe.rbacManager.permissions),
			"policies_count":    len(pe.rbacManager.policies),
		}
		stats["rbac"] = rbacStats
	}

	// ACL statistics
	if pe.aclManager != nil {
		aclStats := map[string]interface{}{
			"entries_count": len(pe.aclManager.entries),
			"groups_count":  len(pe.aclManager.groups),
		}
		stats["acl"] = aclStats
	}

	// ABAC statistics
	if pe.abacManager != nil {
		abacStats := map[string]interface{}{
			"policies_count":   len(pe.abacManager.policies),
			"attributes_count": len(pe.abacManager.attributes),
		}
		stats["abac"] = abacStats
	}

	return stats, nil
}

// ValidatePolicyConfiguration validates the policy configuration
func (pe *PolicyEngine) ValidatePolicyConfiguration(ctx context.Context) error {
	var errors []string

	// Validate RBAC configuration
	if pe.rbacManager != nil {
		if len(pe.rbacManager.roles) == 0 {
			errors = append(errors, "RBAC: No roles configured")
		}
		if len(pe.rbacManager.permissions) == 0 {
			errors = append(errors, "RBAC: No permissions configured")
		}
	}

	// Validate ACL configuration
	if pe.aclManager != nil {
		if len(pe.aclManager.entries) == 0 {
			errors = append(errors, "ACL: No entries configured")
		}
	}

	// Validate ABAC configuration
	if pe.abacManager != nil {
		if len(pe.abacManager.attributes) == 0 {
			errors = append(errors, "ABAC: No attributes configured")
		}
		if len(pe.abacManager.policies) == 0 {
			errors = append(errors, "ABAC: No policies configured")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("policy configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
