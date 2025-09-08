package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// RabbitMQManager handles RabbitMQ messaging
type RabbitMQManager struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *RabbitMQConfig
	logger  *logrus.Logger
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL      string
	Exchange string
	Queue    string
	VHost    string
}

// Message represents a message to be sent
type Message struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, message *Message) error

// NewRabbitMQManager creates a new RabbitMQ manager
func NewRabbitMQManager(config *RabbitMQConfig, logger *logrus.Logger) (*RabbitMQManager, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		config.Exchange, // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	manager := &RabbitMQManager{
		conn:    conn,
		channel: channel,
		config:  config,
		logger:  logger,
	}

	logger.WithFields(logrus.Fields{
		"url":      config.URL,
		"exchange": config.Exchange,
		"queue":    config.Queue,
	}).Info("RabbitMQ manager initialized successfully")

	return manager, nil
}

// PublishMessage publishes a message to RabbitMQ
func (rm *RabbitMQManager) PublishMessage(ctx context.Context, routingKey string, message *Message) error {
	message.ID = uuid.New()
	message.CreatedAt = time.Now()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = rm.channel.PublishWithContext(
		ctx,
		rm.config.Exchange, // exchange
		routingKey,         // routing key
		false,              // mandatory
		false,              // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			MessageId:    message.ID.String(),
			Timestamp:    message.CreatedAt,
			DeliveryMode: amqp091.Persistent, // Make message persistent
		},
	)

	if err != nil {
		rm.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   message.ID,
			"routing_key":  routingKey,
			"message_type": message.Type,
		}).Error("Failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	rm.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"routing_key":  routingKey,
		"message_type": message.Type,
		"source":       message.Source,
		"target":       message.Target,
	}).Info("Message published successfully")

	return nil
}

// SubscribeToQueue subscribes to a queue and handles messages
func (rm *RabbitMQManager) SubscribeToQueue(ctx context.Context, queueName string, handler MessageHandler) error {
	// Declare queue
	queue, err := rm.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = rm.channel.QueueBind(
		queue.Name,         // queue name
		"#",                // routing key (all messages)
		rm.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Consume messages
	msgs, err := rm.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	rm.logger.WithField("queue", queueName).Info("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				rm.logger.WithField("queue", queueName).Info("Stopped consuming messages")
				return
			case delivery := <-msgs:
				rm.handleMessage(ctx, delivery, handler)
			}
		}
	}()

	return nil
}

// handleMessage handles a single message delivery
func (rm *RabbitMQManager) handleMessage(ctx context.Context, delivery amqp091.Delivery, handler MessageHandler) {
	var message Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		rm.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":  delivery.MessageId,
			"routing_key": delivery.RoutingKey,
		}).Error("Failed to unmarshal message")
		delivery.Nack(false, false) // Reject message
		return
	}

	// Check if message is expired
	if message.ExpiresAt != nil && time.Now().After(*message.ExpiresAt) {
		rm.logger.WithFields(logrus.Fields{
			"message_id": message.ID,
			"expires_at": message.ExpiresAt,
		}).Debug("Message expired, discarding")
		delivery.Ack(false) // Acknowledge expired message
		return
	}

	// Handle message
	if err := handler(ctx, &message); err != nil {
		rm.logger.WithError(err).WithFields(logrus.Fields{
			"message_id":   message.ID,
			"message_type": message.Type,
			"source":       message.Source,
			"target":       message.Target,
		}).Error("Failed to handle message")
		delivery.Nack(false, true) // Reject and requeue
		return
	}

	// Acknowledge message
	delivery.Ack(false)

	rm.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"message_type": message.Type,
		"source":       message.Source,
		"target":       message.Target,
	}).Debug("Message handled successfully")
}

// CreateQueue creates a new queue
func (rm *RabbitMQManager) CreateQueue(ctx context.Context, queueName string, durable bool) error {
	_, err := rm.channel.QueueDeclare(
		queueName, // name
		durable,   // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to create queue: %w", err)
	}

	rm.logger.WithFields(logrus.Fields{
		"queue":   queueName,
		"durable": durable,
	}).Info("Queue created successfully")

	return nil
}

// DeleteQueue deletes a queue
func (rm *RabbitMQManager) DeleteQueue(ctx context.Context, queueName string) error {
	_, err := rm.channel.QueueDelete(
		queueName, // name
		false,     // if-unused
		false,     // if-empty
		false,     // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	rm.logger.WithField("queue", queueName).Info("Queue deleted successfully")
	return nil
}

// BindQueue binds a queue to an exchange with a routing key
func (rm *RabbitMQManager) BindQueue(ctx context.Context, queueName, routingKey string) error {
	err := rm.channel.QueueBind(
		queueName,          // queue name
		routingKey,         // routing key
		rm.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	rm.logger.WithFields(logrus.Fields{
		"queue":       queueName,
		"routing_key": routingKey,
		"exchange":    rm.config.Exchange,
	}).Info("Queue bound successfully")

	return nil
}

// UnbindQueue unbinds a queue from an exchange
func (rm *RabbitMQManager) UnbindQueue(ctx context.Context, queueName, routingKey string) error {
	err := rm.channel.QueueUnbind(
		queueName,          // queue name
		routingKey,         // routing key
		rm.config.Exchange, // exchange
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to unbind queue: %w", err)
	}

	rm.logger.WithFields(logrus.Fields{
		"queue":       queueName,
		"routing_key": routingKey,
		"exchange":    rm.config.Exchange,
	}).Info("Queue unbound successfully")

	return nil
}

// GetQueueInfo returns information about a queue
func (rm *RabbitMQManager) GetQueueInfo(ctx context.Context, queueName string) (amqp091.Queue, error) {
	queue, err := rm.channel.QueueInspect(queueName)
	if err != nil {
		return amqp091.Queue{}, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return queue, nil
}

// HealthCheck performs a health check on RabbitMQ connection
func (rm *RabbitMQManager) HealthCheck(ctx context.Context) error {
	if rm.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}

	// Try to declare a temporary queue to test the connection
	tempQueue := fmt.Sprintf("health_check_%d", time.Now().Unix())
	_, err := rm.channel.QueueDeclare(
		tempQueue, // name
		false,     // durable
		true,      // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ health check failed: %w", err)
	}

	// Clean up temporary queue
	rm.channel.QueueDelete(tempQueue, false, false, false)

	return nil
}

// Close closes the RabbitMQ connection
func (rm *RabbitMQManager) Close() error {
	if rm.channel != nil {
		rm.channel.Close()
	}
	if rm.conn != nil {
		return rm.conn.Close()
	}
	return nil
}

// PublishWithRetry publishes a message with retry logic
func (rm *RabbitMQManager) PublishWithRetry(ctx context.Context, routingKey string, message *Message, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := rm.PublishMessage(ctx, routingKey, message); err != nil {
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

// CreateMessage creates a new message
func CreateMessage(messageType, source, target string, payload map[string]interface{}) *Message {
	return &Message{
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
func (m *Message) SetExpiration(duration time.Duration) {
	expiresAt := time.Now().Add(duration)
	m.ExpiresAt = &expiresAt
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
