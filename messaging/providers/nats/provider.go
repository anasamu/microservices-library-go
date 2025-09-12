package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/messaging"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// Provider implements MessagingProvider for NATS
type Provider struct {
	conn      *nats.Conn
	js        nats.JetStreamContext
	config    map[string]interface{}
	logger    *logrus.Logger
	subs      map[string]*nats.Subscription
	consumers map[string]*nats.ConsumerInfo
}

// NewProvider creates a new NATS messaging provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:    make(map[string]interface{}),
		logger:    logger,
		subs:      make(map[string]*nats.Subscription),
		consumers: make(map[string]*nats.ConsumerInfo),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "nats"
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
		gateway.FeatureMessageReplay,
		gateway.FeatureMessageDeadLetter,
		gateway.FeatureMessageScheduling,
		gateway.FeatureMessagePriority,
		gateway.FeatureMessageTTL,
		gateway.FeatureMessageHeaders,
		gateway.FeatureMessageCorrelation,
		gateway.FeatureMessageGrouping,
		gateway.FeatureMessageStreaming,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	if p.conn == nil {
		return &gateway.ConnectionInfo{
			Host:     "unknown",
			Port:     0,
			Protocol: "nats",
			Version:  "unknown",
		}
	}

	return &gateway.ConnectionInfo{
		Host:     p.conn.ConnectedAddr(),
		Port:     4222, // Default NATS port
		Protocol: "nats",
		Version:  p.conn.ConnectedServerVersion(),
	}
}

// Configure configures the NATS provider
func (p *Provider) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Connect connects to NATS server
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("provider not configured")
	}

	// Get connection URL
	url, ok := p.config["url"].(string)
	if !ok {
		url = nats.DefaultURL
	}

	// Get connection options
	opts := []nats.Option{
		nats.Name("microservices-library-go"),
		nats.Timeout(30 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			p.logger.WithError(err).Warn("NATS disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			p.logger.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			p.logger.Info("NATS connection closed")
		}),
	}

	// Add authentication if provided
	if username, ok := p.config["username"].(string); ok {
		if password, ok := p.config["password"].(string); ok {
			opts = append(opts, nats.UserInfo(username, password))
		}
	}

	// Add token if provided
	if token, ok := p.config["token"].(string); ok {
		opts = append(opts, nats.Token(token))
	}

	// Add TLS if configured
	if tlsConfig, ok := p.config["tls"].(bool); ok && tlsConfig {
		opts = append(opts, nats.Secure())
	}

	// Connect to NATS
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	p.conn = conn

	// Initialize JetStream if enabled
	if jetstream, ok := p.config["jetstream"].(bool); ok && jetstream {
		js, err := conn.JetStream()
		if err != nil {
			return fmt.Errorf("failed to initialize JetStream: %w", err)
		}
		p.js = js
	}

	p.logger.Info("Connected to NATS successfully")
	return nil
}

// Disconnect disconnects from NATS server
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.conn == nil {
		return nil
	}

	// Unsubscribe from all subscriptions
	for topic, sub := range p.subs {
		if err := sub.Unsubscribe(); err != nil {
			p.logger.WithError(err).WithField("topic", topic).Warn("Failed to unsubscribe")
		}
	}
	p.subs = make(map[string]*nats.Subscription)

	// Close connection
	p.conn.Close()
	p.conn = nil
	p.js = nil

	p.logger.Info("Disconnected from NATS successfully")
	return nil
}

// Ping pings the NATS server
func (p *Provider) Ping(ctx context.Context) error {
	if p.conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	if !p.conn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	return nil
}

// IsConnected checks if connected to NATS
func (p *Provider) IsConnected() bool {
	return p.conn != nil && p.conn.IsConnected()
}

// PublishMessage publishes a message to a subject
func (p *Provider) PublishMessage(ctx context.Context, request *gateway.PublishRequest) (*gateway.PublishResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	// Marshal message to JSON
	data, err := json.Marshal(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create NATS message
	msg := &nats.Msg{
		Subject: request.Topic,
		Data:    data,
		Header:  make(nats.Header),
	}

	// Add headers
	for key, value := range request.Headers {
		msg.Header.Set(key, fmt.Sprintf("%v", value))
	}

	// Add message headers
	for key, value := range request.Message.Headers {
		msg.Header.Set(key, fmt.Sprintf("%v", value))
	}

	// Add message metadata as headers
	for key, value := range request.Message.Metadata {
		msg.Header.Set("X-Metadata-"+key, fmt.Sprintf("%v", value))
	}

	// Set message ID
	msg.Header.Set("X-Message-ID", request.Message.ID.String())
	msg.Header.Set("X-Message-Type", request.Message.Type)
	msg.Header.Set("X-Source", request.Message.Source)
	msg.Header.Set("X-Target", request.Message.Target)

	// Set correlation ID if provided
	if request.Message.CorrelationID != "" {
		msg.Header.Set("X-Correlation-ID", request.Message.CorrelationID)
	}

	// Set reply-to if provided
	if request.Message.ReplyTo != "" {
		msg.Reply = request.Message.ReplyTo
	}

	// Publish message
	var err2 error
	if p.js != nil {
		// Use JetStream for persistent messaging
		_, err2 = p.js.PublishMsg(msg)
	} else {
		// Use regular NATS messaging
		err2 = p.conn.PublishMsg(msg)
	}

	if err2 != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err2)
	}

	response := &gateway.PublishResponse{
		MessageID: request.Message.ID.String(),
		Topic:     request.Topic,
		Timestamp: time.Now(),
	}

	p.logger.WithFields(logrus.Fields{
		"topic":      request.Topic,
		"message_id": request.Message.ID,
		"type":       request.Message.Type,
	}).Debug("Message published successfully")

	return response, nil
}

// SubscribeToTopic subscribes to a subject
func (p *Provider) SubscribeToTopic(ctx context.Context, request *gateway.SubscribeRequest, handler gateway.MessageHandler) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Create message handler
	natsHandler := func(msg *nats.Msg) {
		// Parse message
		var message gateway.Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			p.logger.WithError(err).Error("Failed to unmarshal message")
			return
		}

		// Extract headers
		message.Headers = make(map[string]interface{})
		for key, values := range msg.Header {
			if len(values) > 0 {
				message.Headers[key] = values[0]
			}
		}

		// Extract metadata from headers
		message.Metadata = make(map[string]interface{})
		for key, values := range msg.Header {
			if len(key) > 11 && key[:11] == "X-Metadata-" {
				metadataKey := key[11:]
				if len(values) > 0 {
					message.Metadata[metadataKey] = values[0]
				}
			}
		}

		// Set reply-to if present
		if msg.Reply != "" {
			message.ReplyTo = msg.Reply
		}

		// Call user handler
		if err := handler(ctx, &message); err != nil {
			p.logger.WithError(err).WithField("message_id", message.ID).Error("Message handler failed")
		}
	}

	// Subscribe to subject
	var sub *nats.Subscription
	var err error

	if p.js != nil {
		// Use JetStream for persistent messaging
		sub, err = p.js.Subscribe(request.Topic, natsHandler, nats.Durable(request.GroupID))
	} else {
		// Use regular NATS messaging
		sub, err = p.conn.Subscribe(request.Topic, natsHandler)
	}

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Store subscription
	p.subs[request.Topic] = sub

	p.logger.WithFields(logrus.Fields{
		"topic":   request.Topic,
		"group_id": request.GroupID,
	}).Info("Subscribed to topic successfully")

	return nil
}

// UnsubscribeFromTopic unsubscribes from a subject
func (p *Provider) UnsubscribeFromTopic(ctx context.Context, request *gateway.UnsubscribeRequest) error {
	sub, exists := p.subs[request.Topic]
	if !exists {
		return fmt.Errorf("subscription not found for topic: %s", request.Topic)
	}

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("failed to unsubscribe from topic: %w", err)
	}

	delete(p.subs, request.Topic)

	p.logger.WithField("topic", request.Topic).Info("Unsubscribed from topic successfully")
	return nil
}

// CreateTopic creates a stream (JetStream only)
func (p *Provider) CreateTopic(ctx context.Context, request *gateway.CreateTopicRequest) error {
	if p.js == nil {
		return fmt.Errorf("JetStream not enabled, cannot create topics")
	}

	// Create stream configuration
	streamConfig := &nats.StreamConfig{
		Name:     request.Topic,
		Subjects: []string{request.Topic + ".*"},
		Storage:  nats.FileStorage,
	}

	// Set retention if provided
	if request.RetentionPeriod != nil {
		streamConfig.MaxAge = *request.RetentionPeriod
	}

	// Set replication factor if provided
	if request.ReplicationFactor > 0 {
		streamConfig.Replicas = request.ReplicationFactor
	}

	// Create stream
	_, err := p.js.AddStream(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Stream created successfully")
	return nil
}

// DeleteTopic deletes a stream (JetStream only)
func (p *Provider) DeleteTopic(ctx context.Context, request *gateway.DeleteTopicRequest) error {
	if p.js == nil {
		return fmt.Errorf("JetStream not enabled, cannot delete topics")
	}

	if err := p.js.DeleteStream(request.Topic); err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Stream deleted successfully")
	return nil
}

// TopicExists checks if a stream exists (JetStream only)
func (p *Provider) TopicExists(ctx context.Context, request *gateway.TopicExistsRequest) (bool, error) {
	if p.js == nil {
		// For regular NATS, subjects always exist
		return true, nil
	}

	stream, err := p.js.StreamInfo(request.Topic)
	if err != nil {
		return false, nil
	}

	return stream != nil, nil
}

// ListTopics lists all streams (JetStream only)
func (p *Provider) ListTopics(ctx context.Context) ([]gateway.TopicInfo, error) {
	if p.js == nil {
		// For regular NATS, we can't list subjects
		return []gateway.TopicInfo{}, nil
	}

	streams := p.js.Streams()
	topics := make([]gateway.TopicInfo, 0)

	for stream := range streams {
		topic := gateway.TopicInfo{
			Name:              stream.Config.Name,
			MessageCount:      int64(stream.State.Msgs),
			Size:              int64(stream.State.Bytes),
			CreatedAt:         stream.Created,
			ProviderData: map[string]interface{}{
				"subjects": stream.Config.Subjects,
				"storage":  stream.Config.Storage.String(),
			},
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

// PublishBatch publishes multiple messages
func (p *Provider) PublishBatch(ctx context.Context, request *gateway.PublishBatchRequest) (*gateway.PublishBatchResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	response := &gateway.PublishBatchResponse{
		PublishedCount: 0,
		FailedCount:    0,
		FailedMessages: make([]*gateway.Message, 0),
	}

	for _, message := range request.Messages {
		publishRequest := &gateway.PublishRequest{
			Topic:   request.Topic,
			Message: message,
			Options: request.Options,
		}

		_, err := p.PublishMessage(ctx, publishRequest)
		if err != nil {
			response.FailedCount++
			response.FailedMessages = append(response.FailedMessages, message)
			p.logger.WithError(err).WithField("message_id", message.ID).Error("Failed to publish message in batch")
		} else {
			response.PublishedCount++
		}
	}

	p.logger.WithFields(logrus.Fields{
		"topic":           request.Topic,
		"published_count": response.PublishedCount,
		"failed_count":    response.FailedCount,
	}).Info("Batch published")

	return response, nil
}

// GetTopicInfo gets stream information (JetStream only)
func (p *Provider) GetTopicInfo(ctx context.Context, request *gateway.GetTopicInfoRequest) (*gateway.TopicInfo, error) {
	if p.js == nil {
		return nil, fmt.Errorf("JetStream not enabled, cannot get topic info")
	}

	stream, err := p.js.StreamInfo(request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	topic := &gateway.TopicInfo{
		Name:         stream.Config.Name,
		MessageCount: int64(stream.State.Msgs),
		Size:         int64(stream.State.Bytes),
		CreatedAt:    stream.Created,
		ProviderData: map[string]interface{}{
			"subjects": stream.Config.Subjects,
			"storage":  stream.Config.Storage.String(),
		},
	}

	return topic, nil
}

// GetStats gets messaging statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.MessagingStats, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}

	stats := &gateway.MessagingStats{
		ActiveConnections:   1,
		ActiveSubscriptions: len(p.subs),
		ProviderData: map[string]interface{}{
			"connected":     p.conn.IsConnected(),
			"server_url":    p.conn.ConnectedUrl(),
			"server_id":     p.conn.ConnectedServerId(),
			"server_version": p.conn.ConnectedServerVersion(),
		},
	}

	// Add JetStream stats if available
	if p.js != nil {
		streams := p.js.Streams()
		totalMessages := int64(0)
		totalBytes := int64(0)

		for stream := range streams {
			totalMessages += int64(stream.State.Msgs)
			totalBytes += int64(stream.State.Bytes)
		}

		stats.PublishedMessages = totalMessages
		stats.ProviderData["jetstream_enabled"] = true
		stats.ProviderData["total_streams"] = len(streams)
		stats.ProviderData["total_messages"] = totalMessages
		stats.ProviderData["total_bytes"] = totalBytes
	}

	return stats, nil
}

// HealthCheck performs health check
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Try to ping the server
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("NATS health check failed: %w", err)
	}

	return nil
}

// Close closes the provider
func (p *Provider) Close() error {
	ctx := context.Background()
	return p.Disconnect(ctx)
}