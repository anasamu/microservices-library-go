package messaging

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// MessagingProvider implements messaging middleware for message queue handling
type MessagingProvider struct {
	name    string
	logger  *logrus.Logger
	config  *MessagingConfig
	enabled bool
	// Message queue handlers registry
	queues map[string]QueueHandler
	mutex  sync.RWMutex
	// Message statistics
	messageStats map[string]*MessageStats
}

// MessagingConfig holds messaging configuration
type MessagingConfig struct {
	Queues          []string               `json:"queues"`            // Supported queues: kafka, nats, rabbitmq, sqs, redis
	DefaultQueue    string                 `json:"default_queue"`     // Default queue to use
	BrokerURL       string                 `json:"broker_url"`        // Message broker URL
	TopicPrefix     string                 `json:"topic_prefix"`      // Topic prefix for messages
	GroupID         string                 `json:"group_id"`          // Consumer group ID
	AutoCommit      bool                   `json:"auto_commit"`       // Auto commit messages
	BatchSize       int                    `json:"batch_size"`        // Batch size for processing
	BatchTimeout    time.Duration          `json:"batch_timeout"`     // Batch timeout
	RetryAttempts   int                    `json:"retry_attempts"`    // Number of retry attempts
	RetryDelay      time.Duration          `json:"retry_delay"`       // Delay between retries
	MaxMessageSize  int                    `json:"max_message_size"`  // Maximum message size
	Compression     bool                   `json:"compression"`       // Enable compression
	Encryption      bool                   `json:"encryption"`        // Enable encryption
	DeadLetterQueue bool                   `json:"dead_letter_queue"` // Enable dead letter queue
	MessageTTL      time.Duration          `json:"message_ttl"`       // Message time-to-live
	Priority        int                    `json:"priority"`          // Message priority
	CustomHeaders   map[string]string      `json:"custom_headers"`    // Custom headers to add
	ExcludedTopics  []string               `json:"excluded_topics"`   // Topics to exclude from processing
	Metadata        map[string]interface{} `json:"metadata"`
}

// QueueHandler interface for different message queue implementations
type QueueHandler interface {
	GetName() string
	GetType() string
	PublishMessage(ctx context.Context, message *Message) error
	ConsumeMessage(ctx context.Context, topic string) (*Message, error)
	AcknowledgeMessage(ctx context.Context, messageID string) error
	RejectMessage(ctx context.Context, messageID string) error
	IsConnected() bool
	GetStats() map[string]interface{}
}

// Message represents a message in the queue
type Message struct {
	ID         string                 `json:"id"`
	Topic      string                 `json:"topic"`
	Key        string                 `json:"key,omitempty"`
	Value      []byte                 `json:"value"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	TTL        time.Duration          `json:"ttl,omitempty"`
	Priority   int                    `json:"priority,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MessageStats represents statistics for message processing
type MessageStats struct {
	TotalMessages     int64         `json:"total_messages"`
	ProcessedMessages int64         `json:"processed_messages"`
	FailedMessages    int64         `json:"failed_messages"`
	RetriedMessages   int64         `json:"retried_messages"`
	AverageLatency    time.Duration `json:"average_latency"`
	LastProcessed     time.Time     `json:"last_processed"`
}

// KafkaQueueHandler handles Kafka messaging
type KafkaQueueHandler struct {
	name   string
	config *KafkaConfig
	logger *logrus.Logger
	stats  *MessageStats
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	Brokers        []string          `json:"brokers"`
	Topic          string            `json:"topic"`
	Partition      int32             `json:"partition"`
	Offset         int64             `json:"offset"`
	GroupID        string            `json:"group_id"`
	AutoCommit     bool              `json:"auto_commit"`
	BatchSize      int               `json:"batch_size"`
	BatchTimeout   time.Duration     `json:"batch_timeout"`
	Compression    string            `json:"compression"`
	RetryAttempts  int               `json:"retry_attempts"`
	RetryDelay     time.Duration     `json:"retry_delay"`
	MaxMessageSize int               `json:"max_message_size"`
	Metadata       map[string]string `json:"metadata"`
}

// NATSQueueHandler handles NATS messaging
type NATSQueueHandler struct {
	name   string
	config *NATSConfig
	logger *logrus.Logger
	stats  *MessageStats
}

// NATSConfig holds NATS-specific configuration
type NATSConfig struct {
	URL            string            `json:"url"`
	Subject        string            `json:"subject"`
	QueueGroup     string            `json:"queue_group"`
	MaxReconnects  int               `json:"max_reconnects"`
	ReconnectWait  time.Duration     `json:"reconnect_wait"`
	Timeout        time.Duration     `json:"timeout"`
	MaxMessages    int               `json:"max_messages"`
	MaxBytes       int               `json:"max_bytes"`
	Compression    bool              `json:"compression"`
	RetryAttempts  int               `json:"retry_attempts"`
	RetryDelay     time.Duration     `json:"retry_delay"`
	MaxMessageSize int               `json:"max_message_size"`
	Metadata       map[string]string `json:"metadata"`
}

// RabbitMQQueueHandler handles RabbitMQ messaging
type RabbitMQQueueHandler struct {
	name   string
	config *RabbitMQConfig
	logger *logrus.Logger
	stats  *MessageStats
}

// RabbitMQConfig holds RabbitMQ-specific configuration
type RabbitMQConfig struct {
	URL            string            `json:"url"`
	Exchange       string            `json:"exchange"`
	Queue          string            `json:"queue"`
	RoutingKey     string            `json:"routing_key"`
	ExchangeType   string            `json:"exchange_type"`
	Durable        bool              `json:"durable"`
	AutoDelete     bool              `json:"auto_delete"`
	Exclusive      bool              `json:"exclusive"`
	NoWait         bool              `json:"no_wait"`
	PrefetchCount  int               `json:"prefetch_count"`
	PrefetchSize   int               `json:"prefetch_size"`
	Compression    bool              `json:"compression"`
	RetryAttempts  int               `json:"retry_attempts"`
	RetryDelay     time.Duration     `json:"retry_delay"`
	MaxMessageSize int               `json:"max_message_size"`
	Metadata       map[string]string `json:"metadata"`
}

// DefaultMessagingConfig returns default messaging configuration
func DefaultMessagingConfig() *MessagingConfig {
	return &MessagingConfig{
		Queues:          []string{"kafka", "nats", "rabbitmq"},
		DefaultQueue:    "kafka",
		BrokerURL:       "localhost:9092",
		TopicPrefix:     "microservice",
		GroupID:         "default-group",
		AutoCommit:      true,
		BatchSize:       100,
		BatchTimeout:    1 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		MaxMessageSize:  1024 * 1024, // 1MB
		Compression:     true,
		Encryption:      false,
		DeadLetterQueue: true,
		MessageTTL:      24 * time.Hour,
		Priority:        0,
		CustomHeaders:   make(map[string]string),
		ExcludedTopics:  []string{"health", "metrics", "admin"},
		Metadata:        make(map[string]interface{}),
	}
}

// NewMessagingProvider creates a new messaging provider
func NewMessagingProvider(config *MessagingConfig, logger *logrus.Logger) *MessagingProvider {
	if config == nil {
		config = DefaultMessagingConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	provider := &MessagingProvider{
		name:         "messaging",
		logger:       logger,
		config:       config,
		enabled:      true,
		queues:       make(map[string]QueueHandler),
		messageStats: make(map[string]*MessageStats),
	}

	// Initialize default queue handlers
	provider.initializeQueueHandlers()

	return provider
}

// GetName returns the provider name
func (mp *MessagingProvider) GetName() string {
	return mp.name
}

// GetType returns the provider type
func (mp *MessagingProvider) GetType() string {
	return "messaging"
}

// GetSupportedFeatures returns supported features
func (mp *MessagingProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureKafka,
		types.FeatureNATS,
		types.FeatureRabbitMQ,
		types.FeatureSQS,
		types.FeatureRedis,
		types.FeatureMessageQueue,
		types.FeaturePubSub,
		types.FeatureEventSourcing,
		types.FeatureCQRS,
		types.FeatureCompression,
		types.FeatureEncryption,
		types.FeatureRetryLogic,
		types.FeatureDeadLetterQueue,
	}
}

// GetConnectionInfo returns connection information
func (mp *MessagingProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: mp.config.DefaultQueue,
		Version:  "1.0.0",
		Secure:   mp.config.Encryption,
	}
}

// ProcessRequest processes a messaging request
func (mp *MessagingProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !mp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if topic is excluded
	if mp.isTopicExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	mp.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
		"queue":      mp.detectQueue(request),
	}).Debug("Processing messaging request")

	// Detect queue and get handler
	queue := mp.detectQueue(request)
	handler, err := mp.getQueueHandler(queue)
	if err != nil {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "unsupported queue"}`),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Process message based on HTTP method
	var response *types.MiddlewareResponse
	switch request.Method {
	case "POST", "PUT":
		response, err = mp.handlePublishMessage(ctx, handler, request)
	case "GET":
		response, err = mp.handleConsumeMessage(ctx, handler, request)
	default:
		response = &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}
	}

	if err != nil {
		mp.logger.WithError(err).Error("Queue handler error")
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "internal server error"}`),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return response, nil
}

// ProcessResponse processes a messaging response
func (mp *MessagingProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	mp.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing messaging response")

	// Add messaging headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Messaging-Provider"] = mp.name
	response.Headers["X-Queue"] = mp.config.DefaultQueue
	response.Headers["X-Messaging-Timestamp"] = time.Now().Format(time.RFC3339)

	// Add custom headers
	for name, value := range mp.config.CustomHeaders {
		response.Headers[name] = value
	}

	return response, nil
}

// CreateChain creates a messaging middleware chain
func (mp *MessagingProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	mp.logger.WithField("chain_name", config.Name).Info("Creating messaging middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return mp.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the messaging middleware chain
func (mp *MessagingProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	mp.logger.WithField("request_id", request.ID).Debug("Executing messaging middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP messaging middleware
func (mp *MessagingProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	mp.logger.WithField("config_type", config.Type).Info("Creating HTTP messaging middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create middleware request
			request := &types.MiddlewareRequest{
				ID:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
				Type:      "http",
				Method:    r.Method,
				Path:      r.URL.Path,
				Headers:   make(map[string]string),
				Query:     make(map[string]string),
				Context:   make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			// Copy headers
			for name, values := range r.Header {
				if len(values) > 0 {
					request.Headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					request.Query[name] = values[0]
				}
			}

			// Process request
			response, err := mp.ProcessRequest(r.Context(), request)
			if err != nil {
				mp.logger.WithError(err).Error("Messaging middleware error")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			if !response.Success {
				// Set response headers
				for name, value := range response.Headers {
					w.Header().Set(name, value)
				}
				w.WriteHeader(response.StatusCode)
				w.Write(response.Body)
				return
			}

			// Add messaging headers
			for name, value := range response.Headers {
				w.Header().Set(name, value)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with messaging middleware
func (mp *MessagingProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := mp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the messaging provider
func (mp *MessagingProvider) Configure(config map[string]interface{}) error {
	mp.logger.Info("Configuring messaging provider")

	if queues, ok := config["queues"].([]string); ok {
		mp.config.Queues = queues
	}

	if defaultQueue, ok := config["default_queue"].(string); ok {
		mp.config.DefaultQueue = defaultQueue
	}

	if brokerURL, ok := config["broker_url"].(string); ok {
		mp.config.BrokerURL = brokerURL
	}

	if topicPrefix, ok := config["topic_prefix"].(string); ok {
		mp.config.TopicPrefix = topicPrefix
	}

	if groupID, ok := config["group_id"].(string); ok {
		mp.config.GroupID = groupID
	}

	if autoCommit, ok := config["auto_commit"].(bool); ok {
		mp.config.AutoCommit = autoCommit
	}

	if batchSize, ok := config["batch_size"].(int); ok {
		mp.config.BatchSize = batchSize
	}

	if batchTimeout, ok := config["batch_timeout"].(time.Duration); ok {
		mp.config.BatchTimeout = batchTimeout
	}

	if retryAttempts, ok := config["retry_attempts"].(int); ok {
		mp.config.RetryAttempts = retryAttempts
	}

	if retryDelay, ok := config["retry_delay"].(time.Duration); ok {
		mp.config.RetryDelay = retryDelay
	}

	if maxMessageSize, ok := config["max_message_size"].(int); ok {
		mp.config.MaxMessageSize = maxMessageSize
	}

	if compression, ok := config["compression"].(bool); ok {
		mp.config.Compression = compression
	}

	if encryption, ok := config["encryption"].(bool); ok {
		mp.config.Encryption = encryption
	}

	if deadLetterQueue, ok := config["dead_letter_queue"].(bool); ok {
		mp.config.DeadLetterQueue = deadLetterQueue
	}

	if messageTTL, ok := config["message_ttl"].(time.Duration); ok {
		mp.config.MessageTTL = messageTTL
	}

	if excludedTopics, ok := config["excluded_topics"].([]string); ok {
		mp.config.ExcludedTopics = excludedTopics
	}

	mp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (mp *MessagingProvider) IsConfigured() bool {
	return mp.enabled && len(mp.config.Queues) > 0
}

// HealthCheck performs health check
func (mp *MessagingProvider) HealthCheck(ctx context.Context) error {
	mp.logger.Debug("Messaging provider health check")

	// Check if all queue handlers are connected
	for name, handler := range mp.queues {
		if !handler.IsConnected() {
			return fmt.Errorf("queue handler %s is not connected", name)
		}
	}

	return nil
}

// GetStats returns messaging statistics
func (mp *MessagingProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	queueStats := make(map[string]interface{})
	for name, handler := range mp.queues {
		queueStats[name] = handler.GetStats()
	}

	totalMessages := int64(0)
	processedMessages := int64(0)
	failedMessages := int64(0)

	for _, stats := range mp.messageStats {
		totalMessages += stats.TotalMessages
		processedMessages += stats.ProcessedMessages
		failedMessages += stats.FailedMessages
	}

	return &types.MiddlewareStats{
		TotalRequests:      totalMessages,
		SuccessfulRequests: processedMessages,
		FailedRequests:     failedMessages,
		AverageLatency:     5 * time.Millisecond,
		MaxLatency:         50 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          float64(failedMessages) / float64(totalMessages),
		Throughput:         float64(processedMessages),
		ActiveConnections:  int64(len(mp.queues)),
		ProviderData: map[string]interface{}{
			"provider_type":       "messaging",
			"supported_queues":    mp.config.Queues,
			"default_queue":       mp.config.DefaultQueue,
			"broker_url":          mp.config.BrokerURL,
			"topic_prefix":        mp.config.TopicPrefix,
			"group_id":            mp.config.GroupID,
			"batch_size":          mp.config.BatchSize,
			"compression_enabled": mp.config.Compression,
			"encryption_enabled":  mp.config.Encryption,
			"dead_letter_queue":   mp.config.DeadLetterQueue,
			"message_ttl":         mp.config.MessageTTL,
			"queue_stats":         queueStats,
		},
	}, nil
}

// Close closes the messaging provider
func (mp *MessagingProvider) Close() error {
	mp.logger.Info("Closing messaging provider")
	mp.enabled = false
	mp.queues = make(map[string]QueueHandler)
	mp.messageStats = make(map[string]*MessageStats)
	return nil
}

// Helper methods

func (mp *MessagingProvider) initializeQueueHandlers() {
	// Initialize Kafka handler
	kafkaHandler := &KafkaQueueHandler{
		name: "kafka",
		config: &KafkaConfig{
			Brokers:        []string{mp.config.BrokerURL},
			Topic:          mp.config.TopicPrefix + "-topic",
			Partition:      0,
			Offset:         -1,
			GroupID:        mp.config.GroupID,
			AutoCommit:     mp.config.AutoCommit,
			BatchSize:      mp.config.BatchSize,
			BatchTimeout:   mp.config.BatchTimeout,
			Compression:    "gzip",
			RetryAttempts:  mp.config.RetryAttempts,
			RetryDelay:     mp.config.RetryDelay,
			MaxMessageSize: mp.config.MaxMessageSize,
			Metadata:       make(map[string]string),
		},
		logger: mp.logger,
		stats:  &MessageStats{},
	}
	mp.queues["kafka"] = kafkaHandler

	// Initialize NATS handler
	natsHandler := &NATSQueueHandler{
		name: "nats",
		config: &NATSConfig{
			URL:            mp.config.BrokerURL,
			Subject:        mp.config.TopicPrefix + ".subject",
			QueueGroup:     mp.config.GroupID,
			MaxReconnects:  5,
			ReconnectWait:  2 * time.Second,
			Timeout:        mp.config.BatchTimeout,
			MaxMessages:    mp.config.BatchSize,
			MaxBytes:       mp.config.MaxMessageSize,
			Compression:    mp.config.Compression,
			RetryAttempts:  mp.config.RetryAttempts,
			RetryDelay:     mp.config.RetryDelay,
			MaxMessageSize: mp.config.MaxMessageSize,
			Metadata:       make(map[string]string),
		},
		logger: mp.logger,
		stats:  &MessageStats{},
	}
	mp.queues["nats"] = natsHandler

	// Initialize RabbitMQ handler
	rabbitmqHandler := &RabbitMQQueueHandler{
		name: "rabbitmq",
		config: &RabbitMQConfig{
			URL:            mp.config.BrokerURL,
			Exchange:       mp.config.TopicPrefix + "-exchange",
			Queue:          mp.config.TopicPrefix + "-queue",
			RoutingKey:     mp.config.TopicPrefix + ".routing.key",
			ExchangeType:   "topic",
			Durable:        true,
			AutoDelete:     false,
			Exclusive:      false,
			NoWait:         false,
			PrefetchCount:  mp.config.BatchSize,
			PrefetchSize:   mp.config.MaxMessageSize,
			Compression:    mp.config.Compression,
			RetryAttempts:  mp.config.RetryAttempts,
			RetryDelay:     mp.config.RetryDelay,
			MaxMessageSize: mp.config.MaxMessageSize,
			Metadata:       make(map[string]string),
		},
		logger: mp.logger,
		stats:  &MessageStats{},
	}
	mp.queues["rabbitmq"] = rabbitmqHandler
}

func (mp *MessagingProvider) detectQueue(request *types.MiddlewareRequest) string {
	// Simple queue detection based on path
	if strings.HasPrefix(request.Path, "/kafka") {
		return "kafka"
	}
	if strings.HasPrefix(request.Path, "/nats") {
		return "nats"
	}
	if strings.HasPrefix(request.Path, "/rabbitmq") {
		return "rabbitmq"
	}
	if strings.HasPrefix(request.Path, "/sqs") {
		return "sqs"
	}
	if strings.HasPrefix(request.Path, "/redis") {
		return "redis"
	}
	return mp.config.DefaultQueue
}

func (mp *MessagingProvider) getQueueHandler(queue string) (QueueHandler, error) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	handler, exists := mp.queues[queue]
	if !exists {
		return nil, fmt.Errorf("queue %s not supported", queue)
	}

	return handler, nil
}

func (mp *MessagingProvider) isTopicExcluded(path string) bool {
	for _, excludedTopic := range mp.config.ExcludedTopics {
		if strings.Contains(path, excludedTopic) {
			return true
		}
	}
	return false
}

func (mp *MessagingProvider) handlePublishMessage(ctx context.Context, handler QueueHandler, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	// Create message from request
	message := &Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Topic:     request.Path,
		Key:       request.Headers["X-Message-Key"],
		Value:     request.Body,
		Headers:   request.Headers,
		Timestamp: time.Now(),
		TTL:       mp.config.MessageTTL,
		Priority:  mp.config.Priority,
		Metadata:  request.Metadata,
	}

	// Publish message
	err := handler.PublishMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	mp.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"topic":      message.Topic,
		"queue":      handler.GetName(),
	}).Info("Message published successfully")

	return &types.MiddlewareResponse{
		ID:      request.ID,
		Success: true,
		Headers: map[string]string{
			"X-Message-ID": message.ID,
			"X-Topic":      message.Topic,
			"X-Queue":      handler.GetName(),
		},
		Context: map[string]interface{}{
			"message": message,
		},
		Timestamp: time.Now(),
	}, nil
}

func (mp *MessagingProvider) handleConsumeMessage(ctx context.Context, handler QueueHandler, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	// Consume message
	message, err := handler.ConsumeMessage(ctx, request.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to consume message: %w", err)
	}

	mp.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"topic":      message.Topic,
		"queue":      handler.GetName(),
	}).Info("Message consumed successfully")

	return &types.MiddlewareResponse{
		ID:      request.ID,
		Success: true,
		Headers: map[string]string{
			"X-Message-ID": message.ID,
			"X-Topic":      message.Topic,
			"X-Queue":      handler.GetName(),
		},
		Body: message.Value,
		Context: map[string]interface{}{
			"message": message,
		},
		Timestamp: time.Now(),
	}, nil
}

// Queue Handler Implementations

// KafkaQueueHandler implementation
func (k *KafkaQueueHandler) GetName() string   { return k.name }
func (k *KafkaQueueHandler) GetType() string   { return "kafka" }
func (k *KafkaQueueHandler) IsConnected() bool { return true } // Mock implementation

func (k *KafkaQueueHandler) PublishMessage(ctx context.Context, message *Message) error {
	k.logger.WithField("message_id", message.ID).Debug("Publishing Kafka message")
	k.stats.TotalMessages++
	k.stats.ProcessedMessages++
	k.stats.LastProcessed = time.Now()
	return nil // Mock implementation
}

func (k *KafkaQueueHandler) ConsumeMessage(ctx context.Context, topic string) (*Message, error) {
	k.logger.WithField("topic", topic).Debug("Consuming Kafka message")
	k.stats.TotalMessages++
	k.stats.ProcessedMessages++
	k.stats.LastProcessed = time.Now()

	// Mock message
	return &Message{
		ID:        fmt.Sprintf("kafka-msg-%d", time.Now().UnixNano()),
		Topic:     topic,
		Value:     []byte(`{"message": "Hello from Kafka"}`),
		Timestamp: time.Now(),
	}, nil
}

func (k *KafkaQueueHandler) AcknowledgeMessage(ctx context.Context, messageID string) error {
	k.logger.WithField("message_id", messageID).Debug("Acknowledging Kafka message")
	return nil
}

func (k *KafkaQueueHandler) RejectMessage(ctx context.Context, messageID string) error {
	k.logger.WithField("message_id", messageID).Debug("Rejecting Kafka message")
	k.stats.FailedMessages++
	return nil
}

func (k *KafkaQueueHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type":       "kafka",
		"brokers":            k.config.Brokers,
		"topic":              k.config.Topic,
		"group_id":           k.config.GroupID,
		"batch_size":         k.config.BatchSize,
		"compression":        k.config.Compression,
		"total_messages":     k.stats.TotalMessages,
		"processed_messages": k.stats.ProcessedMessages,
		"failed_messages":    k.stats.FailedMessages,
	}
}

// NATSQueueHandler implementation
func (n *NATSQueueHandler) GetName() string   { return n.name }
func (n *NATSQueueHandler) GetType() string   { return "nats" }
func (n *NATSQueueHandler) IsConnected() bool { return true } // Mock implementation

func (n *NATSQueueHandler) PublishMessage(ctx context.Context, message *Message) error {
	n.logger.WithField("message_id", message.ID).Debug("Publishing NATS message")
	n.stats.TotalMessages++
	n.stats.ProcessedMessages++
	n.stats.LastProcessed = time.Now()
	return nil // Mock implementation
}

func (n *NATSQueueHandler) ConsumeMessage(ctx context.Context, topic string) (*Message, error) {
	n.logger.WithField("topic", topic).Debug("Consuming NATS message")
	n.stats.TotalMessages++
	n.stats.ProcessedMessages++
	n.stats.LastProcessed = time.Now()

	// Mock message
	return &Message{
		ID:        fmt.Sprintf("nats-msg-%d", time.Now().UnixNano()),
		Topic:     topic,
		Value:     []byte(`{"message": "Hello from NATS"}`),
		Timestamp: time.Now(),
	}, nil
}

func (n *NATSQueueHandler) AcknowledgeMessage(ctx context.Context, messageID string) error {
	n.logger.WithField("message_id", messageID).Debug("Acknowledging NATS message")
	return nil
}

func (n *NATSQueueHandler) RejectMessage(ctx context.Context, messageID string) error {
	n.logger.WithField("message_id", messageID).Debug("Rejecting NATS message")
	n.stats.FailedMessages++
	return nil
}

func (n *NATSQueueHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type":       "nats",
		"url":                n.config.URL,
		"subject":            n.config.Subject,
		"queue_group":        n.config.QueueGroup,
		"max_reconnects":     n.config.MaxReconnects,
		"compression":        n.config.Compression,
		"total_messages":     n.stats.TotalMessages,
		"processed_messages": n.stats.ProcessedMessages,
		"failed_messages":    n.stats.FailedMessages,
	}
}

// RabbitMQQueueHandler implementation
func (r *RabbitMQQueueHandler) GetName() string   { return r.name }
func (r *RabbitMQQueueHandler) GetType() string   { return "rabbitmq" }
func (r *RabbitMQQueueHandler) IsConnected() bool { return true } // Mock implementation

func (r *RabbitMQQueueHandler) PublishMessage(ctx context.Context, message *Message) error {
	r.logger.WithField("message_id", message.ID).Debug("Publishing RabbitMQ message")
	r.stats.TotalMessages++
	r.stats.ProcessedMessages++
	r.stats.LastProcessed = time.Now()
	return nil // Mock implementation
}

func (r *RabbitMQQueueHandler) ConsumeMessage(ctx context.Context, topic string) (*Message, error) {
	r.logger.WithField("topic", topic).Debug("Consuming RabbitMQ message")
	r.stats.TotalMessages++
	r.stats.ProcessedMessages++
	r.stats.LastProcessed = time.Now()

	// Mock message
	return &Message{
		ID:        fmt.Sprintf("rabbitmq-msg-%d", time.Now().UnixNano()),
		Topic:     topic,
		Value:     []byte(`{"message": "Hello from RabbitMQ"}`),
		Timestamp: time.Now(),
	}, nil
}

func (r *RabbitMQQueueHandler) AcknowledgeMessage(ctx context.Context, messageID string) error {
	r.logger.WithField("message_id", messageID).Debug("Acknowledging RabbitMQ message")
	return nil
}

func (r *RabbitMQQueueHandler) RejectMessage(ctx context.Context, messageID string) error {
	r.logger.WithField("message_id", messageID).Debug("Rejecting RabbitMQ message")
	r.stats.FailedMessages++
	return nil
}

func (r *RabbitMQQueueHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type":       "rabbitmq",
		"url":                r.config.URL,
		"exchange":           r.config.Exchange,
		"queue":              r.config.Queue,
		"routing_key":        r.config.RoutingKey,
		"exchange_type":      r.config.ExchangeType,
		"durable":            r.config.Durable,
		"compression":        r.config.Compression,
		"total_messages":     r.stats.TotalMessages,
		"processed_messages": r.stats.ProcessedMessages,
		"failed_messages":    r.stats.FailedMessages,
	}
}
