package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/event/gateway"
	"github.com/anasamu/microservices-library-go/event/providers/kafka"
	"github.com/anasamu/microservices-library-go/event/providers/nats"
	"github.com/anasamu/microservices-library-go/event/providers/postgresql"
	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create event sourcing manager
	managerConfig := &gateway.ManagerConfig{
		DefaultProvider:   "postgresql",
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		Timeout:           30 * time.Second,
		MaxEventSize:      1024 * 1024, // 1MB
		MaxBatchSize:      100,
		SnapshotThreshold: 1000,
		RetentionPeriod:   365 * 24 * time.Hour, // 1 year
		Compression:       false,
		Encryption:        false,
		Metadata:          make(map[string]string),
	}

	manager := gateway.NewEventSourcingManager(managerConfig, logger)

	// Example 1: PostgreSQL Provider
	fmt.Println("=== PostgreSQL Provider Example ===")
	if err := examplePostgreSQL(manager, logger); err != nil {
		logger.WithError(err).Error("PostgreSQL example failed")
	}

	// Example 2: Kafka Provider
	fmt.Println("\n=== Kafka Provider Example ===")
	if err := exampleKafka(manager, logger); err != nil {
		logger.WithError(err).Error("Kafka example failed")
	}

	// Example 3: NATS Provider
	fmt.Println("\n=== NATS Provider Example ===")
	if err := exampleNATS(manager, logger); err != nil {
		logger.WithError(err).Error("NATS example failed")
	}

	// Example 4: Multiple Providers
	fmt.Println("\n=== Multiple Providers Example ===")
	if err := exampleMultipleProviders(manager, logger); err != nil {
		logger.WithError(err).Error("Multiple providers example failed")
	}

	// Close manager
	if err := manager.Close(); err != nil {
		logger.WithError(err).Error("Failed to close manager")
	}
}

func examplePostgreSQL(manager *gateway.EventSourcingManager, logger *logrus.Logger) error {
	// Create PostgreSQL provider
	pgConfig := &postgresql.PostgreSQLConfig{
		Host:            "localhost",
		Port:            5432,
		Database:        "eventsourcing",
		Username:        "postgres",
		Password:        "password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		Metadata:        map[string]string{"provider": "postgresql"},
	}

	pgProvider, err := postgresql.NewPostgreSQLProvider(pgConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL provider: %w", err)
	}

	// Register provider
	if err := manager.RegisterProvider(pgProvider); err != nil {
		return fmt.Errorf("failed to register PostgreSQL provider: %w", err)
	}

	// Connect to PostgreSQL
	ctx := context.Background()
	if err := pgProvider.Connect(ctx); err != nil {
		logger.WithError(err).Warn("Failed to connect to PostgreSQL (this is expected if PostgreSQL is not running)")
		return nil
	}

	// Create a stream
	streamRequest := &types.CreateStreamRequest{
		StreamID:      "user-123",
		Name:          "User Events",
		AggregateID:   "user-123",
		AggregateType: "User",
		Metadata: map[string]interface{}{
			"description": "User aggregate events",
			"version":     "1.0",
		},
	}

	if err := manager.CreateStream(ctx, streamRequest); err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Append events
	events := []*types.AppendEventRequest{
		{
			StreamID:      "user-123",
			EventType:     "UserCreated",
			EventData:     map[string]interface{}{"name": "John Doe", "email": "john@example.com"},
			EventMetadata: map[string]interface{}{"source": "api", "version": "1.0"},
			AggregateID:   "user-123",
			AggregateType: "User",
			UserID:        "admin",
		},
		{
			StreamID:      "user-123",
			EventType:     "UserUpdated",
			EventData:     map[string]interface{}{"name": "John Smith", "email": "john.smith@example.com"},
			EventMetadata: map[string]interface{}{"source": "api", "version": "1.0"},
			AggregateID:   "user-123",
			AggregateType: "User",
			UserID:        "admin",
		},
	}

	for _, eventRequest := range events {
		response, err := manager.AppendEvent(ctx, eventRequest)
		if err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}
		logger.WithField("event_id", response.EventID).Info("Event appended")
	}

	// Retrieve events
	getEventsRequest := &types.GetEventsRequest{
		StreamID: "user-123",
		Limit:    10,
	}

	eventsResponse, err := manager.GetEvents(ctx, getEventsRequest)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	logger.WithField("event_count", len(eventsResponse.Events)).Info("Retrieved events")

	// Create a snapshot
	snapshotRequest := &types.CreateSnapshotRequest{
		StreamID:      "user-123",
		AggregateID:   "user-123",
		AggregateType: "User",
		Version:       2,
		Data: map[string]interface{}{
			"name":  "John Smith",
			"email": "john.smith@example.com",
			"id":    "user-123",
		},
		Metadata: map[string]interface{}{"snapshot_type": "user_state"},
	}

	snapshotResponse, err := manager.CreateSnapshot(ctx, snapshotRequest)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	logger.WithField("snapshot_id", snapshotResponse.SnapshotID).Info("Snapshot created")

	// Get stream info
	streamInfoRequest := &types.GetStreamInfoRequest{
		StreamID: "user-123",
	}

	streamInfo, err := manager.GetStreamInfo(ctx, streamInfoRequest)
	if err != nil {
		return fmt.Errorf("failed to get stream info: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"stream_id":   streamInfo.ID,
		"version":     streamInfo.Version,
		"event_count": streamInfo.EventCount,
		"created_at":  streamInfo.CreatedAt,
		"updated_at":  streamInfo.UpdatedAt,
	}).Info("Stream info")

	return nil
}

func exampleKafka(manager *gateway.EventSourcingManager, logger *logrus.Logger) error {
	// Create Kafka provider
	kafkaConfig := &kafka.KafkaConfig{
		Brokers:         []string{"localhost:9092"},
		Topic:           "events",
		GroupID:         "event-sourcing-group",
		CompressionType: "gzip",
		BatchSize:       10,
		BatchTimeout:    time.Second,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		Metadata:        map[string]string{"provider": "kafka"},
	}

	kafkaProvider, err := kafka.NewKafkaProvider(kafkaConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka provider: %w", err)
	}

	// Register provider
	if err := manager.RegisterProvider(kafkaProvider); err != nil {
		return fmt.Errorf("failed to register Kafka provider: %w", err)
	}

	// Connect to Kafka
	ctx := context.Background()
	if err := kafkaProvider.Connect(ctx); err != nil {
		logger.WithError(err).Warn("Failed to connect to Kafka (this is expected if Kafka is not running)")
		return nil
	}

	// Append events
	events := []*types.AppendEventRequest{
		{
			StreamID:      "order-456",
			EventType:     "OrderCreated",
			EventData:     map[string]interface{}{"order_id": "order-456", "amount": 99.99, "currency": "USD"},
			EventMetadata: map[string]interface{}{"source": "api", "version": "1.0"},
			AggregateID:   "order-456",
			AggregateType: "Order",
			UserID:        "customer-123",
		},
		{
			StreamID:      "order-456",
			EventType:     "OrderPaid",
			EventData:     map[string]interface{}{"payment_method": "credit_card", "transaction_id": "tx-789"},
			EventMetadata: map[string]interface{}{"source": "payment_service", "version": "1.0"},
			AggregateID:   "order-456",
			AggregateType: "Order",
			UserID:        "customer-123",
		},
	}

	for _, eventRequest := range events {
		response, err := manager.AppendEventWithProvider(ctx, "kafka", eventRequest)
		if err != nil {
			return fmt.Errorf("failed to append event to Kafka: %w", err)
		}
		logger.WithField("event_id", response.EventID).Info("Event appended to Kafka")
	}

	// Retrieve events
	getEventsRequest := &types.GetEventsRequest{
		StreamID: "order-456",
		Limit:    10,
	}

	eventsResponse, err := manager.GetEventsWithProvider(ctx, "kafka", getEventsRequest)
	if err != nil {
		return fmt.Errorf("failed to get events from Kafka: %w", err)
	}

	logger.WithField("event_count", len(eventsResponse.Events)).Info("Retrieved events from Kafka")

	return nil
}

func exampleNATS(manager *gateway.EventSourcingManager, logger *logrus.Logger) error {
	// Create NATS provider
	natsConfig := &nats.NATSConfig{
		URL:             "nats://localhost:4222",
		Subject:         "events.>",
		StreamName:      "EVENTS",
		GroupID:         "event-sourcing-group",
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		Timeout:         30 * time.Second,
		RetentionPolicy: "limits",
		MaxAge:          24 * time.Hour,
		MaxBytes:        1024 * 1024 * 1024, // 1GB
		MaxEvents:       1000000,
		Metadata:        map[string]string{"provider": "nats"},
	}

	natsProvider, err := nats.NewNATSProvider(natsConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create NATS provider: %w", err)
	}

	// Register provider
	if err := manager.RegisterProvider(natsProvider); err != nil {
		return fmt.Errorf("failed to register NATS provider: %w", err)
	}

	// Connect to NATS
	ctx := context.Background()
	if err := natsProvider.Connect(ctx); err != nil {
		logger.WithError(err).Warn("Failed to connect to NATS (this is expected if NATS is not running)")
		return nil
	}

	// Append events
	events := []*types.AppendEventRequest{
		{
			StreamID:      "product-789",
			EventType:     "ProductCreated",
			EventData:     map[string]interface{}{"name": "Laptop", "price": 999.99, "category": "Electronics"},
			EventMetadata: map[string]interface{}{"source": "admin_panel", "version": "1.0"},
			AggregateID:   "product-789",
			AggregateType: "Product",
			UserID:        "admin",
		},
		{
			StreamID:      "product-789",
			EventType:     "ProductUpdated",
			EventData:     map[string]interface{}{"price": 899.99, "discount": 0.1},
			EventMetadata: map[string]interface{}{"source": "admin_panel", "version": "1.0"},
			AggregateID:   "product-789",
			AggregateType: "Product",
			UserID:        "admin",
		},
	}

	for _, eventRequest := range events {
		response, err := manager.AppendEventWithProvider(ctx, "nats", eventRequest)
		if err != nil {
			return fmt.Errorf("failed to append event to NATS: %w", err)
		}
		logger.WithField("event_id", response.EventID).Info("Event appended to NATS")
	}

	// Retrieve events
	getEventsRequest := &types.GetEventsRequest{
		StreamID: "product-789",
		Limit:    10,
	}

	eventsResponse, err := manager.GetEventsWithProvider(ctx, "nats", getEventsRequest)
	if err != nil {
		return fmt.Errorf("failed to get events from NATS: %w", err)
	}

	logger.WithField("event_count", len(eventsResponse.Events)).Info("Retrieved events from NATS")

	return nil
}

func exampleMultipleProviders(manager *gateway.EventSourcingManager, logger *logrus.Logger) error {
	ctx := context.Background()

	// Get provider information
	providerInfo := manager.GetProviderInfo()
	logger.WithField("providers", providerInfo).Info("Registered providers")

	// List all providers
	providers := manager.ListProviders()
	logger.WithField("provider_names", providers).Info("Available providers")

	// Health check all providers
	healthResults := manager.HealthCheck(ctx)
	for providerName, err := range healthResults {
		if err != nil {
			logger.WithError(err).WithField("provider", providerName).Warn("Provider health check failed")
		} else {
			logger.WithField("provider", providerName).Info("Provider is healthy")
		}
	}

	// Get statistics from all providers
	stats, err := manager.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	for providerName, stat := range stats {
		logger.WithFields(logrus.Fields{
			"provider":      providerName,
			"total_events":  stat.TotalEvents,
			"total_streams": stat.TotalStreams,
			"uptime":        stat.Uptime,
		}).Info("Provider statistics")
	}

	// Example of batch operations
	batchRequest := &types.AppendEventsBatchRequest{
		StreamID: "batch-stream",
		Events: []*types.Event{
			{
				EventType:     "BatchEvent1",
				EventData:     map[string]interface{}{"message": "First batch event"},
				EventMetadata: map[string]interface{}{"batch": true},
				AggregateID:   "batch-stream",
				AggregateType: "Batch",
			},
			{
				EventType:     "BatchEvent2",
				EventData:     map[string]interface{}{"message": "Second batch event"},
				EventMetadata: map[string]interface{}{"batch": true},
				AggregateID:   "batch-stream",
				AggregateType: "Batch",
			},
		},
	}

	// Try to append batch to the first available provider
	if len(providers) > 0 {
		batchResponse, err := manager.AppendEventsBatchWithProvider(ctx, providers[0], batchRequest)
		if err != nil {
			logger.WithError(err).Warn("Failed to append batch events")
		} else {
			logger.WithFields(logrus.Fields{
				"appended_count": batchResponse.AppendedCount,
				"failed_count":   batchResponse.FailedCount,
				"provider":       providers[0],
			}).Info("Batch events processed")
		}
	}

	return nil
}

// Example of a simple event sourcing aggregate
type UserAggregate struct {
	ID      string
	Name    string
	Email   string
	Version int64
	Events  []*types.Event
	manager *gateway.EventSourcingManager
}

func NewUserAggregate(id string, manager *gateway.EventSourcingManager) *UserAggregate {
	return &UserAggregate{
		ID:      id,
		Version: 0,
		Events:  make([]*types.Event, 0),
		manager: manager,
	}
}

func (u *UserAggregate) CreateUser(ctx context.Context, name, email string) error {
	eventRequest := &types.AppendEventRequest{
		StreamID:        u.ID,
		EventType:       "UserCreated",
		EventData:       map[string]interface{}{"name": name, "email": email},
		EventMetadata:   map[string]interface{}{"source": "aggregate"},
		AggregateID:     u.ID,
		AggregateType:   "User",
		ExpectedVersion: &u.Version,
	}

	response, err := u.manager.AppendEvent(ctx, eventRequest)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update aggregate state
	u.Name = name
	u.Email = email
	u.Version = response.Version

	return nil
}

func (u *UserAggregate) UpdateUser(ctx context.Context, name, email string) error {
	eventRequest := &types.AppendEventRequest{
		StreamID:        u.ID,
		EventType:       "UserUpdated",
		EventData:       map[string]interface{}{"name": name, "email": email},
		EventMetadata:   map[string]interface{}{"source": "aggregate"},
		AggregateID:     u.ID,
		AggregateType:   "User",
		ExpectedVersion: &u.Version,
	}

	response, err := u.manager.AppendEvent(ctx, eventRequest)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Update aggregate state
	u.Name = name
	u.Email = email
	u.Version = response.Version

	return nil
}

func (u *UserAggregate) LoadFromEvents(ctx context.Context) error {
	request := &types.GetEventsByStreamRequest{
		StreamID: u.ID,
		Limit:    1000,
	}

	response, err := u.manager.GetEventsByStream(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to load events: %w", err)
	}

	// Replay events to rebuild aggregate state
	for _, event := range response.Events {
		u.Events = append(u.Events, event)
		u.Version = event.Version

		switch event.EventType {
		case "UserCreated":
			if name, ok := event.EventData["name"].(string); ok {
				u.Name = name
			}
			if email, ok := event.EventData["email"].(string); ok {
				u.Email = email
			}
		case "UserUpdated":
			if name, ok := event.EventData["name"].(string); ok {
				u.Name = name
			}
			if email, ok := event.EventData["email"].(string); ok {
				u.Email = email
			}
		}
	}

	return nil
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
