package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// KafkaProvider implements EventSourcingProvider for Apache Kafka
type KafkaProvider struct {
	writer    *kafka.Writer
	reader    *kafka.Reader
	config    *KafkaConfig
	logger    *logrus.Logger
	connected bool
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	Brokers          []string          `json:"brokers"`
	Topic            string            `json:"topic"`
	GroupID          string            `json:"group_id"`
	Username         string            `json:"username"`
	Password         string            `json:"password"`
	SecurityProtocol string            `json:"security_protocol"`
	CompressionType  string            `json:"compression_type"`
	BatchSize        int               `json:"batch_size"`
	BatchTimeout     time.Duration     `json:"batch_timeout"`
	ReadTimeout      time.Duration     `json:"read_timeout"`
	WriteTimeout     time.Duration     `json:"write_timeout"`
	Metadata         map[string]string `json:"metadata"`
}

// NewKafkaProvider creates a new Kafka event sourcing provider
func NewKafkaProvider(config *KafkaConfig, logger *logrus.Logger) (*KafkaProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("brokers cannot be empty")
	}

	if config.Topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	if logger == nil {
		logger = logrus.New()
	}

	provider := &KafkaProvider{
		config: config,
		logger: logger,
	}

	return provider, nil
}

// GetName returns the provider name
func (k *KafkaProvider) GetName() string {
	return "kafka"
}

// GetSupportedFeatures returns the features supported by this provider
func (k *KafkaProvider) GetSupportedFeatures() []types.EventSourcingFeature {
	return []types.EventSourcingFeature{
		types.FeatureEventAppend,
		types.FeatureEventRetrieval,
		types.FeatureEventQuery,
		types.FeatureEventFiltering,
		types.FeatureEventReplay,
		types.FeatureEventVersioning,
		types.FeatureEventBatching,
		types.FeatureEventPartitioning,
		types.FeatureEventOrdering,
		types.FeatureEventCorrelation,
		types.FeatureEventStreaming,
		types.FeatureEventReplication,
	}
}

// GetConnectionInfo returns connection information
func (k *KafkaProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     k.config.Brokers[0], // Use first broker as host
		Port:     9092,                // Default Kafka port
		Database: k.config.Topic,
		Username: k.config.Username,
		Status:   k.getConnectionStatus(),
		Metadata: k.config.Metadata,
	}
}

// Connect establishes connection to Kafka
func (k *KafkaProvider) Connect(ctx context.Context) error {
	// Create writer
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.config.Brokers...),
		Topic:        k.config.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    k.config.BatchSize,
		BatchTimeout: k.config.BatchTimeout,
		WriteTimeout: k.config.WriteTimeout,
		Compression:  k.getCompressionCodec(),
	}

	// Create reader
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.config.Brokers,
		Topic:       k.config.Topic,
		GroupID:     k.config.GroupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.LastOffset,
	})

	k.connected = true
	k.logger.Info("Connected to Kafka event sourcing provider")
	return nil
}

// Disconnect closes the Kafka connections
func (k *KafkaProvider) Disconnect(ctx context.Context) error {
	var lastErr error

	if k.writer != nil {
		if err := k.writer.Close(); err != nil {
			lastErr = err
		}
	}

	if k.reader != nil {
		if err := k.reader.Close(); err != nil {
			lastErr = err
		}
	}

	k.connected = false
	return lastErr
}

// Ping checks if Kafka is reachable
func (k *KafkaProvider) Ping(ctx context.Context) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to Kafka")
	}

	// Try to create a simple message to test connectivity
	conn, err := kafka.DialLeader(ctx, "tcp", k.config.Brokers[0], k.config.Topic, 0)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	return nil
}

// IsConnected returns true if connected to Kafka
func (k *KafkaProvider) IsConnected() bool {
	return k.connected && k.writer != nil && k.reader != nil
}

// AppendEvent appends an event to Kafka
func (k *KafkaProvider) AppendEvent(ctx context.Context, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kafka")
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
		Version:       0, // Kafka doesn't maintain version per stream
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

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(request.StreamID),
		Value: eventJSON,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(eventID)},
			{Key: "event_type", Value: []byte(request.EventType)},
			{Key: "stream_id", Value: []byte(request.StreamID)},
			{Key: "timestamp", Value: []byte(now.Format(time.RFC3339))},
		},
	}

	// Add optional headers
	if request.CorrelationID != "" {
		message.Headers = append(message.Headers, kafka.Header{Key: "correlation_id", Value: []byte(request.CorrelationID)})
	}
	if request.CausationID != "" {
		message.Headers = append(message.Headers, kafka.Header{Key: "causation_id", Value: []byte(request.CausationID)})
	}
	if request.UserID != "" {
		message.Headers = append(message.Headers, kafka.Header{Key: "user_id", Value: []byte(request.UserID)})
	}
	if request.TenantID != "" {
		message.Headers = append(message.Headers, kafka.Header{Key: "tenant_id", Value: []byte(request.TenantID)})
	}

	// Write message to Kafka
	err = k.writer.WriteMessages(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return &types.AppendEventResponse{
		EventID:   eventID,
		StreamID:  request.StreamID,
		Version:   0, // Kafka doesn't maintain version
		Timestamp: now,
	}, nil
}

// GetEvents retrieves events from Kafka
func (k *KafkaProvider) GetEvents(ctx context.Context, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kafka")
	}

	// Kafka doesn't support complex queries like PostgreSQL
	// This is a simplified implementation that reads recent messages
	var events []*types.Event
	messageCount := 0
	maxMessages := request.Limit
	if maxMessages <= 0 {
		maxMessages = 100 // Default limit
	}

	// Set read timeout
	readCtx, cancel := context.WithTimeout(ctx, k.config.ReadTimeout)
	defer cancel()

	for messageCount < maxMessages {
		message, err := k.reader.ReadMessage(readCtx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			return nil, fmt.Errorf("failed to read message from Kafka: %w", err)
		}

		// Parse event
		var event types.Event
		if err := json.Unmarshal(message.Value, &event); err != nil {
			k.logger.WithError(err).Warn("Failed to unmarshal event from Kafka")
			continue
		}

		// Apply filters
		if k.matchesFilters(&event, request) {
			events = append(events, &event)
			messageCount++
		}
	}

	return &types.GetEventsResponse{
		Events:     events,
		TotalCount: int64(len(events)),
		HasMore:    false, // Kafka doesn't provide total count
	}, nil
}

// GetEventByID retrieves an event by ID (not supported in Kafka)
func (k *KafkaProvider) GetEventByID(ctx context.Context, request *types.GetEventByIDRequest) (*types.Event, error) {
	return nil, fmt.Errorf("GetEventByID not supported in Kafka provider - events are identified by offset, not ID")
}

// GetEventsByStream retrieves events for a specific stream
func (k *KafkaProvider) GetEventsByStream(ctx context.Context, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	getEventsRequest := &types.GetEventsRequest{
		StreamID: request.StreamID,
		Limit:    request.Limit,
		Options:  request.Options,
	}

	return k.GetEvents(ctx, getEventsRequest)
}

// CreateStream creates a stream (not applicable in Kafka)
func (k *KafkaProvider) CreateStream(ctx context.Context, request *types.CreateStreamRequest) error {
	// In Kafka, streams are implicit - they exist when events are written
	// We could create a topic per stream, but for simplicity, we use a single topic
	k.logger.WithField("stream_id", request.StreamID).Info("Stream creation requested (Kafka uses single topic)")
	return nil
}

// DeleteStream deletes a stream (not applicable in Kafka)
func (k *KafkaProvider) DeleteStream(ctx context.Context, request *types.DeleteStreamRequest) error {
	// In Kafka, we can't delete specific streams from a topic
	k.logger.WithField("stream_id", request.StreamID).Info("Stream deletion requested (not supported in Kafka)")
	return nil
}

// StreamExists checks if a stream exists (always true in Kafka)
func (k *KafkaProvider) StreamExists(ctx context.Context, request *types.StreamExistsRequest) (bool, error) {
	// In Kafka, streams always exist when we can write to the topic
	return k.IsConnected(), nil
}

// ListStreams lists all streams (not applicable in Kafka)
func (k *KafkaProvider) ListStreams(ctx context.Context) ([]types.StreamInfo, error) {
	// Kafka doesn't maintain a list of streams
	return []types.StreamInfo{}, nil
}

// CreateSnapshot creates a snapshot (not supported in Kafka)
func (k *KafkaProvider) CreateSnapshot(ctx context.Context, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	return nil, fmt.Errorf("snapshots not supported in Kafka provider")
}

// GetSnapshot retrieves a snapshot (not supported in Kafka)
func (k *KafkaProvider) GetSnapshot(ctx context.Context, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	return nil, fmt.Errorf("snapshots not supported in Kafka provider")
}

// DeleteSnapshot deletes a snapshot (not supported in Kafka)
func (k *KafkaProvider) DeleteSnapshot(ctx context.Context, request *types.DeleteSnapshotRequest) error {
	return fmt.Errorf("snapshots not supported in Kafka provider")
}

// AppendEventsBatch appends multiple events in a batch
func (k *KafkaProvider) AppendEventsBatch(ctx context.Context, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kafka")
	}

	var messages []kafka.Message
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

		// Create Kafka message
		message := kafka.Message{
			Key:   []byte(request.StreamID),
			Value: eventJSON,
			Headers: []kafka.Header{
				{Key: "event_id", Value: []byte(eventID)},
				{Key: "event_type", Value: []byte(event.EventType)},
				{Key: "stream_id", Value: []byte(request.StreamID)},
				{Key: "timestamp", Value: []byte(now.Format(time.RFC3339))},
			},
		}

		messages = append(messages, message)
	}

	// Write messages to Kafka
	err := k.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return nil, fmt.Errorf("failed to write messages to Kafka: %w", err)
	}

	return &types.AppendEventsBatchResponse{
		AppendedCount: len(messages),
		FailedCount:   len(failedEvents),
		FailedEvents:  failedEvents,
		StartVersion:  0,
		EndVersion:    0,
	}, nil
}

// GetStreamInfo gets information about a stream (not applicable in Kafka)
func (k *KafkaProvider) GetStreamInfo(ctx context.Context, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	// Kafka doesn't maintain stream metadata
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
func (k *KafkaProvider) GetStats(ctx context.Context) (*types.EventSourcingStats, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kafka")
	}

	// Kafka doesn't provide detailed statistics like PostgreSQL
	stats := &types.EventSourcingStats{
		TotalEvents:    0,
		TotalStreams:   0,
		TotalSnapshots: 0,
		EventsByType:   make(map[string]int64),
		EventsByStream: make(map[string]int64),
		StorageSize:    0,
		Uptime:         0,
		LastUpdate:     time.Now(),
		Provider:       k.GetName(),
	}

	return stats, nil
}

// HealthCheck performs a health check
func (k *KafkaProvider) HealthCheck(ctx context.Context) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to Kafka")
	}

	return k.Ping(ctx)
}

// Configure configures the provider
func (k *KafkaProvider) Configure(config map[string]interface{}) error {
	k.logger.WithField("config", config).Info("Kafka provider configuration updated")
	return nil
}

// IsConfigured returns true if the provider is configured
func (k *KafkaProvider) IsConfigured() bool {
	return k.config != nil && len(k.config.Brokers) > 0 && k.config.Topic != ""
}

// Close closes the provider
func (k *KafkaProvider) Close() error {
	return k.Disconnect(context.Background())
}

// Helper methods

func (k *KafkaProvider) getConnectionStatus() types.ConnectionStatus {
	if k.IsConnected() {
		return types.StatusConnected
	}
	return types.StatusDisconnected
}

func (k *KafkaProvider) getCompressionCodec() kafka.Compression {
	switch k.config.CompressionType {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	default:
		return kafka.Compression(0) // None
	}
}

func (k *KafkaProvider) matchesFilters(event *types.Event, request *types.GetEventsRequest) bool {
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
