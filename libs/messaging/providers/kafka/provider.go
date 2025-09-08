package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/libs/messaging/gateway"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// Provider implements MessagingProvider for Kafka
type Provider struct {
	config  map[string]interface{}
	logger  *logrus.Logger
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
}

// NewProvider creates a new Kafka messaging provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:  make(map[string]interface{}),
		logger:  logger,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "kafka"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.MessagingFeature {
	return []gateway.MessagingFeature{
		gateway.FeaturePublishSubscribe,
		gateway.FeatureMessageRouting,
		gateway.FeatureMessageOrdering,
		gateway.FeatureMessageDeduplication,
		gateway.FeatureMessageRetention,
		gateway.FeatureMessageCompression,
		gateway.FeatureMessageBatching,
		gateway.FeatureMessagePartitioning,
		gateway.FeatureMessageReplay,
		gateway.FeatureMessageHeaders,
		gateway.FeatureMessageCorrelation,
		gateway.FeatureMessageGrouping,
		gateway.FeatureMessageStreaming,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	brokers, _ := p.config["brokers"].([]string)
	host := "localhost"
	port := 9092

	if len(brokers) > 0 {
		// Parse first broker for host:port
		if len(brokers[0]) > 0 {
			host = brokers[0]
		}
	}

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     port,
		Protocol: "kafka",
		Version:  "2.8+",
	}
}

// Configure configures the Kafka provider
func (p *Provider) Configure(config map[string]interface{}) error {
	brokers, ok := config["brokers"].([]string)
	if !ok || len(brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	p.config = config

	p.logger.Info("Kafka provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	brokers, ok := p.config["brokers"].([]string)
	return ok && len(brokers) > 0
}

// Connect connects to Kafka
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)

	// Test connection by creating a temporary connection
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	p.logger.Info("Kafka connected successfully")
	return nil
}

// Disconnect disconnects from Kafka
func (p *Provider) Disconnect(ctx context.Context) error {
	// Close all writers
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.WithError(err).WithField("topic", topic).Error("Failed to close writer")
		}
	}

	// Close all readers
	for topic, reader := range p.readers {
		if err := reader.Close(); err != nil {
			p.logger.WithError(err).WithField("topic", topic).Error("Failed to close reader")
		}
	}

	p.logger.Info("Kafka disconnected successfully")
	return nil
}

// Ping checks Kafka connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to ping Kafka: %w", err)
	}
	defer conn.Close()

	// Try to get metadata
	_, err = conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("kafka metadata check failed: %w", err)
	}

	return nil
}

// IsConnected checks if Kafka is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.Ping(ctx) == nil
}

// PublishMessage publishes a message to Kafka
func (p *Provider) PublishMessage(ctx context.Context, request *gateway.PublishRequest) (*gateway.PublishResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("kafka provider not configured")
	}

	// Get or create writer
	writer, err := p.getOrCreateWriter(request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get writer: %w", err)
	}

	// Marshal message
	body, err := json.Marshal(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Key:   []byte(request.Message.ID.String()),
		Value: body,
		Headers: []kafka.Header{
			{Key: "message-type", Value: []byte(request.Message.Type)},
			{Key: "source", Value: []byte(request.Message.Source)},
			{Key: "target", Value: []byte(request.Message.Target)},
			{Key: "created-at", Value: []byte(request.Message.CreatedAt.Format(time.RFC3339))},
		},
	}

	// Add custom headers
	for key, value := range request.Headers {
		if str, ok := value.(string); ok {
			kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
				Key:   key,
				Value: []byte(str),
			})
		}
	}

	// Write message
	err = writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	response := &gateway.PublishResponse{
		MessageID: request.Message.ID.String(),
		Topic:     request.Topic,
		Timestamp: time.Now(),
		ProviderData: map[string]interface{}{
			"brokers": p.config["brokers"],
		},
	}

	return response, nil
}

// SubscribeToTopic subscribes to a Kafka topic
func (p *Provider) SubscribeToTopic(ctx context.Context, request *gateway.SubscribeRequest, handler gateway.MessageHandler) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	// Get or create reader
	reader, err := p.getOrCreateReader(request)
	if err != nil {
		return fmt.Errorf("failed to get reader: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.WithField("topic", request.Topic).Info("Stopped consuming messages")
				return
			default:
				p.handleMessage(ctx, reader, handler)
			}
		}
	}()

	return nil
}

// UnsubscribeFromTopic unsubscribes from a Kafka topic
func (p *Provider) UnsubscribeFromTopic(ctx context.Context, request *gateway.UnsubscribeRequest) error {
	if reader, exists := p.readers[request.Topic]; exists {
		if err := reader.Close(); err != nil {
			return fmt.Errorf("failed to close reader: %w", err)
		}
		delete(p.readers, request.Topic)
	}

	p.logger.WithField("topic", request.Topic).Info("Unsubscribed from topic")
	return nil
}

// CreateTopic creates a Kafka topic
func (p *Provider) CreateTopic(ctx context.Context, request *gateway.CreateTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], request.Topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             request.Topic,
		NumPartitions:     request.Partitions,
		ReplicationFactor: request.ReplicationFactor,
	}

	err = conn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"topic":              request.Topic,
		"partitions":         request.Partitions,
		"replication_factor": request.ReplicationFactor,
	}).Info("Topic created successfully")

	return nil
}

// DeleteTopic deletes a Kafka topic
func (p *Provider) DeleteTopic(ctx context.Context, request *gateway.DeleteTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], request.Topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	err = conn.DeleteTopics(request.Topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Topic deleted successfully")
	return nil
}

// TopicExists checks if a Kafka topic exists
func (p *Provider) TopicExists(ctx context.Context, request *gateway.TopicExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], request.Topic, 0)
	if err != nil {
		return false, fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(request.Topic)
	if err != nil {
		return false, fmt.Errorf("failed to read partitions: %w", err)
	}

	return len(partitions) > 0, nil
}

// ListTopics lists Kafka topics
func (p *Provider) ListTopics(ctx context.Context) ([]gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Group partitions by topic
	topicMap := make(map[string]*gateway.TopicInfo)
	for _, partition := range partitions {
		if info, exists := topicMap[partition.Topic]; exists {
			info.Partitions++
		} else {
			topicMap[partition.Topic] = &gateway.TopicInfo{
				Name:       partition.Topic,
				Partitions: 1,
				ProviderData: map[string]interface{}{
					"broker": partition.Leader.Host,
				},
			}
		}
	}

	// Convert map to slice
	topics := make([]gateway.TopicInfo, 0, len(topicMap))
	for _, info := range topicMap {
		topics = append(topics, *info)
	}

	return topics, nil
}

// PublishBatch publishes multiple messages to Kafka
func (p *Provider) PublishBatch(ctx context.Context, request *gateway.PublishBatchRequest) (*gateway.PublishBatchResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("kafka provider not configured")
	}

	// Get or create writer
	writer, err := p.getOrCreateWriter(request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get writer: %w", err)
	}

	// Prepare Kafka messages
	kafkaMessages := make([]kafka.Message, len(request.Messages))
	for i, msg := range request.Messages {
		body, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		kafkaMessages[i] = kafka.Message{
			Key:   []byte(msg.ID.String()),
			Value: body,
			Headers: []kafka.Header{
				{Key: "message-type", Value: []byte(msg.Type)},
				{Key: "source", Value: []byte(msg.Source)},
				{Key: "target", Value: []byte(msg.Target)},
			},
		}
	}

	// Write messages
	err = writer.WriteMessages(ctx, kafkaMessages...)
	if err != nil {
		return nil, fmt.Errorf("failed to write messages: %w", err)
	}

	response := &gateway.PublishBatchResponse{
		PublishedCount: len(request.Messages),
		FailedCount:    0,
		ProviderData: map[string]interface{}{
			"brokers": p.config["brokers"],
		},
	}

	return response, nil
}

// GetTopicInfo gets Kafka topic information
func (p *Provider) GetTopicInfo(ctx context.Context, request *gateway.GetTopicInfoRequest) (*gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], request.Topic, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("topic not found: %s", request.Topic)
	}

	info := &gateway.TopicInfo{
		Name:       request.Topic,
		Partitions: len(partitions),
		ProviderData: map[string]interface{}{
			"brokers": p.config["brokers"],
		},
	}

	return info, nil
}

// GetStats returns Kafka statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.MessagingStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("kafka provider not configured")
	}

	stats := &gateway.MessagingStats{
		ActiveConnections:   len(p.writers) + len(p.readers),
		ActiveSubscriptions: len(p.readers),
		ProviderData: map[string]interface{}{
			"writers": len(p.writers),
			"readers": len(p.readers),
			"brokers": p.config["brokers"],
		},
	}

	return stats, nil
}

// HealthCheck performs a health check on Kafka
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("kafka provider not configured")
	}

	brokers, _ := p.config["brokers"].([]string)
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}
	defer conn.Close()

	// Try to get metadata
	_, err = conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("kafka metadata check failed: %w", err)
	}

	return nil
}

// Close closes the Kafka provider
func (p *Provider) Close() error {
	return p.Disconnect(context.Background())
}

// getOrCreateWriter gets or creates a Kafka writer for a topic
func (p *Provider) getOrCreateWriter(topic string) (*kafka.Writer, error) {
	if writer, exists := p.writers[topic]; exists {
		return writer, nil
	}

	brokers, _ := p.config["brokers"].([]string)
	groupID, _ := p.config["group_id"].(string)

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	p.writers[topic] = writer

	p.logger.WithFields(logrus.Fields{
		"topic":   topic,
		"brokers": brokers,
	}).Info("Kafka writer created successfully")

	return writer, nil
}

// getOrCreateReader gets or creates a Kafka reader for a topic
func (p *Provider) getOrCreateReader(request *gateway.SubscribeRequest) (*kafka.Reader, error) {
	if reader, exists := p.readers[request.Topic]; exists {
		return reader, nil
	}

	brokers, _ := p.config["brokers"].([]string)
	groupID := request.GroupID
	if groupID == "" {
		groupID, _ = p.config["group_id"].(string)
	}
	if groupID == "" {
		groupID = "default-group"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          request.Topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	p.readers[request.Topic] = reader

	p.logger.WithFields(logrus.Fields{
		"topic":    request.Topic,
		"group_id": groupID,
		"brokers":  brokers,
	}).Info("Kafka reader created successfully")

	return reader, nil
}

// handleMessage handles a single Kafka message
func (p *Provider) handleMessage(ctx context.Context, reader *kafka.Reader, handler gateway.MessageHandler) {
	message, err := reader.ReadMessage(ctx)
	if err != nil {
		p.logger.WithError(err).Error("Failed to read message")
		return
	}

	var gatewayMessage gateway.Message
	if err := json.Unmarshal(message.Value, &gatewayMessage); err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"topic":     message.Topic,
			"partition": message.Partition,
			"offset":    message.Offset,
		}).Error("Failed to unmarshal message")
		return
	}

	// Set Kafka-specific fields
	gatewayMessage.ProviderData = map[string]interface{}{
		"partition": message.Partition,
		"offset":    message.Offset,
		"topic":     message.Topic,
	}

	// Check if message is expired
	if gatewayMessage.ExpiresAt != nil && time.Now().After(*gatewayMessage.ExpiresAt) {
		p.logger.WithFields(logrus.Fields{
			"message_id": gatewayMessage.ID,
			"expires_at": gatewayMessage.ExpiresAt,
		}).Debug("Message expired, discarding")
		return
	}

	// Handle message
	if err := handler(ctx, &gatewayMessage); err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   gatewayMessage.ID,
			"message_type": gatewayMessage.Type,
			"source":       gatewayMessage.Source,
			"target":       gatewayMessage.Target,
			"topic":        message.Topic,
			"partition":    message.Partition,
			"offset":       message.Offset,
		}).Error("Failed to handle message")
		return
	}

	p.logger.WithFields(logrus.Fields{
		"message_id":   gatewayMessage.ID,
		"message_type": gatewayMessage.Type,
		"source":       gatewayMessage.Source,
		"target":       gatewayMessage.Target,
		"topic":        message.Topic,
		"partition":    message.Partition,
		"offset":       message.Offset,
	}).Debug("Message handled successfully")
}
