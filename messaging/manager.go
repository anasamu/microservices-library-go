package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MessagingManager manages multiple messaging providers
type MessagingManager struct {
	providers map[string]MessagingProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds messaging manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	MaxMessageSize  int64             `json:"max_message_size"`
	Metadata        map[string]string `json:"metadata"`
}

// MessagingProvider interface for messaging backends
type MessagingProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []MessagingFeature
	GetConnectionInfo() *ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Message operations
	PublishMessage(ctx context.Context, request *PublishRequest) (*PublishResponse, error)
	SubscribeToTopic(ctx context.Context, request *SubscribeRequest, handler MessageHandler) error
	UnsubscribeFromTopic(ctx context.Context, request *UnsubscribeRequest) error

	// Topic/Queue management
	CreateTopic(ctx context.Context, request *CreateTopicRequest) error
	DeleteTopic(ctx context.Context, request *DeleteTopicRequest) error
	TopicExists(ctx context.Context, request *TopicExistsRequest) (bool, error)
	ListTopics(ctx context.Context) ([]TopicInfo, error)

	// Advanced operations
	PublishBatch(ctx context.Context, request *PublishBatchRequest) (*PublishBatchResponse, error)
	GetTopicInfo(ctx context.Context, request *GetTopicInfoRequest) (*TopicInfo, error)
	GetStats(ctx context.Context) (*MessagingStats, error)

	// Health and monitoring
	HealthCheck(ctx context.Context) error

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// MessagingFeature represents a messaging feature
type MessagingFeature string

const (
	FeaturePublishSubscribe     MessagingFeature = "pub_sub"
	FeatureRequestReply         MessagingFeature = "request_reply"
	FeatureMessageRouting       MessagingFeature = "message_routing"
	FeatureMessageFiltering     MessagingFeature = "message_filtering"
	FeatureMessageOrdering      MessagingFeature = "message_ordering"
	FeatureMessageDeduplication MessagingFeature = "message_deduplication"
	FeatureMessageRetention     MessagingFeature = "message_retention"
	FeatureMessageCompression   MessagingFeature = "message_compression"
	FeatureMessageEncryption    MessagingFeature = "message_encryption"
	FeatureMessageBatching      MessagingFeature = "message_batching"
	FeatureMessagePartitioning  MessagingFeature = "message_partitioning"
	FeatureMessageReplay        MessagingFeature = "message_replay"
	FeatureMessageDeadLetter    MessagingFeature = "message_dead_letter"
	FeatureMessageScheduling    MessagingFeature = "message_scheduling"
	FeatureMessagePriority      MessagingFeature = "message_priority"
	FeatureMessageTTL           MessagingFeature = "message_ttl"
	FeatureMessageHeaders       MessagingFeature = "message_headers"
	FeatureMessageCorrelation   MessagingFeature = "message_correlation"
	FeatureMessageGrouping      MessagingFeature = "message_grouping"
	FeatureMessageStreaming     MessagingFeature = "message_streaming"
)

// ConnectionInfo represents messaging connection information
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
}

// Message represents a unified message structure
type Message struct {
	ID            uuid.UUID              `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target"`
	Topic         string                 `json:"topic"`
	RoutingKey    string                 `json:"routing_key,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	ScheduledAt   *time.Time             `json:"scheduled_at,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
	TTL           *time.Duration         `json:"ttl,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ReplyTo       string                 `json:"reply_to,omitempty"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, message *Message) error

// PublishRequest represents a publish message request
type PublishRequest struct {
	Topic      string                 `json:"topic"`
	Message    *Message               `json:"message"`
	RoutingKey string                 `json:"routing_key,omitempty"`
	Headers    map[string]interface{} `json:"headers,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// PublishResponse represents a publish message response
type PublishResponse struct {
	MessageID    string                 `json:"message_id"`
	Topic        string                 `json:"topic"`
	Partition    int                    `json:"partition,omitempty"`
	Offset       int64                  `json:"offset,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// SubscribeRequest represents a subscribe request
type SubscribeRequest struct {
	Topic         string                 `json:"topic"`
	GroupID       string                 `json:"group_id,omitempty"`
	ConsumerID    string                 `json:"consumer_id,omitempty"`
	AutoAck       bool                   `json:"auto_ack"`
	PrefetchCount int                    `json:"prefetch_count,omitempty"`
	StartOffset   string                 `json:"start_offset,omitempty"`
	Filter        map[string]interface{} `json:"filter,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

// UnsubscribeRequest represents an unsubscribe request
type UnsubscribeRequest struct {
	Topic      string `json:"topic"`
	GroupID    string `json:"group_id,omitempty"`
	ConsumerID string `json:"consumer_id,omitempty"`
}

// CreateTopicRequest represents a create topic request
type CreateTopicRequest struct {
	Topic             string                 `json:"topic"`
	Partitions        int                    `json:"partitions,omitempty"`
	ReplicationFactor int                    `json:"replication_factor,omitempty"`
	RetentionPeriod   *time.Duration         `json:"retention_period,omitempty"`
	Config            map[string]interface{} `json:"config,omitempty"`
}

// DeleteTopicRequest represents a delete topic request
type DeleteTopicRequest struct {
	Topic string `json:"topic"`
}

// TopicExistsRequest represents a topic exists request
type TopicExistsRequest struct {
	Topic string `json:"topic"`
}

// TopicInfo represents topic information
type TopicInfo struct {
	Name              string                 `json:"name"`
	Partitions        int                    `json:"partitions,omitempty"`
	ReplicationFactor int                    `json:"replication_factor,omitempty"`
	RetentionPeriod   *time.Duration         `json:"retention_period,omitempty"`
	MessageCount      int64                  `json:"message_count,omitempty"`
	Size              int64                  `json:"size,omitempty"`
	CreatedAt         time.Time              `json:"created_at,omitempty"`
	ProviderData      map[string]interface{} `json:"provider_data,omitempty"`
}

// PublishBatchRequest represents a batch publish request
type PublishBatchRequest struct {
	Topic      string                 `json:"topic"`
	Messages   []*Message             `json:"messages"`
	RoutingKey string                 `json:"routing_key,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// PublishBatchResponse represents a batch publish response
type PublishBatchResponse struct {
	PublishedCount int                    `json:"published_count"`
	FailedCount    int                    `json:"failed_count"`
	FailedMessages []*Message             `json:"failed_messages,omitempty"`
	ProviderData   map[string]interface{} `json:"provider_data,omitempty"`
}

// GetTopicInfoRequest represents a get topic info request
type GetTopicInfoRequest struct {
	Topic string `json:"topic"`
}

// MessagingStats represents messaging statistics
type MessagingStats struct {
	PublishedMessages   int64                  `json:"published_messages"`
	ConsumedMessages    int64                  `json:"consumed_messages"`
	FailedMessages      int64                  `json:"failed_messages"`
	ActiveConnections   int                    `json:"active_connections"`
	ActiveSubscriptions int                    `json:"active_subscriptions"`
	ProviderData        map[string]interface{} `json:"provider_data"`
}

// DefaultManagerConfig returns default messaging manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "kafka",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		MaxMessageSize:  1024 * 1024, // 1MB
		Metadata:        make(map[string]string),
	}
}

// NewMessagingManager creates a new messaging manager
func NewMessagingManager(config *ManagerConfig, logger *logrus.Logger) *MessagingManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &MessagingManager{
		providers: make(map[string]MessagingProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a messaging provider
func (mm *MessagingManager) RegisterProvider(provider MessagingProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	mm.providers[name] = provider
	mm.logger.WithField("provider", name).Info("Messaging provider registered")

	return nil
}

// GetProvider returns a messaging provider by name
func (mm *MessagingManager) GetProvider(name string) (MessagingProvider, error) {
	provider, exists := mm.providers[name]
	if !exists {
		return nil, fmt.Errorf("messaging provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default messaging provider
func (mm *MessagingManager) GetDefaultProvider() (MessagingProvider, error) {
	return mm.GetProvider(mm.config.DefaultProvider)
}

// Connect connects to a messaging system using the specified provider
func (mm *MessagingManager) Connect(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Connect with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.Connect(ctx)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
		}).Warn("Messaging connection failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to messaging system after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithField("provider", providerName).Info("Messaging system connected successfully")
	return nil
}

// Disconnect disconnects from a messaging system using the specified provider
func (mm *MessagingManager) Disconnect(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from messaging system: %w", err)
	}

	mm.logger.WithField("provider", providerName).Info("Messaging system disconnected successfully")
	return nil
}

// Ping pings a messaging system using the specified provider
func (mm *MessagingManager) Ping(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping messaging system: %w", err)
	}

	return nil
}

// PublishMessage publishes a message using the specified provider
func (mm *MessagingManager) PublishMessage(ctx context.Context, providerName string, request *PublishRequest) (*PublishResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := mm.validatePublishRequest(request); err != nil {
		return nil, fmt.Errorf("invalid publish request: %w", err)
	}

	// Check message size limit
	if request.Message != nil && mm.getMessageSize(request.Message) > mm.config.MaxMessageSize {
		return nil, fmt.Errorf("message size %d exceeds maximum allowed size %d", mm.getMessageSize(request.Message), mm.config.MaxMessageSize)
	}

	// Set default values
	if request.Message.ID == uuid.Nil {
		request.Message.ID = uuid.New()
	}
	if request.Message.CreatedAt.IsZero() {
		request.Message.CreatedAt = time.Now()
	}

	// Publish with retry logic
	var response *PublishResponse
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		response, err = provider.PublishMessage(ctx, request)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider":   providerName,
			"attempt":    attempt,
			"topic":      request.Topic,
			"message_id": request.Message.ID,
		}).Warn("Message publish failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to publish message after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"topic":      request.Topic,
		"message_id": request.Message.ID,
		"type":       request.Message.Type,
	}).Info("Message published successfully")

	return response, nil
}

// SubscribeToTopic subscribes to a topic using the specified provider
func (mm *MessagingManager) SubscribeToTopic(ctx context.Context, providerName string, request *SubscribeRequest, handler MessageHandler) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := mm.validateSubscribeRequest(request); err != nil {
		return fmt.Errorf("invalid subscribe request: %w", err)
	}

	err = provider.SubscribeToTopic(ctx, request, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"topic":    request.Topic,
		"group_id": request.GroupID,
	}).Info("Subscribed to topic successfully")

	return nil
}

// UnsubscribeFromTopic unsubscribes from a topic using the specified provider
func (mm *MessagingManager) UnsubscribeFromTopic(ctx context.Context, providerName string, request *UnsubscribeRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.UnsubscribeFromTopic(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"topic":    request.Topic,
		"group_id": request.GroupID,
	}).Info("Unsubscribed from topic successfully")

	return nil
}

// CreateTopic creates a topic using the specified provider
func (mm *MessagingManager) CreateTopic(ctx context.Context, providerName string, request *CreateTopicRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.CreateTopic(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"topic":      request.Topic,
		"partitions": request.Partitions,
	}).Info("Topic created successfully")

	return nil
}

// DeleteTopic deletes a topic using the specified provider
func (mm *MessagingManager) DeleteTopic(ctx context.Context, providerName string, request *DeleteTopicRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.DeleteTopic(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"topic":    request.Topic,
	}).Info("Topic deleted successfully")

	return nil
}

// TopicExists checks if a topic exists using the specified provider
func (mm *MessagingManager) TopicExists(ctx context.Context, providerName string, request *TopicExistsRequest) (bool, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return false, err
	}

	exists, err := provider.TopicExists(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to check topic existence: %w", err)
	}

	return exists, nil
}

// ListTopics lists topics using the specified provider
func (mm *MessagingManager) ListTopics(ctx context.Context, providerName string) ([]TopicInfo, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	topics, err := provider.ListTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(topics),
	}).Debug("Topics listed successfully")

	return topics, nil
}

// PublishBatch publishes multiple messages using the specified provider
func (mm *MessagingManager) PublishBatch(ctx context.Context, providerName string, request *PublishBatchRequest) (*PublishBatchResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := mm.validatePublishBatchRequest(request); err != nil {
		return nil, fmt.Errorf("invalid publish batch request: %w", err)
	}

	response, err := provider.PublishBatch(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to publish batch: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":        providerName,
		"topic":           request.Topic,
		"published_count": response.PublishedCount,
		"failed_count":    response.FailedCount,
	}).Info("Batch published successfully")

	return response, nil
}

// GetTopicInfo gets topic information using the specified provider
func (mm *MessagingManager) GetTopicInfo(ctx context.Context, providerName string, request *GetTopicInfoRequest) (*TopicInfo, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	info, err := provider.GetTopicInfo(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic info: %w", err)
	}

	return info, nil
}

// HealthCheck performs health check on all providers
func (mm *MessagingManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range mm.providers {
		err := provider.HealthCheck(ctx)
		results[name] = err
	}

	return results
}

// GetStats gets statistics from a provider
func (mm *MessagingManager) GetStats(ctx context.Context, providerName string) (*MessagingStats, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	stats, err := provider.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging stats: %w", err)
	}

	return stats, nil
}

// GetSupportedProviders returns a list of registered providers
func (mm *MessagingManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(mm.providers))
	for name := range mm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (mm *MessagingManager) GetProviderCapabilities(providerName string) ([]MessagingFeature, *ConnectionInfo, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	return provider.GetSupportedFeatures(), provider.GetConnectionInfo(), nil
}

// Close closes all messaging connections
func (mm *MessagingManager) Close() error {
	var lastErr error

	for name, provider := range mm.providers {
		if err := provider.Close(); err != nil {
			mm.logger.WithError(err).WithField("provider", name).Error("Failed to close messaging provider")
			lastErr = err
		}
	}

	return lastErr
}

// IsProviderConnected checks if a provider is connected
func (mm *MessagingManager) IsProviderConnected(providerName string) bool {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return false
	}
	return provider.IsConnected()
}

// GetConnectedProviders returns a list of connected providers
func (mm *MessagingManager) GetConnectedProviders() []string {
	connected := make([]string, 0)
	for name, provider := range mm.providers {
		if provider.IsConnected() {
			connected = append(connected, name)
		}
	}
	return connected
}

// validatePublishRequest validates a publish request
func (mm *MessagingManager) validatePublishRequest(request *PublishRequest) error {
	if request.Topic == "" {
		return fmt.Errorf("topic is required")
	}

	if request.Message == nil {
		return fmt.Errorf("message is required")
	}

	if request.Message.Type == "" {
		return fmt.Errorf("message type is required")
	}

	return nil
}

// validateSubscribeRequest validates a subscribe request
func (mm *MessagingManager) validateSubscribeRequest(request *SubscribeRequest) error {
	if request.Topic == "" {
		return fmt.Errorf("topic is required")
	}

	return nil
}

// validatePublishBatchRequest validates a publish batch request
func (mm *MessagingManager) validatePublishBatchRequest(request *PublishBatchRequest) error {
	if request.Topic == "" {
		return fmt.Errorf("topic is required")
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}

	return nil
}

// getMessageSize calculates the approximate size of a message
func (mm *MessagingManager) getMessageSize(message *Message) int64 {
	// This is a simplified calculation
	// In a real implementation, you might want to marshal the message to JSON
	// and get the actual byte size
	return int64(len(message.Type) + len(message.Source) + len(message.Target) + len(message.Topic))
}

// CreateMessage creates a new message with default values
func CreateMessage(messageType, source, target, topic string, payload map[string]interface{}) *Message {
	return &Message{
		ID:        uuid.New(),
		Type:      messageType,
		Source:    source,
		Target:    target,
		Topic:     topic,
		Payload:   payload,
		Headers:   make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// SetExpiration sets message expiration
func (m *Message) SetExpiration(duration time.Duration) {
	expiresAt := time.Now().Add(duration)
	m.ExpiresAt = &expiresAt
}

// SetScheduledTime sets message scheduled time
func (m *Message) SetScheduledTime(scheduledAt time.Time) {
	m.ScheduledAt = &scheduledAt
}

// SetPriority sets message priority
func (m *Message) SetPriority(priority int) {
	m.Priority = priority
}

// SetTTL sets message TTL
func (m *Message) SetTTL(ttl time.Duration) {
	m.TTL = &ttl
}

// SetCorrelationID sets message correlation ID
func (m *Message) SetCorrelationID(correlationID string) {
	m.CorrelationID = correlationID
}

// SetReplyTo sets message reply-to address
func (m *Message) SetReplyTo(replyTo string) {
	m.ReplyTo = replyTo
}

// AddHeader adds a header to the message
func (m *Message) AddHeader(key string, value interface{}) {
	if m.Headers == nil {
		m.Headers = make(map[string]interface{})
	}
	m.Headers[key] = value
}

// GetHeader retrieves a header from the message
func (m *Message) GetHeader(key string) (interface{}, bool) {
	if m.Headers == nil {
		return nil, false
	}
	value, exists := m.Headers[key]
	return value, exists
}

// AddMetadata adds metadata to the message
func (m *Message) AddMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata retrieves metadata from the message
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}
