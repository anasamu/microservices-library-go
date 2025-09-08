package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuditManager handles audit logging for authentication events
type AuditManager struct {
	config *AuditConfig
	logger *logrus.Logger
	events chan *AuditEvent
	stop   chan bool
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID        uuid.UUID              `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType AuditEventType         `json:"event_type"`
	UserID    string                 `json:"user_id,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Result    AuditResult            `json:"result"`
	Reason    string                 `json:"reason,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventTypeLogin              AuditEventType = "login"
	AuditEventTypeLogout             AuditEventType = "logout"
	AuditEventTypeTokenGeneration    AuditEventType = "token_generation"
	AuditEventTypeTokenValidation    AuditEventType = "token_validation"
	AuditEventTypeTokenRefresh       AuditEventType = "token_refresh"
	AuditEventTypeTokenRevocation    AuditEventType = "token_revocation"
	AuditEventTypeAccessGranted      AuditEventType = "access_granted"
	AuditEventTypeAccessDenied       AuditEventType = "access_denied"
	AuditEventTypeRoleAssigned       AuditEventType = "role_assigned"
	AuditEventTypeRoleUnassigned     AuditEventType = "role_unassigned"
	AuditEventTypePermissionGranted  AuditEventType = "permission_granted"
	AuditEventTypePermissionRevoked  AuditEventType = "permission_revoked"
	AuditEventTypePolicyCreated      AuditEventType = "policy_created"
	AuditEventTypePolicyUpdated      AuditEventType = "policy_updated"
	AuditEventTypePolicyDeleted      AuditEventType = "policy_deleted"
	AuditEventTypeUserCreated        AuditEventType = "user_created"
	AuditEventTypeUserUpdated        AuditEventType = "user_updated"
	AuditEventTypeUserDeleted        AuditEventType = "user_deleted"
	AuditEventTypePasswordChanged    AuditEventType = "password_changed"
	AuditEventTypeAccountLocked      AuditEventType = "account_locked"
	AuditEventTypeAccountUnlocked    AuditEventType = "account_unlocked"
	AuditEventTypeSuspiciousActivity AuditEventType = "suspicious_activity"
)

// AuditResult represents the result of an audit event
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultWarning AuditResult = "warning"
	AuditResultInfo    AuditResult = "info"
)

// NewAuditManager creates a new audit manager
func NewAuditManager(config *AuditConfig, logger *logrus.Logger) *AuditManager {
	am := &AuditManager{
		config: config,
		logger: logger,
		events: make(chan *AuditEvent, 1000), // Buffer for 1000 events
		stop:   make(chan bool),
	}

	// Start background worker
	go am.worker()

	return am
}

// LogEvent logs an audit event
func (am *AuditManager) LogEvent(ctx context.Context, event *AuditEvent) error {
	if !am.config.Enabled {
		return nil
	}

	// Set default values
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Check if event type should be logged
	if !am.shouldLogEvent(event.EventType) {
		return nil
	}

	// Send event to worker
	select {
	case am.events <- event:
		return nil
	case <-time.After(5 * time.Second):
		am.logger.WithField("event_type", event.EventType).Warn("Audit event queue full, dropping event")
		return fmt.Errorf("audit event queue full")
	}
}

// LogLogin logs a login event
func (am *AuditManager) LogLogin(ctx context.Context, userID, tenantID, sessionID, ipAddress, userAgent string, result AuditResult, reason string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: AuditEventTypeLogin,
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    result,
		Reason:    reason,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// LogLogout logs a logout event
func (am *AuditManager) LogLogout(ctx context.Context, userID, tenantID, sessionID, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: AuditEventTypeLogout,
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// LogTokenGeneration logs a token generation event
func (am *AuditManager) LogTokenGeneration(ctx context.Context, userID, tenantID, sessionID, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: AuditEventTypeTokenGeneration,
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// LogTokenValidation logs a token validation event
func (am *AuditManager) LogTokenValidation(ctx context.Context, userID, tenantID, sessionID, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: AuditEventTypeTokenValidation,
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// LogAccessDecision logs an access decision event
func (am *AuditManager) LogAccessDecision(ctx context.Context, userID, tenantID, resource, action, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	eventType := AuditEventTypeAccessGranted
	if result == AuditResultFailure {
		eventType = AuditEventTypeAccessDenied
	}

	event := &AuditEvent{
		EventType: eventType,
		UserID:    userID,
		TenantID:  tenantID,
		Resource:  resource,
		Action:    action,
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// LogRoleAssignment logs a role assignment event
func (am *AuditManager) LogRoleAssignment(ctx context.Context, userID, tenantID, roleName, assignedBy, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	eventType := AuditEventTypeRoleAssigned
	if result == AuditResultFailure {
		eventType = AuditEventTypeRoleUnassigned
	}

	event := &AuditEvent{
		EventType: eventType,
		UserID:    userID,
		TenantID:  tenantID,
		Resource:  roleName,
		Action:    "assign",
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
		Metadata: map[string]interface{}{
			"assigned_by": assignedBy,
		},
	}
	return am.LogEvent(ctx, event)
}

// LogPermissionChange logs a permission change event
func (am *AuditManager) LogPermissionChange(ctx context.Context, userID, tenantID, permission, resource, action, changedBy, ipAddress string, result AuditResult, reason string, details map[string]interface{}) error {
	eventType := AuditEventTypePermissionGranted
	if result == AuditResultFailure {
		eventType = AuditEventTypePermissionRevoked
	}

	event := &AuditEvent{
		EventType: eventType,
		UserID:    userID,
		TenantID:  tenantID,
		Resource:  resource,
		Action:    action,
		IPAddress: ipAddress,
		Result:    result,
		Reason:    reason,
		Details:   details,
		Metadata: map[string]interface{}{
			"permission": permission,
			"changed_by": changedBy,
		},
	}
	return am.LogEvent(ctx, event)
}

// LogSuspiciousActivity logs a suspicious activity event
func (am *AuditManager) LogSuspiciousActivity(ctx context.Context, userID, tenantID, ipAddress, userAgent, activity string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: AuditEventTypeSuspiciousActivity,
		UserID:    userID,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    AuditResultWarning,
		Reason:    activity,
		Details:   details,
	}
	return am.LogEvent(ctx, event)
}

// shouldLogEvent checks if an event type should be logged
func (am *AuditManager) shouldLogEvent(eventType AuditEventType) bool {
	for _, logEvent := range am.config.LogEvents {
		if string(eventType) == logEvent {
			return true
		}
	}
	return false
}

// worker processes audit events in the background
func (am *AuditManager) worker() {
	for {
		select {
		case event := <-am.events:
			am.processEvent(event)
		case <-am.stop:
			// Process remaining events
			for {
				select {
				case event := <-am.events:
					am.processEvent(event)
				default:
					return
				}
			}
		}
	}
}

// processEvent processes a single audit event
func (am *AuditManager) processEvent(event *AuditEvent) {
	switch am.config.StorageType {
	case "file":
		am.writeToFile(event)
	case "database":
		am.writeToDatabase(event)
	case "external":
		am.writeToExternal(event)
	default:
		am.logger.WithField("storage_type", am.config.StorageType).Error("Unknown audit storage type")
	}
}

// writeToFile writes audit event to file
func (am *AuditManager) writeToFile(event *AuditEvent) {
	filePath, ok := am.config.StorageConfig["file_path"].(string)
	if !ok {
		filePath = "/var/log/auth/audit.log"
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		am.logger.WithError(err).Error("Failed to create audit log directory")
		return
	}

	// Open file for appending
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		am.logger.WithError(err).Error("Failed to open audit log file")
		return
	}
	defer file.Close()

	// Write event as JSON
	data, err := json.Marshal(event)
	if err != nil {
		am.logger.WithError(err).Error("Failed to marshal audit event")
		return
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		am.logger.WithError(err).Error("Failed to write audit event to file")
		return
	}
}

// writeToDatabase writes audit event to database
func (am *AuditManager) writeToDatabase(event *AuditEvent) {
	// This would be implemented based on your database choice
	am.logger.WithField("event_id", event.ID).Debug("Writing audit event to database")
}

// writeToExternal writes audit event to external service
func (am *AuditManager) writeToExternal(event *AuditEvent) {
	// This would be implemented based on your external service choice
	am.logger.WithField("event_id", event.ID).Debug("Writing audit event to external service")
}

// Stop stops the audit manager
func (am *AuditManager) Stop() {
	close(am.stop)
}

// GetAuditEvents retrieves audit events (implementation depends on storage type)
func (am *AuditManager) GetAuditEvents(ctx context.Context, filter *AuditEventFilter) ([]*AuditEvent, error) {
	// This would be implemented based on your storage choice
	return []*AuditEvent{}, nil
}

// AuditEventFilter represents a filter for audit events
type AuditEventFilter struct {
	UserID    string         `json:"user_id,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	EventType AuditEventType `json:"event_type,omitempty"`
	Result    AuditResult    `json:"result,omitempty"`
	StartTime time.Time      `json:"start_time,omitempty"`
	EndTime   time.Time      `json:"end_time,omitempty"`
	IPAddress string         `json:"ip_address,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	Action    string         `json:"action,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// CleanupOldEvents removes old audit events based on retention policy
func (am *AuditManager) CleanupOldEvents(ctx context.Context) error {
	if am.config.RetentionDays <= 0 {
		return nil
	}

	cutoffTime := time.Now().UTC().AddDate(0, 0, -am.config.RetentionDays)

	// This would be implemented based on your storage choice
	am.logger.WithField("cutoff_time", cutoffTime).Info("Cleaning up old audit events")

	return nil
}
