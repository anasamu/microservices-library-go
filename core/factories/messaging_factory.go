package factories

import (
	"context"
	"fmt"

	"github.com/anasamu/microservices-library-go/core/interfaces"
	"github.com/anasamu/microservices-library-go/messaging"
)

// MessageQueueType represents the type of message queue
type MessageQueueType string

const (
	MessageQueueTypeKafka    MessageQueueType = "kafka"
	MessageQueueTypeRabbitMQ MessageQueueType = "rabbitmq"
	MessageQueueTypeNATS     MessageQueueType = "nats"
	MessageQueueTypeRedis    MessageQueueType = "redis"
)

// MessageQueueFactory creates message queue instances
type MessageQueueFactory struct {
	configs map[MessageQueueType]interface{}
}

// NewMessageQueueFactory creates a new message queue factory
func NewMessageQueueFactory() *MessageQueueFactory {
	return &MessageQueueFactory{
		configs: make(map[MessageQueueType]interface{}),
	}
}

// RegisterConfig registers a message queue configuration
func (f *MessageQueueFactory) RegisterConfig(queueType MessageQueueType, config interface{}) {
	f.configs[queueType] = config
}

// CreateMessageQueue creates a message queue instance based on type
func (f *MessageQueueFactory) CreateMessageQueue(ctx context.Context, queueType MessageQueueType) (interfaces.MessageQueue, error) {
	config, exists := f.configs[queueType]
	if !exists {
		return nil, fmt.Errorf("no configuration found for message queue type: %s", queueType)
	}

	switch queueType {
	case MessageQueueTypeKafka:
		kafkaConfig, ok := config.(*messaging.KafkaConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for Kafka")
		}
		return messaging.NewKafkaProducer(kafkaConfig)

	case MessageQueueTypeRabbitMQ:
		rabbitConfig, ok := config.(*messaging.RabbitMQConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for RabbitMQ")
		}
		return messaging.NewRabbitMQProducer(rabbitConfig)

	case MessageQueueTypeNATS:
		natsConfig, ok := config.(*messaging.NATSConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for NATS")
		}
		return messaging.NewNATSProducer(natsConfig)

	case MessageQueueTypeRedis:
		redisConfig, ok := config.(*messaging.RedisStreamConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for Redis Streams")
		}
		return messaging.NewRedisStreamProducer(redisConfig)

	default:
		return nil, fmt.Errorf("unsupported message queue type: %s", queueType)
	}
}

// CreateMultipleMessageQueues creates multiple message queue instances
func (f *MessageQueueFactory) CreateMultipleMessageQueues(ctx context.Context, queueTypes []MessageQueueType) (map[MessageQueueType]interfaces.MessageQueue, error) {
	queues := make(map[MessageQueueType]interfaces.MessageQueue)

	for _, queueType := range queueTypes {
		queue, err := f.CreateMessageQueue(ctx, queueType)
		if err != nil {
			// Close already created queues
			for _, createdQueue := range queues {
				createdQueue.Close()
			}
			return nil, fmt.Errorf("failed to create message queue %s: %w", queueType, err)
		}
		queues[queueType] = queue
	}

	return queues, nil
}

// GetSupportedTypes returns supported message queue types
func (f *MessageQueueFactory) GetSupportedTypes() []MessageQueueType {
	return []MessageQueueType{
		MessageQueueTypeKafka,
		MessageQueueTypeRabbitMQ,
		MessageQueueTypeNATS,
		MessageQueueTypeRedis,
	}
}

// IsTypeSupported checks if a message queue type is supported
func (f *MessageQueueFactory) IsTypeSupported(queueType MessageQueueType) bool {
	supportedTypes := f.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if supportedType == queueType {
			return true
		}
	}
	return false
}
