package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EventType represents the type of event
type EventType string

const (
	// User events
	EventTypeUserCreated     EventType = "user.created"
	EventTypeUserUpdated     EventType = "user.updated"
	EventTypeUserDeleted     EventType = "user.deleted"
	EventTypeUserActivated   EventType = "user.activated"
	EventTypeUserDeactivated EventType = "user.deactivated"

	// Profile events
	EventTypeProfileCreated EventType = "profile.created"
	EventTypeProfileUpdated EventType = "profile.updated"

	// Contact events
	EventTypeContactCreated EventType = "contact.created"
	EventTypeContactUpdated EventType = "contact.updated"
	EventTypeContactDeleted EventType = "contact.deleted"

	// Address events
	EventTypeAddressCreated EventType = "address.created"
	EventTypeAddressUpdated EventType = "address.updated"
	EventTypeAddressDeleted EventType = "address.deleted"

	// Role events
	EventTypeRoleAssigned   EventType = "role.assigned"
	EventTypeRoleUnassigned EventType = "role.unassigned"

	// Group events
	EventTypeGroupCreated EventType = "group.created"
	EventTypeGroupUpdated EventType = "group.updated"
	EventTypeGroupDeleted EventType = "group.deleted"
	EventTypeGroupJoined  EventType = "group.joined"
	EventTypeGroupLeft    EventType = "group.left"
)

// AggregateType represents the type of aggregate
type AggregateType string

const (
	AggregateTypeUser    AggregateType = "user"
	AggregateTypeProfile AggregateType = "profile"
	AggregateTypeContact AggregateType = "contact"
	AggregateTypeAddress AggregateType = "address"
	AggregateTypeRole    AggregateType = "role"
	AggregateTypeGroup   AggregateType = "group"
)

// Event represents a domain event
type Event struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	AggregateID   uuid.UUID              `json:"aggregate_id"`
	AggregateType AggregateType          `json:"aggregate_type"`
	EventType     EventType              `json:"event_type"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	Version       int                    `json:"version"`
}

// OutboxEvent represents an outbox event for event sourcing
type OutboxEvent struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	TenantID      uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	AggregateID   uuid.UUID              `json:"aggregate_id" db:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type" db:"aggregate_type"`
	EventType     string                 `json:"event_type" db:"event_type"`
	Payload       map[string]interface{} `json:"payload" db:"payload"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	Processed     bool                   `json:"processed" db:"processed"`
}

// EventPublisher handles event publishing
type EventPublisher struct {
	logger *logrus.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(logger *logrus.Logger) *EventPublisher {
	return &EventPublisher{
		logger: logger,
	}
}

// PublishEvent publishes an event
func (p *EventPublisher) PublishEvent(ctx context.Context, event *Event) error {
	// Log the event
	p.logger.WithFields(logrus.Fields{
		"event_id":       event.ID,
		"tenant_id":      event.TenantID,
		"aggregate_id":   event.AggregateID,
		"aggregate_type": event.AggregateType,
		"event_type":     event.EventType,
		"version":        event.Version,
	}).Info("Event published")

	// Here you would typically publish to a message queue (RabbitMQ, Kafka, etc.)
	// For now, we'll just log it
	return nil
}

// CreateEvent creates a new event
func CreateEvent(tenantID, aggregateID uuid.UUID, aggregateType AggregateType, eventType EventType, payload map[string]interface{}) *Event {
	return &Event{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payload,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		Version:       1,
	}
}

// ToOutboxEvent converts an event to an outbox event
func (e *Event) ToOutboxEvent() *OutboxEvent {
	return &OutboxEvent{
		ID:            e.ID,
		TenantID:      e.TenantID,
		AggregateID:   e.AggregateID,
		AggregateType: string(e.AggregateType),
		EventType:     string(e.EventType),
		Payload:       e.Payload,
		CreatedAt:     e.CreatedAt,
		Processed:     false,
	}
}

// ToEvent converts an outbox event to an event
func (oe *OutboxEvent) ToEvent() *Event {
	return &Event{
		ID:            oe.ID,
		TenantID:      oe.TenantID,
		AggregateID:   oe.AggregateID,
		AggregateType: AggregateType(oe.AggregateType),
		EventType:     EventType(oe.EventType),
		Payload:       oe.Payload,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     oe.CreatedAt,
		Version:       1,
	}
}

// EventHandler handles events
type EventHandler interface {
	HandleEvent(ctx context.Context, event *Event) error
	GetEventTypes() []EventType
}

// EventDispatcher dispatches events to handlers
type EventDispatcher struct {
	handlers map[EventType][]EventHandler
	logger   *logrus.Logger
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(logger *logrus.Logger) *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[EventType][]EventHandler),
		logger:   logger,
	}
}

// RegisterHandler registers an event handler
func (d *EventDispatcher) RegisterHandler(eventType EventType, handler EventHandler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
	d.logger.WithField("event_type", eventType).Info("Event handler registered")
}

// Dispatch dispatches an event to all registered handlers
func (d *EventDispatcher) Dispatch(ctx context.Context, event *Event) error {
	handlers, exists := d.handlers[event.EventType]
	if !exists {
		d.logger.WithField("event_type", event.EventType).Debug("No handlers registered for event type")
		return nil
	}

	for _, handler := range handlers {
		if err := handler.HandleEvent(ctx, event); err != nil {
			d.logger.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.EventType,
				"error":      err,
			}).Error("Failed to handle event")
			return fmt.Errorf("failed to handle event: %w", err)
		}
	}

	d.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"handlers":   len(handlers),
	}).Info("Event dispatched successfully")

	return nil
}

// EventStore stores events
type EventStore interface {
	SaveEvent(ctx context.Context, event *Event) error
	GetEvents(ctx context.Context, aggregateID uuid.UUID) ([]*Event, error)
	GetEventsByType(ctx context.Context, eventType EventType, limit, offset int) ([]*Event, error)
}

// OutboxStore stores outbox events
type OutboxStore interface {
	SaveOutboxEvent(ctx context.Context, event *OutboxEvent) error
	GetUnprocessedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkEventAsProcessed(ctx context.Context, eventID uuid.UUID) error
	DeleteProcessedEvents(ctx context.Context, olderThan time.Time) error
}

// EventSourcingService handles event sourcing operations
type EventSourcingService struct {
	eventStore  EventStore
	outboxStore OutboxStore
	dispatcher  *EventDispatcher
	publisher   *EventPublisher
	logger      *logrus.Logger
}

// NewEventSourcingService creates a new event sourcing service
func NewEventSourcingService(
	eventStore EventStore,
	outboxStore OutboxStore,
	dispatcher *EventDispatcher,
	publisher *EventPublisher,
	logger *logrus.Logger,
) *EventSourcingService {
	return &EventSourcingService{
		eventStore:  eventStore,
		outboxStore: outboxStore,
		dispatcher:  dispatcher,
		publisher:   publisher,
		logger:      logger,
	}
}

// PublishAndStore publishes an event and stores it in the outbox
func (s *EventSourcingService) PublishAndStore(ctx context.Context, event *Event) error {
	// Store in event store
	if err := s.eventStore.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Store in outbox for eventual publishing
	outboxEvent := event.ToOutboxEvent()
	if err := s.outboxStore.SaveOutboxEvent(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	// Dispatch to local handlers
	if err := s.dispatcher.Dispatch(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to dispatch event locally")
		// Don't return error here as the event is already stored
	}

	// Publish to external systems
	if err := s.publisher.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish event externally")
		// Don't return error here as the event is already stored
	}

	return nil
}

// ProcessOutboxEvents processes unprocessed outbox events
func (s *EventSourcingService) ProcessOutboxEvents(ctx context.Context, batchSize int) error {
	events, err := s.outboxStore.GetUnprocessedEvents(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed events: %w", err)
	}

	for _, outboxEvent := range events {
		event := outboxEvent.ToEvent()

		// Publish the event
		if err := s.publisher.PublishEvent(ctx, event); err != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id": event.ID,
				"error":    err,
			}).Error("Failed to publish outbox event")
			continue
		}

		// Mark as processed
		if err := s.outboxStore.MarkEventAsProcessed(ctx, event.ID); err != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id": event.ID,
				"error":    err,
			}).Error("Failed to mark event as processed")
			continue
		}

		s.logger.WithField("event_id", event.ID).Info("Outbox event processed successfully")
	}

	return nil
}

// CleanupProcessedEvents cleans up old processed events
func (s *EventSourcingService) CleanupProcessedEvents(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	return s.outboxStore.DeleteProcessedEvents(ctx, cutoffTime)
}

// GetEventHistory gets the event history for an aggregate
func (s *EventSourcingService) GetEventHistory(ctx context.Context, aggregateID uuid.UUID) ([]*Event, error) {
	return s.eventStore.GetEvents(ctx, aggregateID)
}

// ReplayEvents replays events for an aggregate
func (s *EventSourcingService) ReplayEvents(ctx context.Context, aggregateID uuid.UUID) error {
	events, err := s.eventStore.GetEvents(ctx, aggregateID)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	for _, event := range events {
		if err := s.dispatcher.Dispatch(ctx, event); err != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id":     event.ID,
				"aggregate_id": aggregateID,
				"error":        err,
			}).Error("Failed to replay event")
			return fmt.Errorf("failed to replay event: %w", err)
		}
	}

	s.logger.WithField("aggregate_id", aggregateID).Info("Events replayed successfully")
	return nil
}

// EventBuilder helps build events
type EventBuilder struct {
	event *Event
}

// NewEventBuilder creates a new event builder
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		event: &Event{
			ID:        uuid.New(),
			Payload:   make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			Version:   1,
		},
	}
}

// WithTenantID sets the tenant ID
func (b *EventBuilder) WithTenantID(tenantID uuid.UUID) *EventBuilder {
	b.event.TenantID = tenantID
	return b
}

// WithAggregateID sets the aggregate ID
func (b *EventBuilder) WithAggregateID(aggregateID uuid.UUID) *EventBuilder {
	b.event.AggregateID = aggregateID
	return b
}

// WithAggregateType sets the aggregate type
func (b *EventBuilder) WithAggregateType(aggregateType AggregateType) *EventBuilder {
	b.event.AggregateType = aggregateType
	return b
}

// WithEventType sets the event type
func (b *EventBuilder) WithEventType(eventType EventType) *EventBuilder {
	b.event.EventType = eventType
	return b
}

// WithPayload sets the payload
func (b *EventBuilder) WithPayload(payload map[string]interface{}) *EventBuilder {
	b.event.Payload = payload
	return b
}

// WithMetadata sets the metadata
func (b *EventBuilder) WithMetadata(metadata map[string]interface{}) *EventBuilder {
	b.event.Metadata = metadata
	return b
}

// WithVersion sets the version
func (b *EventBuilder) WithVersion(version int) *EventBuilder {
	b.event.Version = version
	return b
}

// Build builds the event
func (b *EventBuilder) Build() *Event {
	return b.event
}

// ToJSON converts an event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}
