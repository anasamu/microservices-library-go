package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ACLManager handles Access Control List operations
type ACLManager struct {
	entries map[string]*ACLEntry
	groups  map[string]*ACLGroup
	logger  *logrus.Logger
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

// NewACLManager creates a new ACL manager
func NewACLManager(logger *logrus.Logger) *ACLManager {
	return &ACLManager{
		entries: make(map[string]*ACLEntry),
		groups:  make(map[string]*ACLGroup),
		logger:  logger,
	}
}

// CreateACLEntry creates a new ACL entry
func (am *ACLManager) CreateACLEntry(ctx context.Context, principal string, principalType PrincipalType, resource, action string, effect ACLEffect, condition map[string]interface{}, priority int, description, createdBy string, metadata map[string]interface{}) (*ACLEntry, error) {
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

	key := am.generateEntryKey(principal, resource, action)
	am.entries[key] = entry

	am.logger.WithFields(logrus.Fields{
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
func (am *ACLManager) GetACLEntry(ctx context.Context, principal, resource, action string) (*ACLEntry, error) {
	key := am.generateEntryKey(principal, resource, action)
	entry, exists := am.entries[key]
	if !exists {
		return nil, fmt.Errorf("ACL entry not found")
	}
	return entry, nil
}

// UpdateACLEntry updates an existing ACL entry
func (am *ACLManager) UpdateACLEntry(ctx context.Context, principal, resource, action string, effect ACLEffect, condition map[string]interface{}, priority int, description string, metadata map[string]interface{}) (*ACLEntry, error) {
	key := am.generateEntryKey(principal, resource, action)
	entry, exists := am.entries[key]
	if !exists {
		return nil, fmt.Errorf("ACL entry not found")
	}

	entry.Effect = effect
	entry.Condition = condition
	entry.Priority = priority
	entry.Description = description
	entry.Metadata = metadata
	entry.UpdatedAt = time.Now()

	am.logger.WithFields(logrus.Fields{
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
func (am *ACLManager) DeleteACLEntry(ctx context.Context, principal, resource, action string) error {
	key := am.generateEntryKey(principal, resource, action)
	_, exists := am.entries[key]
	if !exists {
		return fmt.Errorf("ACL entry not found")
	}

	delete(am.entries, key)

	am.logger.WithFields(logrus.Fields{
		"principal": principal,
		"resource":  resource,
		"action":    action,
	}).Info("ACL entry deleted successfully")

	return nil
}

// CreateACLGroup creates a new ACL group
func (am *ACLManager) CreateACLGroup(ctx context.Context, name, description string, members []string, createdBy string, metadata map[string]interface{}) (*ACLGroup, error) {
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

	am.groups[name] = group

	am.logger.WithFields(logrus.Fields{
		"group_id": groupID,
		"name":     name,
		"members":  members,
	}).Info("ACL group created successfully")

	return group, nil
}

// GetACLGroup retrieves an ACL group
func (am *ACLManager) GetACLGroup(ctx context.Context, name string) (*ACLGroup, error) {
	group, exists := am.groups[name]
	if !exists {
		return nil, fmt.Errorf("ACL group not found: %s", name)
	}
	return group, nil
}

// AddGroupMember adds a member to an ACL group
func (am *ACLManager) AddGroupMember(ctx context.Context, groupName, member string) error {
	group, exists := am.groups[groupName]
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

	am.logger.WithFields(logrus.Fields{
		"group_name": groupName,
		"member":     member,
	}).Info("Member added to ACL group")

	return nil
}

// RemoveGroupMember removes a member from an ACL group
func (am *ACLManager) RemoveGroupMember(ctx context.Context, groupName, member string) error {
	group, exists := am.groups[groupName]
	if !exists {
		return fmt.Errorf("ACL group not found: %s", groupName)
	}

	for i, existingMember := range group.Members {
		if existingMember == member {
			group.Members = append(group.Members[:i], group.Members[i+1:]...)
			group.UpdatedAt = time.Now()

			am.logger.WithFields(logrus.Fields{
				"group_name": groupName,
				"member":     member,
			}).Info("Member removed from ACL group")

			return nil
		}
	}

	return fmt.Errorf("member not found in group")
}

// CheckAccess checks if a principal has access to a resource
func (am *ACLManager) CheckAccess(ctx context.Context, request *ACLRequest) (*ACLDecision, error) {
	decision := &ACLDecision{
		Decision: ACLEffectDeny,
		Reason:   "No matching ACL entries found",
		Context:  request.Context,
	}

	// Get all applicable entries
	applicableEntries := am.getApplicableEntries(request.Principal, request.Resource, request.Action)

	// Sort by priority (higher priority first)
	am.sortEntriesByPriority(applicableEntries)

	// Evaluate entries in priority order
	for _, entry := range applicableEntries {
		if am.evaluateEntry(entry, request) {
			decision.Decision = entry.Effect
			decision.Reason = fmt.Sprintf("ACL entry matched: %s", entry.Description)
			decision.EntryID = entry.ID
			break
		}
	}

	am.logger.WithFields(logrus.Fields{
		"principal": request.Principal,
		"resource":  request.Resource,
		"action":    request.Action,
		"decision":  decision.Decision,
		"reason":    decision.Reason,
	}).Debug("ACL access check completed")

	return decision, nil
}

// getApplicableEntries returns all ACL entries that could apply to the request
func (am *ACLManager) getApplicableEntries(principal, resource, action string) []*ACLEntry {
	var applicable []*ACLEntry

	for _, entry := range am.entries {
		if am.entryMatches(entry, principal, resource, action) {
			applicable = append(applicable, entry)
		}
	}

	return applicable
}

// entryMatches checks if an ACL entry matches the request
func (am *ACLManager) entryMatches(entry *ACLEntry, principal, resource, action string) bool {
	// Check principal match
	if !am.matchesPrincipal(entry.Principal, entry.PrincipalType, principal) {
		return false
	}

	// Check resource match
	if !am.matchesResource(entry.Resource, resource) {
		return false
	}

	// Check action match
	if !am.matchesAction(entry.Action, action) {
		return false
	}

	return true
}

// matchesPrincipal checks if a principal pattern matches the request principal
func (am *ACLManager) matchesPrincipal(pattern string, principalType PrincipalType, principal string) bool {
	switch principalType {
	case PrincipalTypeAny:
		return pattern == "*"
	case PrincipalTypeUser, PrincipalTypeRole:
		return pattern == principal || pattern == "*"
	case PrincipalTypeGroup:
		// Check if principal is a member of the group
		if group, exists := am.groups[pattern]; exists {
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
func (am *ACLManager) matchesResource(pattern, resource string) bool {
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
func (am *ACLManager) matchesAction(pattern, action string) bool {
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
func (am *ACLManager) evaluateEntry(entry *ACLEntry, request *ACLRequest) bool {
	// Check conditions
	if !am.evaluateConditions(entry.Condition, request.Context) {
		return false
	}

	return true
}

// evaluateConditions evaluates ACL entry conditions
func (am *ACLManager) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
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
func (am *ACLManager) sortEntriesByPriority(entries []*ACLEntry) {
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
func (am *ACLManager) generateEntryKey(principal, resource, action string) string {
	return fmt.Sprintf("%s:%s:%s", principal, resource, action)
}

// ListACLEntries returns all ACL entries
func (am *ACLManager) ListACLEntries(ctx context.Context) ([]*ACLEntry, error) {
	var entries []*ACLEntry
	for _, entry := range am.entries {
		entries = append(entries, entry)
	}
	return entries, nil
}

// ListACLGroups returns all ACL groups
func (am *ACLManager) ListACLGroups(ctx context.Context) ([]*ACLGroup, error) {
	var groups []*ACLGroup
	for _, group := range am.groups {
		groups = append(groups, group)
	}
	return groups, nil
}
