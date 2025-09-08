package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/siakad/microservices/libs/events"
	"github.com/siakad/microservices/libs/messaging"
	"github.com/siakad/microservices/libs/tracing"
	"github.com/siakad/microservices/libs/utils"
)

// EventDrivenService demonstrates event-driven architecture
type EventDrivenService struct {
	eventService *events.EventSourcingService
	rabbitMQ     *messaging.RabbitMQManager
	kafka        *messaging.KafkaManager
	tracer       *tracing.TracerManager
	logger       *logrus.Logger
}

// UserCreatedHandler handles user created events
type UserCreatedHandler struct {
	logger *logrus.Logger
}

// HandleEvent implements the EventHandler interface
func (h *UserCreatedHandler) HandleEvent(ctx context.Context, event *events.Event) error {
	h.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"user_id":    event.Payload["user_id"],
		"email":      event.Payload["email"],
	}).Info("Handling user created event")

	// Process the event (e.g., send welcome email, create user profile, etc.)
	time.Sleep(100 * time.Millisecond) // Simulate processing

	return nil
}

// GetEventTypes returns the event types this handler can handle
func (h *UserCreatedHandler) GetEventTypes() []events.EventType {
	return []events.EventType{events.EventTypeUserCreated}
}

// NewEventDrivenService creates a new event-driven service
func NewEventDrivenService() (*EventDrivenService, error) {
	logger := logrus.New()

	// Initialize tracing
	tracer, err := tracing.NewTracerManager(&tracing.TracingConfig{
		ServiceName:    "event-driven-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		JaegerEndpoint: "http://localhost:14268/api/traces",
		SampleRate:     1.0,
		Enabled:        true,
	}, logger)
	if err != nil {
		return nil, err
	}

	// Initialize RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQManager(&messaging.RabbitMQConfig{
		URL:      "amqp://guest:guest@localhost:5672/",
		Exchange: "siakad.events",
		Queue:    "user-events",
	}, logger)
	if err != nil {
		return nil, err
	}

	// Initialize Kafka
	kafka := messaging.NewKafkaManager(&messaging.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "event-driven-service",
		Topic:   "user-events",
	}, logger)

	// Initialize event system (simplified - in real implementation, you'd have proper stores)
	eventStore := &MockEventStore{}
	outboxStore := &MockOutboxStore{}
	dispatcher := events.NewEventDispatcher(logger)
	publisher := events.NewEventPublisher(logger)

	eventService := events.NewEventSourcingService(
		eventStore,
		outboxStore,
		dispatcher,
		publisher,
		logger,
	)

	return &EventDrivenService{
		eventService: eventService,
		rabbitMQ:     rabbitMQ,
		kafka:        kafka,
		tracer:       tracer,
		logger:       logger,
	}, nil
}

// Start starts the event-driven service
func (s *EventDrivenService) Start(ctx context.Context) error {
	s.logger.Info("Starting event-driven service")

	// Register event handlers
	userCreatedHandler := &UserCreatedHandler{logger: s.logger}
	s.eventService.dispatcher.RegisterHandler(events.EventTypeUserCreated, userCreatedHandler)

	// Start RabbitMQ consumer
	err := s.rabbitMQ.SubscribeToQueue(ctx, "user-events", s.handleRabbitMQMessage)
	if err != nil {
		return err
	}

	// Start Kafka consumer
	err = s.kafka.SubscribeToTopic(ctx, "user-events", s.handleKafkaMessage)
	if err != nil {
		return err
	}

	// Start outbox event processing
	go s.eventService.ProcessOutboxEvents(ctx, 100)

	// Example: Create and publish an event
	s.createAndPublishUserEvent(ctx)

	s.logger.Info("Event-driven service started successfully")
	return nil
}

// createAndPublishUserEvent creates and publishes a user created event
func (s *EventDrivenService) createAndPublishUserEvent(ctx context.Context) {
	// Start a span for tracing
	ctx, span := s.tracer.StartSpan(ctx, "create_user_event")
	defer span.End()

	// Create event
	event := events.CreateEvent(
		utils.GenerateUUID(), // tenant ID
		utils.GenerateUUID(), // user ID
		events.AggregateTypeUser,
		events.EventTypeUserCreated,
		map[string]interface{}{
			"user_id":    utils.GenerateUUID(),
			"email":      "user@example.com",
			"name":       "Example User",
			"created_at": time.Now(),
		},
	)

	// Add metadata
	event.Metadata["source"] = "event-driven-service"
	event.Metadata["version"] = "1.0"

	// Publish and store event
	err := s.eventService.PublishAndStore(ctx, event)
	if err != nil {
		s.logger.WithError(err).Error("Failed to publish and store event")
		return
	}

	// Publish to RabbitMQ
	message := messaging.CreateMessage(
		"user.created",
		"event-driven-service",
		"notification-service",
		event.Payload,
	)
	message.SetExpiration(24 * time.Hour)

	err = s.rabbitMQ.PublishMessage(ctx, "user.created", message)
	if err != nil {
		s.logger.WithError(err).Error("Failed to publish message to RabbitMQ")
	}

	// Publish to Kafka
	kafkaMessage := messaging.CreateKafkaMessage(
		"user.created",
		"event-driven-service",
		"analytics-service",
		event.Payload,
	)
	kafkaMessage.SetExpiration(24 * time.Hour)

	err = s.kafka.PublishMessage(ctx, "user-events", kafkaMessage)
	if err != nil {
		s.logger.WithError(err).Error("Failed to publish message to Kafka")
	}

	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"user_id":    event.Payload["user_id"],
	}).Info("User event created and published")
}

// handleRabbitMQMessage handles messages from RabbitMQ
func (s *EventDrivenService) handleRabbitMQMessage(ctx context.Context, message *messaging.Message) error {
	s.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"message_type": message.Type,
		"source":       message.Source,
		"target":       message.Target,
	}).Info("Received RabbitMQ message")

	// Process the message
	switch message.Type {
	case "user.created":
		return s.handleUserCreatedEvent(ctx, message.Payload)
	case "user.updated":
		return s.handleUserUpdatedEvent(ctx, message.Payload)
	default:
		s.logger.WithField("message_type", message.Type).Warn("Unknown message type")
		return nil
	}
}

// handleKafkaMessage handles messages from Kafka
func (s *EventDrivenService) handleKafkaMessage(ctx context.Context, message *messaging.KafkaMessage) error {
	s.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"message_type": message.Type,
		"source":       message.Source,
		"target":       message.Target,
		"partition":    message.Partition,
		"offset":       message.Offset,
	}).Info("Received Kafka message")

	// Process the message
	switch message.Type {
	case "user.created":
		return s.handleUserCreatedEvent(ctx, message.Payload)
	case "user.updated":
		return s.handleUserUpdatedEvent(ctx, message.Payload)
	default:
		s.logger.WithField("message_type", message.Type).Warn("Unknown message type")
		return nil
	}
}

// handleUserCreatedEvent handles user created events
func (s *EventDrivenService) handleUserCreatedEvent(ctx context.Context, payload map[string]interface{}) error {
	// Start a span for tracing
	ctx, span := s.tracer.StartSpan(ctx, "handle_user_created")
	defer span.End()

	// Add span attributes
	span.SetAttributes(map[string]interface{}{
		"user_id": payload["user_id"],
		"email":   payload["email"],
		"event":   "user.created",
	})

	// Process the user created event
	s.logger.WithFields(logrus.Fields{
		"user_id": payload["user_id"],
		"email":   payload["email"],
	}).Info("Processing user created event")

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// Add event to span
	span.AddEvent("user_created_processed", map[string]interface{}{
		"processing_time": "50ms",
	})

	return nil
}

// handleUserUpdatedEvent handles user updated events
func (s *EventDrivenService) handleUserUpdatedEvent(ctx context.Context, payload map[string]interface{}) error {
	// Start a span for tracing
	ctx, span := s.tracer.StartSpan(ctx, "handle_user_updated")
	defer span.End()

	// Add span attributes
	span.SetAttributes(map[string]interface{}{
		"user_id": payload["user_id"],
		"email":   payload["email"],
		"event":   "user.updated",
	})

	// Process the user updated event
	s.logger.WithFields(logrus.Fields{
		"user_id": payload["user_id"],
		"email":   payload["email"],
	}).Info("Processing user updated event")

	// Simulate processing time
	time.Sleep(30 * time.Millisecond)

	return nil
}

// Stop stops the event-driven service
func (s *EventDrivenService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping event-driven service")

	// Close RabbitMQ connection
	if err := s.rabbitMQ.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close RabbitMQ connection")
	}

	// Close Kafka connection
	if err := s.kafka.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close Kafka connection")
	}

	// Close tracer
	if err := s.tracer.Close(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to close tracer")
	}

	s.logger.Info("Event-driven service stopped")
	return nil
}

// Mock implementations for demonstration
type MockEventStore struct{}

func (m *MockEventStore) SaveEvent(ctx context.Context, event *events.Event) error {
	return nil
}

func (m *MockEventStore) GetEvents(ctx context.Context, aggregateID uuid.UUID) ([]*events.Event, error) {
	return []*events.Event{}, nil
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType events.EventType, limit, offset int) ([]*events.Event, error) {
	return []*events.Event{}, nil
}

type MockOutboxStore struct{}

func (m *MockOutboxStore) SaveOutboxEvent(ctx context.Context, event *events.OutboxEvent) error {
	return nil
}

func (m *MockOutboxStore) GetUnprocessedEvents(ctx context.Context, limit int) ([]*events.OutboxEvent, error) {
	return []*events.OutboxEvent{}, nil
}

func (m *MockOutboxStore) MarkEventAsProcessed(ctx context.Context, eventID uuid.UUID) error {
	return nil
}

func (m *MockOutboxStore) DeleteProcessedEvents(ctx context.Context, olderThan time.Time) error {
	return nil
}

func main() {
	// Create event-driven service
	service, err := NewEventDrivenService()
	if err != nil {
		log.Fatal(err)
	}

	// Create context
	ctx := context.Background()

	// Start service
	if err := service.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Stop service
	if err := service.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}
