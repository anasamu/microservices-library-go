package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// NATSProvider implements EventSourcingProvider for NATS
type NATSProvider struct {
	conn      *nats.Conn
	js        nats.JetStreamContext
	config    *NATSConfig
	logger    *logrus.Logger
	connected bool
}

// NATSConfig holds NATS-specific configuration
type NATSConfig struct {
	URL             string            `json:"url"`
	Subject         string            `json:"subject"`
	StreamName      string            `json:"stream_name"`
	Username        string            `json:"username"`
	Password        string            `json:"password"`
	Token           string            `json:"token"`
	MaxReconnects   int               `json:"max_reconnects"`
	ReconnectWait   time.Duration     `json:"reconnect_wait"`
	Timeout         time.Duration     `json:"timeout"`
	RetentionPolicy string            `json:"retention_policy"`
	MaxAge          time.Duration     `json:"max_age"`
	MaxBytes        int64             `json:"max_bytes"`
	MaxEvents       int64             `json:"max_events"`
	Metadata        map[string]string `json:"metadata"`
}

// NewNATSProvider creates a new NATS event sourcing provider
func NewNATSProvider(config *NATSConfig, logger *logrus.Logger) (*NATSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.URL == "" {
		config.URL = nats.DefaultURL
	}

	if config.Subject == "" {
		config.Subject = "events.>"
	}

	if config.StreamName == "" {
		config.StreamName = "EVENTS"
	}

	if logger == nil {
		logger = logrus.New()
	}

	provider := &NATSProvider{
		config: config,
		logger: logger,
	}

	return provider, nil
}

// GetName returns the provider name
func (n *NATSProvider) GetName() string {
	return "nats"
}

// GetSupportedFeatures returns the features supported by this provider
func (n *NATSProvider) GetSupportedFeatures() []types.EventSourcingFeature {
	return []types.EventSourcingFeature{
		types.FeatureEventAppend,
		types.FeatureEventRetrieval,
		types.FeatureEventQuery,
		types.FeatureEventFiltering,
		types.FeatureEventReplay,
		types.FeatureEventVersioning,
		types.FeatureEventBatching,
		types.FeatureEventOrdering,
		types.FeatureEventCorrelation,
		types.FeatureEventStreaming,
		types.FeatureEventReplication,
		types.FeatureEventRetention,
	}
}

// GetConnectionInfo returns connection information
func (n *NATSProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     n.config.URL,
		Port:     4222, // Default NATS port
		Database: n.config.StreamName,
		Username: n.config.Username,
		Status:   n.getConnectionStatus(),
		Metadata: n.config.Metadata,
	}
}

// Connect establishes connection to NATS
func (n *NATSProvider) Connect(ctx context.Context) error {
	opts := []nats.Option{
		nats.MaxReconnects(n.config.MaxReconnects),
		nats.ReconnectWait(n.config.ReconnectWait),
		nats.Timeout(n.config.Timeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			n.logger.WithError(err).Warn("NATS disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			n.logger.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			n.logger.Info("NATS connection closed")
		}),
	}

	// Add authentication if provided
	if n.config.Username != "" && n.config.Password != "" {
		opts = append(opts, nats.UserInfo(n.config.Username, n.config.Password))
	} else if n.config.Token != "" {
		opts = append(opts, nats.Token(n.config.Token))
	}

	// Connect to NATS
	conn, err := nats.Connect(n.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	n.conn = conn

	// Create JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}

	n.js = js

	// Create stream if it doesn't exist
	if err := n.createStream(ctx); err != nil {
		conn.Close()
		return fmt.Errorf("failed to create stream: %w", err)
	}

	n.connected = true
	n.logger.Info("Connected to NATS event sourcing provider")
	return nil
}

// Disconnect closes the NATS connection
func (n *NATSProvider) Disconnect(ctx context.Context) error {
	if n.conn != nil {
		n.conn.Close()
	}
	n.connected = false
	return nil
}

// Ping checks if NATS is reachable
func (n *NATSProvider) Ping(ctx context.Context) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Send a ping message
	subject := "ping"
	_, err := n.conn.RequestWithContext(ctx, subject, []byte("ping"))
	if err != nil {
		return fmt.Errorf("failed to ping NATS: %w", err)
	}

	return nil
}

// IsConnected returns true if connected to NATS
func (n *NATSProvider) IsConnected() bool {
	return n.connected && n.conn != nil && n.conn.IsConnected()
}

// AppendEvent appends an event to NATS
func (n *NATSProvider) AppendEvent(ctx context.Context, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	eventID := uuid.New().String()
	now := time.Now()

	// Create event
	event := &types.Event{
		ID:            eventID,
		StreamID:      request.StreamID,
		EventType:     request.EventType,
		EventData:     request.EventData,
		EventMetadata: request.EventMetadata,
		Version:       0, // NATS doesn't maintain version per stream
		Timestamp:     now,
		CorrelationID: request.CorrelationID,
		CausationID:   request.CausationID,
		UserID:        request.UserID,
		TenantID:      request.TenantID,
		AggregateID:   request.AggregateID,
		AggregateType: request.AggregateType,
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create subject for the event
	subject := fmt.Sprintf("%s.%s.%s", n.config.Subject, request.StreamID, request.EventType)

	// Publish event to JetStream
	ack, err := n.js.PublishAsync(subject, eventJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to publish event to NATS: %w", err)
	}

	// Wait for acknowledgment
	select {
	case <-ack.Ok():
		// Success
	case err := <-ack.Err():
		return nil, fmt.Errorf("failed to get acknowledgment from NATS: %w", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return &types.AppendEventResponse{
		EventID:   eventID,
		StreamID:  request.StreamID,
		Version:   0, // NATS doesn't maintain version
		Timestamp: now,
	}, nil
}

// GetEvents retrieves events from NATS
func (n *NATSProvider) GetEvents(ctx context.Context, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	// Create consumer for reading events
	consumerName := fmt.Sprintf("consumer_%s", uuid.New().String())
	subject := n.config.Subject

	// Add stream filter if specified
	if request.StreamID != "" {
		subject = fmt.Sprintf("%s.%s.>", n.config.Subject, request.StreamID)
	}

	_, err := n.js.AddConsumer(n.config.StreamName, &nats.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: subject,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Clean up consumer when done
	defer func() {
		if err := n.js.DeleteConsumer(n.config.StreamName, consumerName); err != nil {
			n.logger.WithError(err).Warn("Failed to delete consumer")
		}
	}()

	// Create subscription
	sub, err := n.js.PullSubscribe(subject, consumerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	defer sub.Unsubscribe()

	var events []*types.Event
	maxMessages := request.Limit
	if maxMessages <= 0 {
		maxMessages = 100 // Default limit
	}

	// Fetch messages
	messages, err := sub.Fetch(maxMessages, nats.MaxWait(time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	for _, msg := range messages {
		// Parse event
		var event types.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			n.logger.WithError(err).Warn("Failed to unmarshal event from NATS")
			msg.Ack()
			continue
		}

		// Apply filters
		if n.matchesFilters(&event, request) {
			events = append(events, &event)
		}

		// Acknowledge message
		msg.Ack()
	}

	return &types.GetEventsResponse{
		Events:     events,
		TotalCount: int64(len(events)),
		HasMore:    false, // NATS doesn't provide total count
	}, nil
}

// GetEventByID retrieves an event by ID (not supported in NATS)
func (n *NATSProvider) GetEventByID(ctx context.Context, request *types.GetEventByIDRequest) (*types.Event, error) {
	return nil, fmt.Errorf("GetEventByID not supported in NATS provider - events are identified by sequence number, not ID")
}

// GetEventsByStream retrieves events for a specific stream
func (n *NATSProvider) GetEventsByStream(ctx context.Context, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	getEventsRequest := &types.GetEventsRequest{
		StreamID: request.StreamID,
		Limit:    request.Limit,
		Options:  request.Options,
	}

	return n.GetEvents(ctx, getEventsRequest)
}

// CreateStream creates a stream (not applicable in NATS)
func (n *NATSProvider) CreateStream(ctx context.Context, request *types.CreateStreamRequest) error {
	// In NATS, streams are implicit - they exist when events are published
	n.logger.WithField("stream_id", request.StreamID).Info("Stream creation requested (NATS uses single stream)")
	return nil
}

// DeleteStream deletes a stream (not applicable in NATS)
func (n *NATSProvider) DeleteStream(ctx context.Context, request *types.DeleteStreamRequest) error {
	// In NATS, we can't delete specific streams from a stream
	n.logger.WithField("stream_id", request.StreamID).Info("Stream deletion requested (not supported in NATS)")
	return nil
}

// StreamExists checks if a stream exists (always true in NATS)
func (n *NATSProvider) StreamExists(ctx context.Context, request *types.StreamExistsRequest) (bool, error) {
	// In NATS, streams always exist when we can publish to the stream
	return n.IsConnected(), nil
}

// ListStreams lists all streams (not applicable in NATS)
func (n *NATSProvider) ListStreams(ctx context.Context) ([]types.StreamInfo, error) {
	// NATS doesn't maintain a list of streams
	return []types.StreamInfo{}, nil
}

// CreateSnapshot creates a snapshot (not supported in NATS)
func (n *NATSProvider) CreateSnapshot(ctx context.Context, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	return nil, fmt.Errorf("snapshots not supported in NATS provider")
}

// GetSnapshot retrieves a snapshot (not supported in NATS)
func (n *NATSProvider) GetSnapshot(ctx context.Context, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	return nil, fmt.Errorf("snapshots not supported in NATS provider")
}

// DeleteSnapshot deletes a snapshot (not supported in NATS)
func (n *NATSProvider) DeleteSnapshot(ctx context.Context, request *types.DeleteSnapshotRequest) error {
	return fmt.Errorf("snapshots not supported in NATS provider")
}

// AppendEventsBatch appends multiple events in a batch
func (n *NATSProvider) AppendEventsBatch(ctx context.Context, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	var failedEvents []*types.Event
	now := time.Now()

	for _, event := range request.Events {
		eventID := uuid.New().String()

		// Create event with ID
		eventWithID := *event
		eventWithID.ID = eventID
		eventWithID.StreamID = request.StreamID
		eventWithID.Timestamp = now

		// Serialize event
		eventJSON, err := json.Marshal(&eventWithID)
		if err != nil {
			failedEvents = append(failedEvents, event)
			continue
		}

		// Create subject for the event
		subject := fmt.Sprintf("%s.%s.%s", n.config.Subject, request.StreamID, event.EventType)

		// Publish event to JetStream
		ack, err := n.js.PublishAsync(subject, eventJSON)
		if err != nil {
			failedEvents = append(failedEvents, event)
			continue
		}

		// Wait for acknowledgment
		select {
		case <-ack.Ok():
			// Success
		case err := <-ack.Err():
			failedEvents = append(failedEvents, event)
			n.logger.WithError(err).Warn("Failed to get acknowledgment for event")
		case <-ctx.Done():
			failedEvents = append(failedEvents, event)
		}
	}

	appendedCount := len(request.Events) - len(failedEvents)

	return &types.AppendEventsBatchResponse{
		AppendedCount: appendedCount,
		FailedCount:   len(failedEvents),
		FailedEvents:  failedEvents,
		StartVersion:  0,
		EndVersion:    0,
	}, nil
}

// GetStreamInfo gets information about a stream (not applicable in NATS)
func (n *NATSProvider) GetStreamInfo(ctx context.Context, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	// NATS doesn't maintain stream metadata
	return &types.StreamInfo{
		ID:            request.StreamID,
		Name:          request.StreamID,
		AggregateID:   request.StreamID,
		AggregateType: "unknown",
		Version:       0,
		EventCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// GetStats returns event sourcing statistics
func (n *NATSProvider) GetStats(ctx context.Context) (*types.EventSourcingStats, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	// Get stream info
	streamInfo, err := n.js.StreamInfo(n.config.StreamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	stats := &types.EventSourcingStats{
		TotalEvents:    int64(streamInfo.State.Msgs),
		TotalStreams:   1, // Single stream in NATS
		TotalSnapshots: 0,
		EventsByType:   make(map[string]int64),
		EventsByStream: make(map[string]int64),
		StorageSize:    int64(streamInfo.State.Bytes),
		Uptime:         time.Since(streamInfo.Created),
		LastUpdate:     time.Now(),
		Provider:       n.GetName(),
	}

	return stats, nil
}

// HealthCheck performs a health check
func (n *NATSProvider) HealthCheck(ctx context.Context) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	return n.Ping(ctx)
}

// Configure configures the provider
func (n *NATSProvider) Configure(config map[string]interface{}) error {
	n.logger.WithField("config", config).Info("NATS provider configuration updated")
	return nil
}

// IsConfigured returns true if the provider is configured
func (n *NATSProvider) IsConfigured() bool {
	return n.config != nil && n.config.URL != "" && n.config.StreamName != ""
}

// Close closes the provider
func (n *NATSProvider) Close() error {
	return n.Disconnect(context.Background())
}

// Helper methods

func (n *NATSProvider) getConnectionStatus() types.ConnectionStatus {
	if n.IsConnected() {
		return types.StatusConnected
	}
	return types.StatusDisconnected
}

func (n *NATSProvider) createStream(ctx context.Context) error {
	// Check if stream already exists
	_, err := n.js.StreamInfo(n.config.StreamName)
	if err == nil {
		// Stream already exists
		return nil
	}

	// Create stream configuration
	streamConfig := &nats.StreamConfig{
		Name:       n.config.StreamName,
		Subjects:   []string{n.config.Subject},
		Retention:  n.getRetentionPolicy(),
		MaxAge:     n.config.MaxAge,
		MaxBytes:   n.config.MaxBytes,
		MaxMsgs:    n.config.MaxEvents,
		Storage:    nats.FileStorage,
		Replicas:   1,
		NoAck:      false,
		Duplicates: time.Minute,
	}

	// Create stream
	_, err = n.js.AddStream(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	n.logger.WithField("stream", n.config.StreamName).Info("Created NATS stream")
	return nil
}

func (n *NATSProvider) getRetentionPolicy() nats.RetentionPolicy {
	switch n.config.RetentionPolicy {
	case "limits":
		return nats.LimitsPolicy
	case "interest":
		return nats.InterestPolicy
	case "workqueue":
		return nats.WorkQueuePolicy
	default:
		return nats.LimitsPolicy
	}
}

func (n *NATSProvider) matchesFilters(event *types.Event, request *types.GetEventsRequest) bool {
	// Filter by stream ID
	if request.StreamID != "" && event.StreamID != request.StreamID {
		return false
	}

	// Filter by event types
	if len(request.EventTypes) > 0 {
		found := false
		for _, eventType := range request.EventTypes {
			if event.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by time range
	if request.StartTime != nil && event.Timestamp.Before(*request.StartTime) {
		return false
	}
	if request.EndTime != nil && event.Timestamp.After(*request.EndTime) {
		return false
	}

	return true
}
