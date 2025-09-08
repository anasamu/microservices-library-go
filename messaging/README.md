# Messaging Gateway Library

A comprehensive, modular, and production-ready messaging library for Go microservices. This library provides a unified interface for multiple messaging providers including Kafka, RabbitMQ, and AWS SQS.

## 🚀 Features

### 🔧 Multi-Provider Support
- **Kafka**: Full Kafka support with topics, partitions, and consumer groups
- **RabbitMQ**: Complete RabbitMQ integration with exchanges, queues, and routing
- **AWS SQS**: AWS SQS support for scalable message queuing

### 📊 Core Operations
- **Message Publishing**: Publish messages to topics/queues
- **Message Subscription**: Subscribe to topics/queues with handlers
- **Topic Management**: Create, delete, and manage topics/queues
- **Batch Operations**: Publish and consume messages in batches
- **Message Routing**: Route messages based on routing keys and filters

### 🔗 Advanced Features
- **Message Headers**: Custom headers and metadata support
- **Message Expiration**: TTL and expiration handling
- **Message Scheduling**: Delayed message delivery
- **Message Priority**: Priority-based message processing
- **Message Correlation**: Request-reply pattern support
- **Message Deduplication**: Prevent duplicate message processing
- **Message Compression**: Automatic message compression
- **Message Encryption**: Secure message transmission

### 🏥 Production Features
- **Connection Management**: Automatic connection lifecycle management
- **Retry Logic**: Configurable retry with exponential backoff
- **Health Monitoring**: Real-time messaging system health monitoring
- **Statistics**: Detailed messaging statistics and metrics
- **Error Handling**: Comprehensive error reporting with context
- **Logging**: Structured logging with detailed context

## 📁 Project Structure

```
libs/messaging/
├── gateway/                    # Core messaging gateway
│   ├── manager.go             # Messaging manager implementation
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Messaging provider implementations
│   ├── kafka/                 # Kafka provider
│   │   ├── provider.go        # Kafka implementation
│   │   └── go.mod             # Kafka dependencies
│   ├── rabbitmq/              # RabbitMQ provider
│   │   ├── provider.go        # RabbitMQ implementation
│   │   └── go.mod             # RabbitMQ dependencies
│   └── sqs/                   # AWS SQS provider
│       ├── provider.go        # SQS implementation
│       └── go.mod             # SQS dependencies
├── go.mod                     # Main module dependencies
└── README.md                  # This file
```

## 🛠️ Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Basic Installation

```bash
# Clone the repository
git clone https://github.com/anasamu/microservices-library-go.git
cd microservices-library-go/libs/messaging

# Install dependencies
go mod tidy
```

### Using Specific Providers

```bash
# For Kafka support
go get github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka

# For RabbitMQ support
go get github.com/anasamu/microservices-library-go/libs/messaging/providers/rabbitmq

# For AWS SQS support
go get github.com/anasamu/microservices-library-go/libs/messaging/providers/sqs
```

## 📖 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/libs/messaging/gateway"
    "github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create messaging manager
    config := gateway.DefaultManagerConfig()
    config.DefaultProvider = "kafka"
    config.MaxMessageSize = 1024 * 1024 // 1MB
    
    messagingManager := gateway.NewMessagingManager(config, logger)
    
    // Register Kafka provider
    kafkaProvider := kafka.NewProvider(logger)
    kafkaConfig := map[string]interface{}{
        "brokers":  []string{"localhost:9092"},
        "group_id": "example-group",
    }
    
    if err := kafkaProvider.Configure(kafkaConfig); err != nil {
        log.Fatal(err)
    }
    
    messagingManager.RegisterProvider(kafkaProvider)
    
    // Use the messaging manager...
}
```

### Connect to Messaging System

```go
// Connect to messaging system
ctx := context.Background()
err := messagingManager.Connect(ctx, "kafka")
if err != nil {
    log.Fatal(err)
}

// Check connection
if messagingManager.IsProviderConnected("kafka") {
    log.Println("Messaging system connected successfully")
}

// Ping messaging system
err = messagingManager.Ping(ctx, "kafka")
if err != nil {
    log.Fatal(err)
}
```

### Publish Messages

```go
// Create a message
message := gateway.CreateMessage("user.created", "user-service", "notification-service", "user-events", map[string]interface{}{
    "user_id":    "12345",
    "email":      "user@example.com",
    "first_name": "John",
    "last_name":  "Doe",
})

// Add metadata
message.AddMetadata("priority", "high")
message.AddMetadata("version", "1.0")

// Set expiration
message.SetExpiration(24 * time.Hour)

// Publish message
request := &gateway.PublishRequest{
    Topic:   "user-events",
    Message: message,
}

response, err := messagingManager.PublishMessage(ctx, "kafka", request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Message published: %s", response.MessageID)
```

### Subscribe to Topics

```go
// Create message handler
handler := func(ctx context.Context, message *gateway.Message) error {
    log.Printf("Received message: %s from %s", message.Type, message.Source)
    log.Printf("Payload: %+v", message.Payload)
    return nil
}

// Subscribe to topic
request := &gateway.SubscribeRequest{
    Topic:         "user-events",
    GroupID:       "example-group",
    AutoAck:       true,
    PrefetchCount: 10,
}

err := messagingManager.SubscribeToTopic(ctx, "kafka", request, handler)
if err != nil {
    log.Fatal(err)
}
```

### Topic Management

```go
// Create topic
createRequest := &gateway.CreateTopicRequest{
    Topic:             "user-events",
    Partitions:        3,
    ReplicationFactor: 1,
}

err := messagingManager.CreateTopic(ctx, "kafka", createRequest)
if err != nil {
    log.Fatal(err)
}

// Check if topic exists
existsRequest := &gateway.TopicExistsRequest{
    Topic: "user-events",
}

exists, err := messagingManager.TopicExists(ctx, "kafka", existsRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Topic exists: %t", exists)

// List topics
topics, err := messagingManager.ListTopics(ctx, "kafka")
if err != nil {
    log.Fatal(err)
}

log.Printf("Topics: %d found", len(topics))
```

### Batch Operations

```go
// Create multiple messages
messages := make([]*gateway.Message, 3)
for i := 0; i < 3; i++ {
    messages[i] = gateway.CreateMessage("batch.test", "batch-service", "consumer-service", "batch-events", map[string]interface{}{
        "batch_id": fmt.Sprintf("batch-%d", i+1),
        "data":     fmt.Sprintf("message-%d", i+1),
    })
}

// Publish batch
batchRequest := &gateway.PublishBatchRequest{
    Topic:    "batch-events",
    Messages: messages,
}

response, err := messagingManager.PublishBatch(ctx, "kafka", batchRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Batch published: %d successful, %d failed", response.PublishedCount, response.FailedCount)
```

### Advanced Message Features

```go
// Create message with advanced features
message := gateway.CreateMessage("advanced.test", "advanced-service", "consumer-service", "advanced-events", map[string]interface{}{
    "data": "advanced message",
})

// Add headers
message.AddHeader("content-type", "application/json")
message.AddHeader("version", "2.0")

// Add metadata
message.AddMetadata("priority", "high")
message.AddMetadata("retry-count", 0)

// Set advanced properties
message.SetExpiration(1 * time.Hour)
message.SetPriority(10)
message.SetCorrelationID("advanced-123")
message.SetReplyTo("advanced-service-reply")

// Publish with custom headers
request := &gateway.PublishRequest{
    Topic:   "advanced-events",
    Message: message,
    Headers: map[string]interface{}{
        "custom-header": "custom-value",
    },
}

response, err := messagingManager.PublishMessage(ctx, "kafka", request)
if err != nil {
    log.Fatal(err)
}
```

### Health Checks

```go
// Perform health checks
results := messagingManager.HealthCheck(ctx)

for provider, err := range results {
    if err != nil {
        log.Printf("%s: ❌ %v", provider, err)
    } else {
        log.Printf("%s: ✅ Healthy", provider)
    }
}
```

### Messaging Statistics

```go
// Get messaging statistics
stats, err := messagingManager.GetStats(ctx, "kafka")
if err != nil {
    log.Fatal(err)
}

log.Printf("Published Messages: %d", stats.PublishedMessages)
log.Printf("Consumed Messages: %d", stats.ConsumedMessages)
log.Printf("Failed Messages: %d", stats.FailedMessages)
log.Printf("Active Connections: %d", stats.ActiveConnections)
log.Printf("Active Subscriptions: %d", stats.ActiveSubscriptions)
```

## 🔧 Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Messaging Manager Configuration
export MESSAGING_DEFAULT_PROVIDER="kafka"
export MESSAGING_MAX_MESSAGE_SIZE="1048576"
export MESSAGING_RETRY_ATTEMPTS="3"
export MESSAGING_RETRY_DELAY="5s"
export MESSAGING_TIMEOUT="30s"

# Kafka Configuration
export KAFKA_BROKERS="localhost:9092"
export KAFKA_GROUP_ID="example-group"

# RabbitMQ Configuration
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export RABBITMQ_EXCHANGE="example-exchange"

# AWS SQS Configuration
export SQS_REGION="us-east-1"
export SQS_ACCESS_KEY="your-access-key"
export SQS_SECRET_KEY="your-secret-key"
```

### Configuration Files

You can also use configuration files:

```json
{
  "messaging": {
    "default_provider": "kafka",
    "max_message_size": 1048576,
    "retry_attempts": 3,
    "retry_delay": "5s",
    "timeout": "30s"
  },
  "providers": {
    "kafka": {
      "brokers": ["localhost:9092"],
      "group_id": "example-group"
    },
    "rabbitmq": {
      "url": "amqp://guest:guest@localhost:5672/",
      "exchange": "example-exchange"
    },
    "sqs": {
      "region": "us-east-1",
      "access_key": "your-access-key",
      "secret_key": "your-secret-key"
    }
  }
}
```

## 🧪 Testing

Run tests for all modules:

```bash
# Run all tests
go test ./...

# Run tests for specific provider
go test ./providers/kafka/...
go test ./providers/rabbitmq/...
go test ./providers/sqs/...

# Run gateway tests
go test ./gateway/...
```

## 📚 API Documentation

### Messaging Manager API

- `NewMessagingManager(config, logger)` - Create messaging manager
- `RegisterProvider(provider)` - Register a messaging provider
- `Connect(ctx, provider)` - Connect to a messaging system
- `Disconnect(ctx, provider)` - Disconnect from a messaging system
- `Ping(ctx, provider)` - Ping a messaging system
- `PublishMessage(ctx, provider, request)` - Publish a message
- `SubscribeToTopic(ctx, provider, request, handler)` - Subscribe to a topic
- `UnsubscribeFromTopic(ctx, provider, request)` - Unsubscribe from a topic
- `CreateTopic(ctx, provider, request)` - Create a topic
- `DeleteTopic(ctx, provider, request)` - Delete a topic
- `TopicExists(ctx, provider, request)` - Check if topic exists
- `ListTopics(ctx, provider)` - List topics
- `PublishBatch(ctx, provider, request)` - Publish multiple messages
- `GetTopicInfo(ctx, provider, request)` - Get topic information
- `HealthCheck(ctx)` - Check provider health
- `GetStats(ctx, provider)` - Get messaging statistics

### Provider Interface

All providers implement the `MessagingProvider` interface:

```go
type MessagingProvider interface {
    GetName() string
    GetSupportedFeatures() []MessagingFeature
    GetConnectionInfo() *ConnectionInfo
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Ping(ctx context.Context) error
    IsConnected() bool
    PublishMessage(ctx context.Context, request *PublishRequest) (*PublishResponse, error)
    SubscribeToTopic(ctx context.Context, request *SubscribeRequest, handler MessageHandler) error
    UnsubscribeFromTopic(ctx context.Context, request *UnsubscribeRequest) error
    CreateTopic(ctx context.Context, request *CreateTopicRequest) error
    DeleteTopic(ctx context.Context, request *DeleteTopicRequest) error
    TopicExists(ctx context.Context, request *TopicExistsRequest) (bool, error)
    ListTopics(ctx context.Context) ([]TopicInfo, error)
    PublishBatch(ctx context.Context, request *PublishBatchRequest) (*PublishBatchResponse, error)
    GetTopicInfo(ctx context.Context, request *GetTopicInfoRequest) (*TopicInfo, error)
    GetStats(ctx context.Context) (*MessagingStats, error)
    HealthCheck(ctx context.Context) error
    Configure(config map[string]interface{}) error
    IsConfigured() bool
    Close() error
}
```

### Supported Features

- `FeaturePublishSubscribe` - Publish-subscribe messaging
- `FeatureRequestReply` - Request-reply pattern
- `FeatureMessageRouting` - Message routing
- `FeatureMessageFiltering` - Message filtering
- `FeatureMessageOrdering` - Message ordering
- `FeatureMessageDeduplication` - Message deduplication
- `FeatureMessageRetention` - Message retention
- `FeatureMessageCompression` - Message compression
- `FeatureMessageEncryption` - Message encryption
- `FeatureMessageBatching` - Message batching
- `FeatureMessagePartitioning` - Message partitioning
- `FeatureMessageReplay` - Message replay
- `FeatureMessageDeadLetter` - Dead letter queues
- `FeatureMessageScheduling` - Message scheduling
- `FeatureMessagePriority` - Message priority
- `FeatureMessageTTL` - Message TTL
- `FeatureMessageHeaders` - Message headers
- `FeatureMessageCorrelation` - Message correlation
- `FeatureMessageGrouping` - Message grouping
- `FeatureMessageStreaming` - Message streaming

## 🔒 Security Considerations

### Connection Security

- **SSL/TLS**: All providers support encrypted connections
- **Authentication**: Multiple authentication methods per provider
- **Connection Pooling**: Secure connection management
- **Timeout Handling**: Prevents hanging connections

### Message Security

- **Message Encryption**: Encrypted message transmission
- **Message Signing**: Message integrity verification
- **Access Control**: Topic/queue access control
- **Message Validation**: Message content validation

## 🚀 Performance

### Optimization Features

- **Connection Pooling**: Efficient connection management
- **Message Batching**: Efficient bulk operations
- **Retry Logic**: Automatic retry with backoff
- **Connection Reuse**: Minimize connection overhead
- **Message Compression**: Reduce message size

### Monitoring

- **Health Checks**: Real-time messaging system health monitoring
- **Statistics**: Detailed performance metrics
- **Connection Monitoring**: Connection pool statistics
- **Message Performance**: Message processing monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our Discord](https://discord.gg/example)
- 📖 Documentation: [Full Documentation](https://docs.example.com)
- 🐛 Issues: [GitHub Issues](https://github.com/anasamu/microservices-library-go/issues)

## 🙏 Acknowledgments

- [Apache Kafka](https://kafka.apache.org/) for the distributed streaming platform
- [RabbitMQ](https://www.rabbitmq.com/) for the message broker
- [AWS SQS](https://aws.amazon.com/sqs/) for the managed message queuing service
- [Go Kafka Client](https://github.com/segmentio/kafka-go) for Kafka integration
- [RabbitMQ Go Client](https://github.com/rabbitmq/amqp091-go) for RabbitMQ integration
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go-v2) for AWS SQS integration

---

Made with ❤️ for the Go microservices community
