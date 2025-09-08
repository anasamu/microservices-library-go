package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka"
	"github.com/anasamu/microservices-library-go/libs/messaging/providers/rabbitmq"
	"github.com/anasamu/microservices-library-go/libs/messaging/providers/sqs"
	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the messaging gateway system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create messaging manager
	config := DefaultManagerConfig()
	config.DefaultProvider = "kafka"
	config.MaxMessageSize = 1024 * 1024 // 1MB
	config.RetryAttempts = 3
	config.RetryDelay = 5 * time.Second

	messagingManager := NewMessagingManager(config, logger)

	// Register messaging providers
	registerProviders(messagingManager, logger)

	// Example 1: Connect to Kafka
	exampleKafkaConnection(messagingManager)

	// Example 2: Connect to RabbitMQ
	exampleRabbitMQConnection(messagingManager)

	// Example 3: Connect to AWS SQS
	exampleSQSConnection(messagingManager)

	// Example 4: Publish messages
	examplePublishMessages(messagingManager)

	// Example 5: Subscribe to topics
	exampleSubscribeToTopics(messagingManager)

	// Example 6: Topic management
	exampleTopicManagement(messagingManager)

	// Example 7: Batch operations
	exampleBatchOperations(messagingManager)

	// Example 8: Health checks
	exampleHealthChecks(messagingManager)

	// Example 9: Messaging statistics
	exampleMessagingStats(messagingManager)
}

// registerProviders registers all messaging providers
func registerProviders(messagingManager *MessagingManager, logger *logrus.Logger) {
	// Register Kafka provider
	kafkaProvider := kafka.NewProvider(logger)
	kafkaConfig := map[string]interface{}{
		"brokers":  []string{"localhost:9092"},
		"group_id": "example-group",
	}
	if err := kafkaProvider.Configure(kafkaConfig); err != nil {
		log.Printf("Failed to configure Kafka: %v", err)
	} else {
		messagingManager.RegisterProvider(kafkaProvider)
	}

	// Register RabbitMQ provider
	rabbitmqProvider := rabbitmq.NewProvider(logger)
	rabbitmqConfig := map[string]interface{}{
		"url":      "amqp://guest:guest@localhost:5672/",
		"exchange": "example-exchange",
	}
	if err := rabbitmqProvider.Configure(rabbitmqConfig); err != nil {
		log.Printf("Failed to configure RabbitMQ: %v", err)
	} else {
		messagingManager.RegisterProvider(rabbitmqProvider)
	}

	// Register AWS SQS provider
	sqsProvider := sqs.NewProvider(logger)
	sqsConfig := map[string]interface{}{
		"region":     "us-east-1",
		"access_key": "your-access-key",
		"secret_key": "your-secret-key",
	}
	if err := sqsProvider.Configure(sqsConfig); err != nil {
		log.Printf("Failed to configure SQS: %v", err)
	} else {
		messagingManager.RegisterProvider(sqsProvider)
	}
}

// exampleKafkaConnection demonstrates Kafka connection
func exampleKafkaConnection(messagingManager *MessagingManager) {
	fmt.Println("=== Kafka Connection Example ===")

	ctx := context.Background()

	// Connect to Kafka
	err := messagingManager.Connect(ctx, "kafka")
	if err != nil {
		log.Printf("Failed to connect to Kafka: %v", err)
		return
	}

	// Check if connected
	if messagingManager.IsProviderConnected("kafka") {
		fmt.Println("Kafka connected successfully")
	}

	// Ping Kafka
	err = messagingManager.Ping(ctx, "kafka")
	if err != nil {
		log.Printf("Failed to ping Kafka: %v", err)
	} else {
		fmt.Println("Kafka ping successful")
	}
}

// exampleRabbitMQConnection demonstrates RabbitMQ connection
func exampleRabbitMQConnection(messagingManager *MessagingManager) {
	fmt.Println("\n=== RabbitMQ Connection Example ===")

	ctx := context.Background()

	// Connect to RabbitMQ
	err := messagingManager.Connect(ctx, "rabbitmq")
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return
	}

	// Check if connected
	if messagingManager.IsProviderConnected("rabbitmq") {
		fmt.Println("RabbitMQ connected successfully")
	}

	// Ping RabbitMQ
	err = messagingManager.Ping(ctx, "rabbitmq")
	if err != nil {
		log.Printf("Failed to ping RabbitMQ: %v", err)
	} else {
		fmt.Println("RabbitMQ ping successful")
	}
}

// exampleSQSConnection demonstrates AWS SQS connection
func exampleSQSConnection(messagingManager *MessagingManager) {
	fmt.Println("\n=== AWS SQS Connection Example ===")

	ctx := context.Background()

	// Connect to SQS
	err := messagingManager.Connect(ctx, "sqs")
	if err != nil {
		log.Printf("Failed to connect to SQS: %v", err)
		return
	}

	// Check if connected
	if messagingManager.IsProviderConnected("sqs") {
		fmt.Println("AWS SQS connected successfully")
	}

	// Ping SQS
	err = messagingManager.Ping(ctx, "sqs")
	if err != nil {
		log.Printf("Failed to ping SQS: %v", err)
	} else {
		fmt.Println("AWS SQS ping successful")
	}
}

// examplePublishMessages demonstrates message publishing
func examplePublishMessages(messagingManager *MessagingManager) {
	fmt.Println("\n=== Publish Messages Example ===")

	ctx := context.Background()

	// Example with Kafka
	if messagingManager.IsProviderConnected("kafka") {
		fmt.Println("Publishing to Kafka:")

		// Create a message
		message := CreateMessage("user.created", "user-service", "notification-service", "user-events", map[string]interface{}{
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
		request := &PublishRequest{
			Topic:   "user-events",
			Message: message,
		}

		response, err := messagingManager.PublishMessage(ctx, "kafka", request)
		if err != nil {
			log.Printf("Failed to publish to Kafka: %v", err)
		} else {
			fmt.Printf("Message published to Kafka: %s\n", response.MessageID)
		}
	}

	// Example with RabbitMQ
	if messagingManager.IsProviderConnected("rabbitmq") {
		fmt.Println("\nPublishing to RabbitMQ:")

		// Create a message
		message := CreateMessage("order.processed", "order-service", "payment-service", "order-events", map[string]interface{}{
			"order_id":    "67890",
			"amount":      99.99,
			"currency":    "USD",
			"customer_id": "12345",
		})

		// Set correlation ID for request-reply pattern
		message.SetCorrelationID("order-67890")
		message.SetReplyTo("order-service-reply")

		// Publish message
		request := &PublishRequest{
			Topic:      "order-events",
			Message:    message,
			RoutingKey: "order.processed",
		}

		response, err := messagingManager.PublishMessage(ctx, "rabbitmq", request)
		if err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
		} else {
			fmt.Printf("Message published to RabbitMQ: %s\n", response.MessageID)
		}
	}

	// Example with AWS SQS
	if messagingManager.IsProviderConnected("sqs") {
		fmt.Println("\nPublishing to AWS SQS:")

		// Create a message
		message := CreateMessage("notification.send", "notification-service", "email-service", "notifications", map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Welcome!",
			"body":    "Welcome to our service!",
		})

		// Set priority
		message.SetPriority(5)

		// Publish message
		request := &PublishRequest{
			Topic:   "notifications",
			Message: message,
		}

		response, err := messagingManager.PublishMessage(ctx, "sqs", request)
		if err != nil {
			log.Printf("Failed to publish to SQS: %v", err)
		} else {
			fmt.Printf("Message published to SQS: %s\n", response.MessageID)
		}
	}
}

// exampleSubscribeToTopics demonstrates topic subscription
func exampleSubscribeToTopics(messagingManager *MessagingManager) {
	fmt.Println("\n=== Subscribe to Topics Example ===")

	ctx := context.Background()

	// Example with Kafka
	if messagingManager.IsProviderConnected("kafka") {
		fmt.Println("Subscribing to Kafka topic:")

		// Create message handler
		handler := func(ctx context.Context, message *Message) error {
			fmt.Printf("Received Kafka message: %s from %s\n", message.Type, message.Source)
			fmt.Printf("Payload: %+v\n", message.Payload)
			return nil
		}

		// Subscribe to topic
		request := &SubscribeRequest{
			Topic:         "user-events",
			GroupID:       "example-group",
			AutoAck:       true,
			PrefetchCount: 10,
		}

		err := messagingManager.SubscribeToTopic(ctx, "kafka", request, handler)
		if err != nil {
			log.Printf("Failed to subscribe to Kafka topic: %v", err)
		} else {
			fmt.Println("Subscribed to Kafka topic successfully")
		}
	}

	// Example with RabbitMQ
	if messagingManager.IsProviderConnected("rabbitmq") {
		fmt.Println("\nSubscribing to RabbitMQ topic:")

		// Create message handler
		handler := func(ctx context.Context, message *Message) error {
			fmt.Printf("Received RabbitMQ message: %s from %s\n", message.Type, message.Source)
			fmt.Printf("Payload: %+v\n", message.Payload)
			return nil
		}

		// Subscribe to topic
		request := &SubscribeRequest{
			Topic:   "order-events",
			AutoAck: false,
			Filter: map[string]interface{}{
				"routing_key": "order.processed",
			},
		}

		err := messagingManager.SubscribeToTopic(ctx, "rabbitmq", request, handler)
		if err != nil {
			log.Printf("Failed to subscribe to RabbitMQ topic: %v", err)
		} else {
			fmt.Println("Subscribed to RabbitMQ topic successfully")
		}
	}

	// Example with AWS SQS
	if messagingManager.IsProviderConnected("sqs") {
		fmt.Println("\nSubscribing to AWS SQS topic:")

		// Create message handler
		handler := func(ctx context.Context, message *Message) error {
			fmt.Printf("Received SQS message: %s from %s\n", message.Type, message.Source)
			fmt.Printf("Payload: %+v\n", message.Payload)
			return nil
		}

		// Subscribe to topic
		request := &SubscribeRequest{
			Topic:   "notifications",
			AutoAck: true,
		}

		err := messagingManager.SubscribeToTopic(ctx, "sqs", request, handler)
		if err != nil {
			log.Printf("Failed to subscribe to SQS topic: %v", err)
		} else {
			fmt.Println("Subscribed to SQS topic successfully")
		}
	}
}

// exampleTopicManagement demonstrates topic management
func exampleTopicManagement(messagingManager *MessagingManager) {
	fmt.Println("\n=== Topic Management Example ===")

	ctx := context.Background()

	// Example with Kafka
	if messagingManager.IsProviderConnected("kafka") {
		fmt.Println("Kafka Topic Management:")

		// Create topic
		createRequest := &CreateTopicRequest{
			Topic:             "test-topic",
			Partitions:        3,
			ReplicationFactor: 1,
		}

		err := messagingManager.CreateTopic(ctx, "kafka", createRequest)
		if err != nil {
			log.Printf("Failed to create Kafka topic: %v", err)
		} else {
			fmt.Println("Kafka topic created successfully")
		}

		// Check if topic exists
		existsRequest := &TopicExistsRequest{
			Topic: "test-topic",
		}

		exists, err := messagingManager.TopicExists(ctx, "kafka", existsRequest)
		if err != nil {
			log.Printf("Failed to check Kafka topic existence: %v", err)
		} else {
			fmt.Printf("Kafka topic exists: %t\n", exists)
		}

		// Get topic info
		infoRequest := &GetTopicInfoRequest{
			Topic: "test-topic",
		}

		info, err := messagingManager.GetTopicInfo(ctx, "kafka", infoRequest)
		if err != nil {
			log.Printf("Failed to get Kafka topic info: %v", err)
		} else {
			fmt.Printf("Kafka topic info: %+v\n", info)
		}

		// List topics
		topics, err := messagingManager.ListTopics(ctx, "kafka")
		if err != nil {
			log.Printf("Failed to list Kafka topics: %v", err)
		} else {
			fmt.Printf("Kafka topics: %d found\n", len(topics))
		}
	}

	// Example with RabbitMQ
	if messagingManager.IsProviderConnected("rabbitmq") {
		fmt.Println("\nRabbitMQ Topic Management:")

		// Create topic (queue)
		createRequest := &CreateTopicRequest{
			Topic: "test-queue",
		}

		err := messagingManager.CreateTopic(ctx, "rabbitmq", createRequest)
		if err != nil {
			log.Printf("Failed to create RabbitMQ topic: %v", err)
		} else {
			fmt.Println("RabbitMQ topic created successfully")
		}

		// Check if topic exists
		existsRequest := &TopicExistsRequest{
			Topic: "test-queue",
		}

		exists, err := messagingManager.TopicExists(ctx, "rabbitmq", existsRequest)
		if err != nil {
			log.Printf("Failed to check RabbitMQ topic existence: %v", err)
		} else {
			fmt.Printf("RabbitMQ topic exists: %t\n", exists)
		}
	}

	// Example with AWS SQS
	if messagingManager.IsProviderConnected("sqs") {
		fmt.Println("\nAWS SQS Topic Management:")

		// Create topic (queue)
		createRequest := &CreateTopicRequest{
			Topic: "test-queue",
		}

		err := messagingManager.CreateTopic(ctx, "sqs", createRequest)
		if err != nil {
			log.Printf("Failed to create SQS topic: %v", err)
		} else {
			fmt.Println("SQS topic created successfully")
		}

		// Check if topic exists
		existsRequest := &TopicExistsRequest{
			Topic: "test-queue",
		}

		exists, err := messagingManager.TopicExists(ctx, "sqs", existsRequest)
		if err != nil {
			log.Printf("Failed to check SQS topic existence: %v", err)
		} else {
			fmt.Printf("SQS topic exists: %t\n", exists)
		}

		// List topics
		topics, err := messagingManager.ListTopics(ctx, "sqs")
		if err != nil {
			log.Printf("Failed to list SQS topics: %v", err)
		} else {
			fmt.Printf("SQS topics: %d found\n", len(topics))
		}
	}
}

// exampleBatchOperations demonstrates batch operations
func exampleBatchOperations(messagingManager *MessagingManager) {
	fmt.Println("\n=== Batch Operations Example ===")

	ctx := context.Background()

	// Example with Kafka
	if messagingManager.IsProviderConnected("kafka") {
		fmt.Println("Kafka Batch Operations:")

		// Create multiple messages
		messages := make([]*Message, 3)
		for i := 0; i < 3; i++ {
			messages[i] = CreateMessage("batch.test", "batch-service", "consumer-service", "batch-events", map[string]interface{}{
				"batch_id": fmt.Sprintf("batch-%d", i+1),
				"data":     fmt.Sprintf("message-%d", i+1),
			})
		}

		// Publish batch
		batchRequest := &PublishBatchRequest{
			Topic:    "batch-events",
			Messages: messages,
		}

		response, err := messagingManager.PublishBatch(ctx, "kafka", batchRequest)
		if err != nil {
			log.Printf("Failed to publish batch to Kafka: %v", err)
		} else {
			fmt.Printf("Kafka batch published: %d successful, %d failed\n", response.PublishedCount, response.FailedCount)
		}
	}

	// Example with RabbitMQ
	if messagingManager.IsProviderConnected("rabbitmq") {
		fmt.Println("\nRabbitMQ Batch Operations:")

		// Create multiple messages
		messages := make([]*Message, 3)
		for i := 0; i < 3; i++ {
			messages[i] = CreateMessage("batch.test", "batch-service", "consumer-service", "batch-events", map[string]interface{}{
				"batch_id": fmt.Sprintf("batch-%d", i+1),
				"data":     fmt.Sprintf("message-%d", i+1),
			})
		}

		// Publish batch
		batchRequest := &PublishBatchRequest{
			Topic:      "batch-events",
			Messages:   messages,
			RoutingKey: "batch.test",
		}

		response, err := messagingManager.PublishBatch(ctx, "rabbitmq", batchRequest)
		if err != nil {
			log.Printf("Failed to publish batch to RabbitMQ: %v", err)
		} else {
			fmt.Printf("RabbitMQ batch published: %d successful, %d failed\n", response.PublishedCount, response.FailedCount)
		}
	}

	// Example with AWS SQS
	if messagingManager.IsProviderConnected("sqs") {
		fmt.Println("\nAWS SQS Batch Operations:")

		// Create multiple messages
		messages := make([]*Message, 3)
		for i := 0; i < 3; i++ {
			messages[i] = CreateMessage("batch.test", "batch-service", "consumer-service", "batch-events", map[string]interface{}{
				"batch_id": fmt.Sprintf("batch-%d", i+1),
				"data":     fmt.Sprintf("message-%d", i+1),
			})
		}

		// Publish batch
		batchRequest := &PublishBatchRequest{
			Topic:    "batch-events",
			Messages: messages,
		}

		response, err := messagingManager.PublishBatch(ctx, "sqs", batchRequest)
		if err != nil {
			log.Printf("Failed to publish batch to SQS: %v", err)
		} else {
			fmt.Printf("SQS batch published: %d successful, %d failed\n", response.PublishedCount, response.FailedCount)
		}
	}
}

// exampleHealthChecks demonstrates health checks
func exampleHealthChecks(messagingManager *MessagingManager) {
	fmt.Println("\n=== Health Checks Example ===")

	ctx := context.Background()

	// Perform health checks on all providers
	results := messagingManager.HealthCheck(ctx)

	fmt.Printf("Health Check Results:\n")
	for provider, err := range results {
		if err != nil {
			fmt.Printf("  %s: ❌ %v\n", provider, err)
		} else {
			fmt.Printf("  %s: ✅ Healthy\n", provider)
		}
	}

	// Get connected providers
	connected := messagingManager.GetConnectedProviders()
	fmt.Printf("\nConnected Providers: %v\n", connected)
}

// exampleMessagingStats demonstrates messaging statistics
func exampleMessagingStats(messagingManager *MessagingManager) {
	fmt.Println("\n=== Messaging Statistics Example ===")

	ctx := context.Background()

	// Get statistics for each provider
	providers := messagingManager.GetSupportedProviders()
	for _, providerName := range providers {
		if messagingManager.IsProviderConnected(providerName) {
			stats, err := messagingManager.GetStats(ctx, providerName)
			if err != nil {
				log.Printf("Failed to get stats for %s: %v", providerName, err)
				continue
			}

			fmt.Printf("%s Statistics:\n", providerName)
			fmt.Printf("  Published Messages: %d\n", stats.PublishedMessages)
			fmt.Printf("  Consumed Messages: %d\n", stats.ConsumedMessages)
			fmt.Printf("  Failed Messages: %d\n", stats.FailedMessages)
			fmt.Printf("  Active Connections: %d\n", stats.ActiveConnections)
			fmt.Printf("  Active Subscriptions: %d\n", stats.ActiveSubscriptions)
		}
	}
}

// ExampleProviderCapabilities demonstrates provider capabilities
func ExampleProviderCapabilities() {
	fmt.Println("\n=== Provider Capabilities Example ===")

	// Create logger and messaging manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	messagingManager := NewMessagingManager(config, logger)

	// Register providers
	registerProviders(messagingManager, logger)

	// Get capabilities for each provider
	providers := messagingManager.GetSupportedProviders()
	for _, providerName := range providers {
		features, connInfo, err := messagingManager.GetProviderCapabilities(providerName)
		if err != nil {
			log.Printf("Failed to get capabilities for %s: %v", providerName, err)
			continue
		}

		fmt.Printf("%s Capabilities:\n", providerName)
		fmt.Printf("  Features: %v\n", features)
		fmt.Printf("  Connection Info: %+v\n", connInfo)
	}
}

// ExampleConnectionManagement demonstrates connection management
func ExampleConnectionManagement() {
	fmt.Println("\n=== Connection Management Example ===")

	// Create logger and messaging manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	messagingManager := NewMessagingManager(config, logger)

	// Register providers
	registerProviders(messagingManager, logger)

	ctx := context.Background()

	// Connect to multiple providers
	providers := []string{"kafka", "rabbitmq", "sqs"}
	for _, providerName := range providers {
		if err := messagingManager.Connect(ctx, providerName); err != nil {
			log.Printf("Failed to connect to %s: %v", providerName, err)
		} else {
			fmt.Printf("Connected to %s\n", providerName)
		}
	}

	// Check connection status
	fmt.Printf("Connected providers: %v\n", messagingManager.GetConnectedProviders())

	// Disconnect from providers
	for _, providerName := range providers {
		if err := messagingManager.Disconnect(ctx, providerName); err != nil {
			log.Printf("Failed to disconnect from %s: %v", providerName, err)
		} else {
			fmt.Printf("Disconnected from %s\n", providerName)
		}
	}

	// Close all connections
	if err := messagingManager.Close(); err != nil {
		log.Printf("Failed to close messaging manager: %v", err)
	} else {
		fmt.Println("Messaging manager closed successfully")
	}
}

// ExampleAdvancedMessaging demonstrates advanced messaging features
func ExampleAdvancedMessaging() {
	fmt.Println("\n=== Advanced Messaging Example ===")

	// Create logger and messaging manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	messagingManager := NewMessagingManager(config, logger)

	// Register providers
	registerProviders(messagingManager, logger)

	ctx := context.Background()

	// Connect to Kafka
	if err := messagingManager.Connect(ctx, "kafka"); err != nil {
		log.Printf("Failed to connect to Kafka: %v", err)
		return
	}

	// Example 1: Message with headers and metadata
	message := CreateMessage("advanced.test", "advanced-service", "consumer-service", "advanced-events", map[string]interface{}{
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
	request := &PublishRequest{
		Topic:   "advanced-events",
		Message: message,
		Headers: map[string]interface{}{
			"custom-header": "custom-value",
		},
	}

	response, err := messagingManager.PublishMessage(ctx, "kafka", request)
	if err != nil {
		log.Printf("Failed to publish advanced message: %v", err)
	} else {
		fmt.Printf("Advanced message published: %s\n", response.MessageID)
	}

	// Example 2: Scheduled message
	scheduledMessage := CreateMessage("scheduled.test", "scheduler-service", "consumer-service", "scheduled-events", map[string]interface{}{
		"data": "scheduled message",
	})

	// Schedule message for 1 minute from now
	scheduledMessage.SetScheduledTime(time.Now().Add(1 * time.Minute))

	scheduledRequest := &PublishRequest{
		Topic:   "scheduled-events",
		Message: scheduledMessage,
	}

	response, err = messagingManager.PublishMessage(ctx, "kafka", scheduledRequest)
	if err != nil {
		log.Printf("Failed to publish scheduled message: %v", err)
	} else {
		fmt.Printf("Scheduled message published: %s\n", response.MessageID)
	}

	// Example 3: Message with TTL
	ttlMessage := CreateMessage("ttl.test", "ttl-service", "consumer-service", "ttl-events", map[string]interface{}{
		"data": "ttl message",
	})

	// Set TTL
	ttlMessage.SetTTL(30 * time.Minute)

	ttlRequest := &PublishRequest{
		Topic:   "ttl-events",
		Message: ttlMessage,
	}

	response, err = messagingManager.PublishMessage(ctx, "kafka", ttlRequest)
	if err != nil {
		log.Printf("Failed to publish TTL message: %v", err)
	} else {
		fmt.Printf("TTL message published: %s\n", response.MessageID)
	}
}
