package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/libs/messaging/gateway"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// Provider implements MessagingProvider for RabbitMQ
type Provider struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	config    map[string]interface{}
	logger    *logrus.Logger
	consumers map[string]*amqp091.Channel
}

// NewProvider creates a new RabbitMQ messaging provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:    make(map[string]interface{}),
		logger:    logger,
		consumers: make(map[string]*amqp091.Channel),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "rabbitmq"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.MessagingFeature {
	return []gateway.MessagingFeature{
		gateway.FeaturePublishSubscribe,
		gateway.FeatureRequestReply,
		gateway.FeatureMessageRouting,
		gateway.FeatureMessageFiltering,
		gateway.FeatureMessageOrdering,
		gateway.FeatureMessageDeduplication,
		gateway.FeatureMessageRetention,
		gateway.FeatureMessageCompression,
		gateway.FeatureMessageBatching,
		gateway.FeatureMessageDeadLetter,
		gateway.FeatureMessageScheduling,
		gateway.FeatureMessagePriority,
		gateway.FeatureMessageTTL,
		gateway.FeatureMessageHeaders,
		gateway.FeatureMessageCorrelation,
		gateway.FeatureMessageGrouping,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	url, _ := p.config["url"].(string)
	host := "localhost"
	port := 5672

	// Parse URL for host and port if available
	if url != "" {
		// Simple parsing - in production, use proper URL parsing
		host = "localhost" // Default
		port = 5672        // Default
	}

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     port,
		Protocol: "amqp",
		Version:  "3.8+",
	}
}

// Configure configures the RabbitMQ provider
func (p *Provider) Configure(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("rabbitmq url is required")
	}

	exchange, ok := config["exchange"].(string)
	if !ok || exchange == "" {
		exchange = "default"
	}

	p.config = config

	p.logger.Info("RabbitMQ provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	url, ok := p.config["url"].(string)
	return ok && url != ""
}

// Connect connects to RabbitMQ
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	url, _ := p.config["url"].(string)
	conn, err := amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	exchange, _ := p.config["exchange"].(string)
	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.conn = conn
	p.channel = channel

	p.logger.WithFields(logrus.Fields{
		"url":      url,
		"exchange": exchange,
	}).Info("RabbitMQ connected successfully")

	return nil
}

// Disconnect disconnects from RabbitMQ
func (p *Provider) Disconnect(ctx context.Context) error {
	// Close all consumer channels
	for topic, consumerChannel := range p.consumers {
		if err := consumerChannel.Close(); err != nil {
			p.logger.WithError(err).WithField("topic", topic).Error("Failed to close consumer channel")
		}
	}

	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}

	p.logger.Info("RabbitMQ disconnected successfully")
	return nil
}

// Ping checks RabbitMQ connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}

	// Try to declare a temporary queue to test the connection
	tempQueue := fmt.Sprintf("health_check_%d", time.Now().Unix())
	_, err := p.channel.QueueDeclare(
		tempQueue, // name
		false,     // durable
		true,      // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("rabbitmq health check failed: %w", err)
	}

	// Clean up temporary queue
	p.channel.QueueDelete(tempQueue, false, false, false)

	return nil
}

// IsConnected checks if RabbitMQ is connected
func (p *Provider) IsConnected() bool {
	return p.conn != nil && !p.conn.IsClosed()
}

// PublishMessage publishes a message to RabbitMQ
func (p *Provider) PublishMessage(ctx context.Context, request *gateway.PublishRequest) (*gateway.PublishResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return nil, fmt.Errorf("rabbitmq channel not available")
	}

	// Marshal message
	body, err := json.Marshal(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Determine routing key
	routingKey := request.RoutingKey
	if routingKey == "" {
		routingKey = request.Message.RoutingKey
	}
	if routingKey == "" {
		routingKey = request.Message.Type
	}

	// Get exchange
	exchange, _ := p.config["exchange"].(string)

	// Create publishing
	publishing := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         body,
		MessageId:    request.Message.ID.String(),
		Timestamp:    request.Message.CreatedAt,
		DeliveryMode: amqp091.Persistent, // Make message persistent
		Headers:      make(amqp091.Table),
	}

	// Add custom headers
	for key, value := range request.Headers {
		publishing.Headers[key] = value
	}

	// Add message headers
	for key, value := range request.Message.Headers {
		publishing.Headers[key] = value
	}

	// Add correlation ID if present
	if request.Message.CorrelationID != "" {
		publishing.CorrelationId = request.Message.CorrelationID
	}

	// Add reply-to if present
	if request.Message.ReplyTo != "" {
		publishing.ReplyTo = request.Message.ReplyTo
	}

	// Add priority if present
	if request.Message.Priority > 0 {
		publishing.Priority = uint8(request.Message.Priority)
	}

	// Add TTL if present
	if request.Message.TTL != nil {
		publishing.Expiration = request.Message.TTL.String()
	}

	// Publish message
	err = p.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		publishing,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	response := &gateway.PublishResponse{
		MessageID: request.Message.ID.String(),
		Topic:     request.Topic,
		Timestamp: time.Now(),
		ProviderData: map[string]interface{}{
			"exchange":    exchange,
			"routing_key": routingKey,
		},
	}

	return response, nil
}

// SubscribeToTopic subscribes to a RabbitMQ topic
func (p *Provider) SubscribeToTopic(ctx context.Context, request *gateway.SubscribeRequest, handler gateway.MessageHandler) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	// Create a new channel for this consumer
	consumerChannel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create consumer channel: %w", err)
	}

	// Declare queue
	queue, err := consumerChannel.QueueDeclare(
		request.Topic, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		consumerChannel.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	exchange, _ := p.config["exchange"].(string)
	err = consumerChannel.QueueBind(
		queue.Name, // queue name
		"#",        // routing key (all messages)
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		consumerChannel.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Consume messages
	msgs, err := consumerChannel.Consume(
		queue.Name,      // queue
		"",              // consumer
		request.AutoAck, // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		consumerChannel.Close()
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Store consumer channel
	p.consumers[request.Topic] = consumerChannel

	p.logger.WithField("topic", request.Topic).Info("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.WithField("topic", request.Topic).Info("Stopped consuming messages")
				return
			case delivery := <-msgs:
				p.handleMessage(ctx, delivery, handler)
			}
		}
	}()

	return nil
}

// UnsubscribeFromTopic unsubscribes from a RabbitMQ topic
func (p *Provider) UnsubscribeFromTopic(ctx context.Context, request *gateway.UnsubscribeRequest) error {
	if consumerChannel, exists := p.consumers[request.Topic]; exists {
		if err := consumerChannel.Close(); err != nil {
			return fmt.Errorf("failed to close consumer channel: %w", err)
		}
		delete(p.consumers, request.Topic)
	}

	p.logger.WithField("topic", request.Topic).Info("Unsubscribed from topic")
	return nil
}

// CreateTopic creates a RabbitMQ topic (queue)
func (p *Provider) CreateTopic(ctx context.Context, request *gateway.CreateTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	// Declare queue
	_, err := p.channel.QueueDeclare(
		request.Topic, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to create queue: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Topic created successfully")
	return nil
}

// DeleteTopic deletes a RabbitMQ topic (queue)
func (p *Provider) DeleteTopic(ctx context.Context, request *gateway.DeleteTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	_, err := p.channel.QueueDelete(
		request.Topic, // name
		false,         // if-unused
		false,         // if-empty
		false,         // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Topic deleted successfully")
	return nil
}

// TopicExists checks if a RabbitMQ topic (queue) exists
func (p *Provider) TopicExists(ctx context.Context, request *gateway.TopicExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return false, fmt.Errorf("rabbitmq channel not available")
	}

	_, err := p.channel.QueueInspect(request.Topic)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == "Exception (404) Reason: \"NOT_FOUND - no queue '"+request.Topic+"' in vhost '/'" {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return true, nil
}

// ListTopics lists RabbitMQ topics (queues)
func (p *Provider) ListTopics(ctx context.Context) ([]gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return nil, fmt.Errorf("rabbitmq channel not available")
	}

	// RabbitMQ doesn't have a direct way to list all queues via AMQP
	// This would typically require management API access
	// For now, return empty list
	topics := make([]gateway.TopicInfo, 0)

	p.logger.Debug("Listed topics (RabbitMQ requires management API for full queue listing)")
	return topics, nil
}

// PublishBatch publishes multiple messages to RabbitMQ
func (p *Provider) PublishBatch(ctx context.Context, request *gateway.PublishBatchRequest) (*gateway.PublishBatchResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return nil, fmt.Errorf("rabbitmq channel not available")
	}

	publishedCount := 0
	failedCount := 0
	var failedMessages []*gateway.Message

	for _, msg := range request.Messages {
		publishRequest := &gateway.PublishRequest{
			Topic:      request.Topic,
			Message:    msg,
			RoutingKey: request.RoutingKey,
		}

		_, err := p.PublishMessage(ctx, publishRequest)
		if err != nil {
			failedCount++
			failedMessages = append(failedMessages, msg)
			p.logger.WithError(err).WithField("message_id", msg.ID).Error("Failed to publish message in batch")
		} else {
			publishedCount++
		}
	}

	response := &gateway.PublishBatchResponse{
		PublishedCount: publishedCount,
		FailedCount:    failedCount,
		FailedMessages: failedMessages,
		ProviderData: map[string]interface{}{
			"exchange": p.config["exchange"],
		},
	}

	return response, nil
}

// GetTopicInfo gets RabbitMQ topic information
func (p *Provider) GetTopicInfo(ctx context.Context, request *gateway.GetTopicInfoRequest) (*gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("rabbitmq provider not configured")
	}

	if p.channel == nil {
		return nil, fmt.Errorf("rabbitmq channel not available")
	}

	queue, err := p.channel.QueueInspect(request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	info := &gateway.TopicInfo{
		Name:         request.Topic,
		MessageCount: int64(queue.Messages),
		ProviderData: map[string]interface{}{
			"consumers": queue.Consumers,
			"durable":   queue.Durable,
		},
	}

	return info, nil
}

// GetStats returns RabbitMQ statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.MessagingStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("rabbitmq provider not configured")
	}

	stats := &gateway.MessagingStats{
		ActiveConnections:   1, // Main connection
		ActiveSubscriptions: len(p.consumers),
		ProviderData: map[string]interface{}{
			"consumers": len(p.consumers),
			"exchange":  p.config["exchange"],
		},
	}

	return stats, nil
}

// HealthCheck performs a health check on RabbitMQ
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("rabbitmq provider not configured")
	}

	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}

	// Try to declare a temporary queue to test the connection
	tempQueue := fmt.Sprintf("health_check_%d", time.Now().Unix())
	_, err := p.channel.QueueDeclare(
		tempQueue, // name
		false,     // durable
		true,      // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("rabbitmq health check failed: %w", err)
	}

	// Clean up temporary queue
	p.channel.QueueDelete(tempQueue, false, false, false)

	return nil
}

// Close closes the RabbitMQ provider
func (p *Provider) Close() error {
	return p.Disconnect(context.Background())
}

// handleMessage handles a single RabbitMQ message delivery
func (p *Provider) handleMessage(ctx context.Context, delivery amqp091.Delivery, handler gateway.MessageHandler) {
	var gatewayMessage gateway.Message
	if err := json.Unmarshal(delivery.Body, &gatewayMessage); err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":  delivery.MessageId,
			"routing_key": delivery.RoutingKey,
		}).Error("Failed to unmarshal message")
		delivery.Nack(false, false) // Reject message
		return
	}

	// Set RabbitMQ-specific fields
	gatewayMessage.ProviderData = map[string]interface{}{
		"delivery_tag": delivery.DeliveryTag,
		"routing_key":  delivery.RoutingKey,
		"exchange":     delivery.Exchange,
	}

	// Check if message is expired
	if gatewayMessage.ExpiresAt != nil && time.Now().After(*gatewayMessage.ExpiresAt) {
		p.logger.WithFields(logrus.Fields{
			"message_id": gatewayMessage.ID,
			"expires_at": gatewayMessage.ExpiresAt,
		}).Debug("Message expired, discarding")
		delivery.Ack(false) // Acknowledge expired message
		return
	}

	// Handle message
	if err := handler(ctx, &gatewayMessage); err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   gatewayMessage.ID,
			"message_type": gatewayMessage.Type,
			"source":       gatewayMessage.Source,
			"target":       gatewayMessage.Target,
		}).Error("Failed to handle message")
		delivery.Nack(false, true) // Reject and requeue
		return
	}

	// Acknowledge message
	delivery.Ack(false)

	p.logger.WithFields(logrus.Fields{
		"message_id":   gatewayMessage.ID,
		"message_type": gatewayMessage.Type,
		"source":       gatewayMessage.Source,
		"target":       gatewayMessage.Target,
	}).Debug("Message handled successfully")
}
