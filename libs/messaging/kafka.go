package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// KafkaManager handles Kafka messaging
type KafkaManager struct {
	config  *KafkaConfig
	logger  *logrus.Logger
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

// KafkaMessage represents a Kafka message
type KafkaMessage struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Partition int                    `json:"partition"`
	Offset    int64                  `json:"offset"`
}

// KafkaMessageHandler handles incoming Kafka messages
type KafkaMessageHandler func(ctx context.Context, message *KafkaMessage) error

// NewKafkaManager creates a new Kafka manager
func NewKafkaManager(config *KafkaConfig, logger *logrus.Logger) *KafkaManager {
	return &KafkaManager{
		config:  config,
		logger:  logger,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}
}

// CreateWriter creates a Kafka writer for a topic
func (km *KafkaManager) CreateWriter(ctx context.Context, topic string) (*kafka.Writer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(km.config.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	km.writers[topic] = writer

	km.logger.WithFields(logrus.Fields{
		"topic":   topic,
		"brokers": km.config.Brokers,
	}).Info("Kafka writer created successfully")

	return writer, nil
}

// CreateReader creates a Kafka reader for a topic
func (km *KafkaManager) CreateReader(ctx context.Context, topic string) (*kafka.Reader, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        km.config.Brokers,
		Topic:          topic,
		GroupID:        km.config.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	km.readers[topic] = reader

	km.logger.WithFields(logrus.Fields{
		"topic":    topic,
		"group_id": km.config.GroupID,
		"brokers":  km.config.Brokers,
	}).Info("Kafka reader created successfully")

	return reader, nil
}

// PublishMessage publishes a message to Kafka
func (km *KafkaManager) PublishMessage(ctx context.Context, topic string, message *KafkaMessage) error {
	writer, exists := km.writers[topic]
	if !exists {
		var err error
		writer, err = km.CreateWriter(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to create writer: %w", err)
		}
	}

	message.ID = uuid.New()
	message.CreatedAt = time.Now()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMessage := kafka.Message{
		Key:   []byte(message.ID.String()),
		Value: body,
		Headers: []kafka.Header{
			{Key: "message-type", Value: []byte(message.Type)},
			{Key: "source", Value: []byte(message.Source)},
			{Key: "target", Value: []byte(message.Target)},
			{Key: "created-at", Value: []byte(message.CreatedAt.Format(time.RFC3339))},
		},
	}

	err = writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		km.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   message.ID,
			"topic":        topic,
			"message_type": message.Type,
		}).Error("Failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	km.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"topic":        topic,
		"message_type": message.Type,
		"source":       message.Source,
		"target":       message.Target,
	}).Info("Message published successfully")

	return nil
}

// SubscribeToTopic subscribes to a Kafka topic and handles messages
func (km *KafkaManager) SubscribeToTopic(ctx context.Context, topic string, handler KafkaMessageHandler) error {
	reader, exists := km.readers[topic]
	if !exists {
		var err error
		reader, err = km.CreateReader(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to create reader: %w", err)
		}
	}

	km.logger.WithField("topic", topic).Info("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				km.logger.WithField("topic", topic).Info("Stopped consuming messages")
				return
			default:
				km.handleMessage(ctx, reader, handler)
			}
		}
	}()

	return nil
}

// handleMessage handles a single Kafka message
func (km *KafkaManager) handleMessage(ctx context.Context, reader *kafka.Reader, handler KafkaMessageHandler) {
	message, err := reader.ReadMessage(ctx)
	if err != nil {
		km.logger.WithError(err).Error("Failed to read message")
		return
	}

	var kafkaMessage KafkaMessage
	if err := json.Unmarshal(message.Value, &kafkaMessage); err != nil {
		km.logger.WithError(err).WithFields(logrus.Fields{
			"topic":     message.Topic,
			"partition": message.Partition,
			"offset":    message.Offset,
		}).Error("Failed to unmarshal message")
		return
	}

	// Set Kafka-specific fields
	kafkaMessage.Partition = message.Partition
	kafkaMessage.Offset = message.Offset

	// Check if message is expired
	if kafkaMessage.ExpiresAt != nil && time.Now().After(*kafkaMessage.ExpiresAt) {
		km.logger.WithFields(logrus.Fields{
			"message_id": kafkaMessage.ID,
			"expires_at": kafkaMessage.ExpiresAt,
		}).Debug("Message expired, discarding")
		return
	}

	// Handle message
	if err := handler(ctx, &kafkaMessage); err != nil {
		km.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   kafkaMessage.ID,
			"message_type": kafkaMessage.Type,
			"source":       kafkaMessage.Source,
			"target":       kafkaMessage.Target,
			"topic":        message.Topic,
			"partition":    message.Partition,
			"offset":       message.Offset,
		}).Error("Failed to handle message")
		return
	}

	km.logger.WithFields(logrus.Fields{
		"message_id":   kafkaMessage.ID,
		"message_type": kafkaMessage.Type,
		"source":       kafkaMessage.Source,
		"target":       kafkaMessage.Target,
		"topic":        message.Topic,
		"partition":    message.Partition,
		"offset":       message.Offset,
	}).Debug("Message handled successfully")
}

// CreateTopic creates a new Kafka topic
func (km *KafkaManager) CreateTopic(ctx context.Context, topic string, partitions int, replicationFactor int) error {
	conn, err := kafka.DialLeader(ctx, "tcp", km.config.Brokers[0], topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	err = conn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	km.logger.WithFields(logrus.Fields{
		"topic":              topic,
		"partitions":         partitions,
		"replication_factor": replicationFactor,
	}).Info("Topic created successfully")

	return nil
}

// DeleteTopic deletes a Kafka topic
func (km *KafkaManager) DeleteTopic(ctx context.Context, topic string) error {
	conn, err := kafka.DialLeader(ctx, "tcp", km.config.Brokers[0], topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	err = conn.DeleteTopics(topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	km.logger.WithField("topic", topic).Info("Topic deleted successfully")
	return nil
}

// GetTopicInfo returns information about a topic
func (km *KafkaManager) GetTopicInfo(ctx context.Context, topic string) (kafka.Partition, error) {
	conn, err := kafka.DialLeader(ctx, "tcp", km.config.Brokers[0], topic, 0)
	if err != nil {
		return kafka.Partition{}, fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return kafka.Partition{}, fmt.Errorf("failed to read partitions: %w", err)
	}

	if len(partitions) == 0 {
		return kafka.Partition{}, fmt.Errorf("no partitions found for topic: %s", topic)
	}

	return partitions[0], nil
}

// HealthCheck performs a health check on Kafka connection
func (km *KafkaManager) HealthCheck(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", km.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("Kafka health check failed: %w", err)
	}
	defer conn.Close()

	// Try to get metadata
	_, err = conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("Kafka metadata check failed: %w", err)
	}

	return nil
}

// Close closes all Kafka connections
func (km *KafkaManager) Close() error {
	var lastErr error

	// Close all writers
	for topic, writer := range km.writers {
		if err := writer.Close(); err != nil {
			km.logger.WithError(err).WithField("topic", topic).Error("Failed to close writer")
			lastErr = err
		}
	}

	// Close all readers
	for topic, reader := range km.readers {
		if err := reader.Close(); err != nil {
			km.logger.WithError(err).WithField("topic", topic).Error("Failed to close reader")
			lastErr = err
		}
	}

	return lastErr
}

// PublishWithRetry publishes a message with retry logic
func (km *KafkaManager) PublishWithRetry(ctx context.Context, topic string, message *KafkaMessage, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := km.PublishMessage(ctx, topic, message); err != nil {
			lastErr = err
			if i < maxRetries {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to publish message after %d retries: %w", maxRetries, lastErr)
}

// CreateKafkaMessage creates a new Kafka message
func CreateKafkaMessage(messageType, source, target string, payload map[string]interface{}) *KafkaMessage {
	return &KafkaMessage{
		ID:        uuid.New(),
		Type:      messageType,
		Source:    source,
		Target:    target,
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// SetExpiration sets message expiration
func (m *KafkaMessage) SetExpiration(duration time.Duration) {
	expiresAt := time.Now().Add(duration)
	m.ExpiresAt = &expiresAt
}

// AddMetadata adds metadata to the message
func (m *KafkaMessage) AddMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata retrieves metadata from the message
func (m *KafkaMessage) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}
