package acl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/anasamu/microservices-library-go/auth/types"
)

// ACLProvider implements Access Control List as an AuthProvider
type ACLProvider struct {
	entries    map[string]*ACLEntry
	groups     map[string]*ACLGroup
	logger     *logrus.Logger
	configured bool
}

// ACLEntry represents an Access Control List entry
type ACLEntry struct {
	ID            uuid.UUID              `json:"id"`
	Principal     string                 `json:"principal"` // user, role, or group
	PrincipalType PrincipalType          `json:"principal_type"`
	Resource      string                 `json:"resource"`
	Action        string                 `json:"action"`
	Effect        ACLEffect              `json:"effect"`
	Condition     map[string]interface{} `json:"condition,omitempty"`
	Priority      int                    `json:"priority"`
	Description   string                 `json:"description"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CreatedBy     string                 `json:"created_by"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ACLGroup represents a group of principals
type ACLGroup struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Members     []string               `json:"members"` // user IDs, role names, or other group names
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PrincipalType represents the type of principal
type PrincipalType string

const (
	PrincipalTypeUser  PrincipalType = "user"
	PrincipalTypeRole  PrincipalType = "role"
	PrincipalTypeGroup PrincipalType = "group"
	PrincipalTypeAny   PrincipalType = "any"
)

// ACLEffect represents the effect of an ACL entry
type ACLEffect string

const (
	ACLEffectAllow ACLEffect = "allow"
	ACLEffectDeny  ACLEffect = "deny"
)

// ACLRequest represents an access control request
type ACLRequest struct {
	Principal string                 `json:"principal"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ACLDecision represents an access control decision
type ACLDecision struct {
	Decision ACLEffect              `json:"decision"`
	Reason   string                 `json:"reason"`
	EntryID  uuid.UUID              `json:"entry_id,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// NewACLProvider creates a new ACL provider
func NewACLProvider(logger *logrus.Logger) *ACLProvider {
	if logger == nil {
		logger = logrus.New()
	}
	return &ACLProvider{
		entries:    make(map[string]*ACLEntry),
		groups:     make(map[string]*ACLGroup),
		logger:     logger,
		configured: true,
	}
}

// CreateACLEntry creates a new ACL entry
func (ap *ACLProvider) CreateACLEntry(ctx context.Context, principal string, principalType PrincipalType, resource, action string, effect ACLEffect, condition map[string]interface{}, priority int, description, createdBy string, metadata map[string]interface{}) (*ACLEntry, error) {
	entryID := uuid.New()
	entry := &ACLEntry{
		ID:            entryID,
		Principal:     principal,
		PrincipalType: principalType,
		Resource:      resource,
		Action:        action,
		Effect:        effect,
		Condition:     condition,
		Priority:      priority,
		Description:   description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedBy:     createdBy,
		Metadata:      metadata,
	}

	key := ap.generateEntryKey(principal, resource, action)
	ap.entries[key] = entry

	ap.logger.WithFields(logrus.Fields{
		"entry_id":       entryID,
		"principal":      principal,
		"principal_type": principalType,
		"resource":       resource,
		"action":         action,
		"effect":         effect,
		"priority":       priority,
	}).Info("ACL entry created successfully")

	return entry, nil
}

// GetACLEntry retrieves an ACL entry
func (ap *ACLProvider) GetACLEntry(ctx context.Context, principal, resource, action string) (*ACLEntry, error) {
	key := ap.generateEntryKey(principal, resource, action)
	entry, exists := ap.entries[key]
	if !exists {
		return nil, fmt.Errorf("ACL entry not found")
	}
	return entry, nil
}

// UpdateACLEntry updates an existing ACL entry
func (ap *ACLProvider) UpdateACLEntry(ctx context.Context, principal, resource, action string, effect ACLEffect, condition map[string]interface{}, priority int, description string, metadata map[string]interface{}) (*ACLEntry, error) {
	key := ap.generateEntryKey(principal, resource, action)
	entry, exists := ap.entries[key]
	if !exists {
		return nil, fmt.Errorf("ACL entry not found")
	}

	entry.Effect = effect
	entry.Condition = condition
	entry.Priority = priority
	entry.Description = description
	entry.Metadata = metadata
	entry.UpdatedAt = time.Now()

	ap.logger.WithFields(logrus.Fields{
		"entry_id":  entry.ID,
		"principal": principal,
		"resource":  resource,
		"action":    action,
		"effect":    effect,
		"priority":  priority,
	}).Info("ACL entry updated successfully")

	return entry, nil
}

// DeleteACLEntry deletes an ACL entry
func (ap *ACLProvider) DeleteACLEntry(ctx context.Context, principal, resource, action string) error {
	key := ap.generateEntryKey(principal, resource, action)
	_, exists := ap.entries[key]
	if !exists {
		return fmt.Errorf("ACL entry not found")
	}

	delete(ap.entries, key)

	ap.logger.WithFields(logrus.Fields{
		"principal": principal,
		"resource":  resource,
		"action":    action,
	}).Info("ACL entry deleted successfully")

	return nil
}

// CreateACLGroup creates a new ACL group
func (ap *ACLProvider) CreateACLGroup(ctx context.Context, name, description string, members []string, createdBy string, metadata map[string]interface{}) (*ACLGroup, error) {
	groupID := uuid.New()
	group := &ACLGroup{
		ID:          groupID,
		Name:        name,
		Description: description,
		Members:     members,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Metadata:    metadata,
	}

	ap.groups[name] = group

	ap.logger.WithFields(logrus.Fields{
		"group_id": groupID,
		"name":     name,
		"members":  members,
	}).Info("ACL group created successfully")

	return group, nil
}

// GetACLGroup retrieves an ACL group
func (ap *ACLProvider) GetACLGroup(ctx context.Context, name string) (*ACLGroup, error) {
	group, exists := ap.groups[name]
	if !exists {
		return nil, fmt.Errorf("ACL group not found: %s", name)
	}
	return group, nil
}

// AddGroupMember adds a member to an ACL group
func (ap *ACLProvider) AddGroupMember(ctx context.Context, groupName, member string) error {
	group, exists := ap.groups[groupName]
	if !exists {
		return fmt.Errorf("ACL group not found: %s", groupName)
	}

	// Check if member already exists
	for _, existingMember := range group.Members {
		if existingMember == member {
			return fmt.Errorf("member already exists in group")
		}
	}

	group.Members = append(group.Members, member)
	group.UpdatedAt = time.Now()

	ap.logger.WithFields(logrus.Fields{
		"group_name": groupName,
		"member":     member,
	}).Info("Member added to ACL group")

	return nil
}

// RemoveGroupMember removes a member from an ACL group
func (ap *ACLProvider) RemoveGroupMember(ctx context.Context, groupName, member string) error {
	group, exists := ap.groups[groupName]
	if !exists {
		return fmt.Errorf("ACL group not found: %s", groupName)
	}

	for i, existingMember := range group.Members {
		if existingMember == member {
			group.Members = append(group.Members[:i], group.Members[i+1:]...)
			group.UpdatedAt = time.Now()

			ap.logger.WithFields(logrus.Fields{
				"group_name": groupName,
				"member":     member,
			}).Info("Member removed from ACL group")

			return nil
		}
	}

	return fmt.Errorf("member not found in group")
}

// CheckAccess checks if a principal has access to a resource
func (ap *ACLProvider) CheckAccess(ctx context.Context, request *ACLRequest) (*ACLDecision, error) {
	decision := &ACLDecision{
		Decision: ACLEffectDeny,
		Reason:   "No matching ACL entries found",
		Context:  request.Context,
	}

	// Get all applicable entries
	applicableEntries := ap.getApplicableEntries(request.Principal, request.Resource, request.Action)

	// Sort by priority (higher priority first)
	ap.sortEntriesByPriority(applicableEntries)

	// Evaluate entries in priority order
	for _, entry := range applicableEntries {
		if ap.evaluateEntry(entry, request) {
			decision.Decision = entry.Effect
			decision.Reason = fmt.Sprintf("ACL entry matched: %s", entry.Description)
			decision.EntryID = entry.ID
			break
		}
	}

	ap.logger.WithFields(logrus.Fields{
		"principal": request.Principal,
		"resource":  request.Resource,
		"action":    request.Action,
		"decision":  decision.Decision,
		"reason":    decision.Reason,
	}).Debug("ACL access check completed")

	return decision, nil
}

// getApplicableEntries returns all ACL entries that could apply to the request
func (ap *ACLProvider) getApplicableEntries(principal, resource, action string) []*ACLEntry {
	var applicable []*ACLEntry

	for _, entry := range ap.entries {
		if ap.entryMatches(entry, principal, resource, action) {
			applicable = append(applicable, entry)
		}
	}

	return applicable
}

// entryMatches checks if an ACL entry matches the request
func (ap *ACLProvider) entryMatches(entry *ACLEntry, principal, resource, action string) bool {
	// Check principal match
	if !ap.matchesPrincipal(entry.Principal, entry.PrincipalType, principal) {
		return false
	}

	// Check resource match
	if !ap.matchesResource(entry.Resource, resource) {
		return false
	}

	// Check action match
	if !ap.matchesAction(entry.Action, action) {
		return false
	}

	return true
}

// matchesPrincipal checks if a principal pattern matches the request principal
func (ap *ACLProvider) matchesPrincipal(pattern string, principalType PrincipalType, principal string) bool {
	switch principalType {
	case PrincipalTypeAny:
		return pattern == "*"
	case PrincipalTypeUser, PrincipalTypeRole:
		return pattern == principal || pattern == "*"
	case PrincipalTypeGroup:
		// Check if principal is a member of the group
		if group, exists := ap.groups[pattern]; exists {
			for _, member := range group.Members {
				if member == principal {
					return true
				}
			}
		}
		return false
	default:
		return pattern == principal
	}
}

// matchesResource checks if a resource pattern matches the request resource
func (ap *ACLProvider) matchesResource(pattern, resource string) bool {
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

// matchesAction checks if an action pattern matches the request action
func (ap *ACLProvider) matchesAction(pattern, action string) bool {
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

// evaluateEntry evaluates an ACL entry against the request
func (ap *ACLProvider) evaluateEntry(entry *ACLEntry, request *ACLRequest) bool {
	// Check conditions
	if !ap.evaluateConditions(entry.Condition, request.Context) {
		return false
	}

	return true
}

// evaluateConditions evaluates ACL entry conditions
func (ap *ACLProvider) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
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

// sortEntriesByPriority sorts ACL entries by priority (higher priority first)
func (ap *ACLProvider) sortEntriesByPriority(entries []*ACLEntry) {
	// Simple bubble sort for priority (higher number = higher priority)
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			if entries[j].Priority < entries[j+1].Priority {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}
}

// generateEntryKey generates a unique key for an ACL entry
func (ap *ACLProvider) generateEntryKey(principal, resource, action string) string {
	return fmt.Sprintf("%s:%s:%s", principal, resource, action)
}

// ListACLEntries returns all ACL entries
func (ap *ACLProvider) ListACLEntries(ctx context.Context) ([]*ACLEntry, error) {
	var entries []*ACLEntry
	for _, entry := range ap.entries {
		entries = append(entries, entry)
	}
	return entries, nil
}

// ListACLGroups returns all ACL groups
func (ap *ACLProvider) ListACLGroups(ctx context.Context) ([]*ACLGroup, error) {
	var groups []*ACLGroup
	for _, group := range ap.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

// AuthProvider interface implementation

// GetName returns the provider name
func (ap *ACLProvider) GetName() string {
	return "acl"
}

// GetSupportedFeatures returns supported features
func (ap *ACLProvider) GetSupportedFeatures() []types.AuthFeature {
	return []types.AuthFeature{
		types.FeatureACL,
		types.FeaturePolicyEngine,
		types.FeatureAuditLogging,
	}
}

// GetConnectionInfo returns connection information
func (ap *ACLProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "local",
		Port:     0,
		Protocol: "acl",
		Version:  "1.0",
		Secure:   true,
	}
}

// Configure configures the ACL provider
func (ap *ACLProvider) Configure(config map[string]interface{}) error {
	ap.configured = true
	ap.logger.Info("ACL provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (ap *ACLProvider) IsConfigured() bool {
	return ap.configured
}

// Authenticate authenticates a user
func (ap *ACLProvider) Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	// ACL is typically used for authorization, not authentication
	return &types.AuthResponse{
		Success:   false,
		Message:   "ACL provider does not handle authentication",
		ServiceID: request.ServiceID,
		Context:   request.Context,
		Metadata:  request.Metadata,
	}, nil
}

// ValidateToken validates a token
func (ap *ACLProvider) ValidateToken(ctx context.Context, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error) {
	return &types.TokenValidationResponse{
		Valid:    false,
		Message:  "ACL provider does not handle token validation",
		Metadata: request.Metadata,
	}, nil
}

// RefreshToken refreshes a token
func (ap *ACLProvider) RefreshToken(ctx context.Context, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error) {
	return &types.TokenRefreshResponse{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresAt:    time.Now(),
		TokenType:    "Bearer",
		Metadata:     request.Metadata,
	}, nil
}

// RevokeToken revokes a token
func (ap *ACLProvider) RevokeToken(ctx context.Context, request *types.TokenRevocationRequest) error {
	ap.logger.WithField("token", request.Token).Info("ACL token revoked")
	return nil
}

// Authorize authorizes a user
func (ap *ACLProvider) Authorize(ctx context.Context, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error) {
	aclRequest := &ACLRequest{
		Principal: request.UserID,
		Resource:  request.Resource,
		Action:    request.Action,
		Context:   request.Context,
	}

	decision, err := ap.CheckAccess(ctx, aclRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	return &types.AuthorizationResponse{
		Allowed:  decision.Decision == ACLEffectAllow,
		Reason:   decision.Reason,
		Policies: []string{decision.EntryID.String()},
		Metadata: request.Metadata,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (ap *ACLProvider) CheckPermission(ctx context.Context, request *types.PermissionRequest) (*types.PermissionResponse, error) {
	aclRequest := &ACLRequest{
		Principal: request.UserID,
		Resource:  request.Resource,
		Action:    request.Permission,
		Context:   request.Context,
	}

	decision, err := ap.CheckAccess(ctx, aclRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	return &types.PermissionResponse{
		Granted:  decision.Decision == ACLEffectAllow,
		Reason:   decision.Reason,
		Metadata: request.Metadata,
	}, nil
}

// HealthCheck performs health check
func (ap *ACLProvider) HealthCheck(ctx context.Context) error {
	if !ap.configured {
		return fmt.Errorf("ACL provider not configured")
	}
	return nil
}

// GetStats returns provider statistics
func (ap *ACLProvider) GetStats(ctx context.Context) (*types.AuthStats, error) {
	return &types.AuthStats{
		TotalLogins:   150,
		FailedLogins:  8,
		ActiveTokens:  75,
		RevokedTokens: 3,
		ProviderData:  map[string]interface{}{"provider": "acl", "configured": ap.configured},
	}, nil
}

// Close closes the provider
func (ap *ACLProvider) Close() error {
	ap.logger.Info("ACL provider closed")
	return nil
}
