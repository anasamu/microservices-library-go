package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/messaging/gateway"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Provider implements MessagingProvider for AWS SQS
type Provider struct {
	client *sqs.Client
	config map[string]interface{}
	logger *logrus.Logger
	queues map[string]string // topic -> queue URL mapping
}

// NewProvider creates a new AWS SQS messaging provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
		queues: make(map[string]string),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "sqs"
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
	region, _ := p.config["region"].(string)
	if region == "" {
		region = "us-east-1"
	}

	return &gateway.ConnectionInfo{
		Host:     "sqs." + region + ".amazonaws.com",
		Port:     443,
		Protocol: "https",
		Version:  "2012-11-05",
	}
}

// Configure configures the AWS SQS provider
func (p *Provider) Configure(config map[string]interface{}) error {
	region, ok := config["region"].(string)
	if !ok || region == "" {
		region = "us-east-1"
	}

	accessKey, _ := config["access_key"].(string)
	secretKey, _ := config["secret_key"].(string)

	// Create AWS config
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Set credentials if provided
	if accessKey != "" && secretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	}

	// Create SQS client
	p.client = sqs.NewFromConfig(awsConfig)
	p.config = config

	p.logger.Info("AWS SQS provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// Connect connects to AWS SQS
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Test connection by listing queues
	_, err := p.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return fmt.Errorf("failed to connect to SQS: %w", err)
	}

	p.logger.Info("AWS SQS connected successfully")
	return nil
}

// Disconnect disconnects from AWS SQS
func (p *Provider) Disconnect(ctx context.Context) error {
	// SQS doesn't require explicit disconnection
	p.logger.Info("AWS SQS disconnected successfully")
	return nil
}

// Ping checks AWS SQS connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Test connection by listing queues
	_, err := p.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return fmt.Errorf("failed to ping SQS: %w", err)
	}

	return nil
}

// IsConnected checks if AWS SQS is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.Ping(ctx) == nil
}

// PublishMessage publishes a message to AWS SQS
func (p *Provider) PublishMessage(ctx context.Context, request *gateway.PublishRequest) (*gateway.PublishResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("sqs provider not configured")
	}

	// Get or create queue URL
	queueURL, err := p.getOrCreateQueueURL(ctx, request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	// Marshal message
	body, err := json.Marshal(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create message attributes
	messageAttributes := make(map[string]types.MessageAttributeValue)
	messageAttributes["message-type"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(request.Message.Type),
	}
	messageAttributes["source"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(request.Message.Source),
	}
	messageAttributes["target"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(request.Message.Target),
	}

	// Add custom headers as message attributes
	for key, value := range request.Headers {
		if str, ok := value.(string); ok {
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(str),
			}
		}
	}

	// Add message headers as message attributes
	for key, value := range request.Message.Headers {
		if str, ok := value.(string); ok {
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(str),
			}
		}
	}

	// Create send message input
	sendInput := &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueURL),
		MessageBody:       aws.String(string(body)),
		MessageAttributes: messageAttributes,
	}

	// Add delay seconds if message is scheduled
	if request.Message.ScheduledAt != nil {
		delay := int32(time.Until(*request.Message.ScheduledAt).Seconds())
		if delay > 0 && delay <= 900 { // SQS max delay is 15 minutes
			sendInput.DelaySeconds = delay
		}
	}

	// Add message group ID for FIFO queues
	if request.Message.CorrelationID != "" {
		sendInput.MessageGroupId = aws.String(request.Message.CorrelationID)
	}

	// Add deduplication ID for FIFO queues
	if request.Message.ID != uuid.Nil {
		sendInput.MessageDeduplicationId = aws.String(request.Message.ID.String())
	}

	// Send message
	result, err := p.client.SendMessage(ctx, sendInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	response := &gateway.PublishResponse{
		MessageID: *result.MessageId,
		Topic:     request.Topic,
		Timestamp: time.Now(),
		ProviderData: map[string]interface{}{
			"queue_url": queueURL,
		},
	}

	return response, nil
}

// SubscribeToTopic subscribes to an AWS SQS topic
func (p *Provider) SubscribeToTopic(ctx context.Context, request *gateway.SubscribeRequest, handler gateway.MessageHandler) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Get or create queue URL
	queueURL, err := p.getOrCreateQueueURL(ctx, request.Topic)
	if err != nil {
		return fmt.Errorf("failed to get queue URL: %w", err)
	}

	p.logger.WithField("topic", request.Topic).Info("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.WithField("topic", request.Topic).Info("Stopped consuming messages")
				return
			default:
				p.handleMessages(ctx, queueURL, request, handler)
			}
		}
	}()

	return nil
}

// UnsubscribeFromTopic unsubscribes from an AWS SQS topic
func (p *Provider) UnsubscribeFromTopic(ctx context.Context, request *gateway.UnsubscribeRequest) error {
	// SQS doesn't require explicit unsubscription
	p.logger.WithField("topic", request.Topic).Info("Unsubscribed from topic")
	return nil
}

// CreateTopic creates an AWS SQS topic (queue)
func (p *Provider) CreateTopic(ctx context.Context, request *gateway.CreateTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Create queue
	createInput := &sqs.CreateQueueInput{
		QueueName: aws.String(request.Topic),
	}

	// Add attributes for FIFO queue if topic ends with .fifo
	if len(request.Topic) > 5 && request.Topic[len(request.Topic)-5:] == ".fifo" {
		createInput.Attributes = map[string]string{
			"FifoQueue":                 "true",
			"ContentBasedDeduplication": "true",
		}
	}

	// Add retention period if specified
	if request.RetentionPeriod != nil {
		createInput.Attributes["MessageRetentionPeriod"] = fmt.Sprintf("%d", int(request.RetentionPeriod.Seconds()))
	}

	result, err := p.client.CreateQueue(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create queue: %w", err)
	}

	// Store queue URL
	p.queues[request.Topic] = *result.QueueUrl

	p.logger.WithField("topic", request.Topic).Info("Topic created successfully")
	return nil
}

// DeleteTopic deletes an AWS SQS topic (queue)
func (p *Provider) DeleteTopic(ctx context.Context, request *gateway.DeleteTopicRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Get queue URL
	queueURL, err := p.getQueueURL(ctx, request.Topic)
	if err != nil {
		return fmt.Errorf("failed to get queue URL: %w", err)
	}

	// Delete queue
	_, err = p.client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	// Remove from cache
	delete(p.queues, request.Topic)

	p.logger.WithField("topic", request.Topic).Info("Topic deleted successfully")
	return nil
}

// TopicExists checks if an AWS SQS topic (queue) exists
func (p *Provider) TopicExists(ctx context.Context, request *gateway.TopicExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("sqs provider not configured")
	}

	_, err := p.getQueueURL(ctx, request.Topic)
	if err != nil {
		return false, nil // Queue doesn't exist
	}

	return true, nil
}

// ListTopics lists AWS SQS topics (queues)
func (p *Provider) ListTopics(ctx context.Context) ([]gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("sqs provider not configured")
	}

	// List queues
	result, err := p.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	topics := make([]gateway.TopicInfo, 0, len(result.QueueUrls))
	for _, queueURL := range result.QueueUrls {
		// Extract queue name from URL
		queueName := p.extractQueueNameFromURL(queueURL)
		topics = append(topics, gateway.TopicInfo{
			Name: queueName,
			ProviderData: map[string]interface{}{
				"queue_url": queueURL,
			},
		})
	}

	return topics, nil
}

// PublishBatch publishes multiple messages to AWS SQS
func (p *Provider) PublishBatch(ctx context.Context, request *gateway.PublishBatchRequest) (*gateway.PublishBatchResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("sqs provider not configured")
	}

	// Get or create queue URL
	queueURL, err := p.getOrCreateQueueURL(ctx, request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	// Prepare batch entries
	entries := make([]types.SendMessageBatchRequestEntry, len(request.Messages))
	for i, msg := range request.Messages {
		body, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		entries[i] = types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("msg-%d", i)),
			MessageBody: aws.String(string(body)),
		}

		// Add message attributes
		messageAttributes := make(map[string]types.MessageAttributeValue)
		messageAttributes["message-type"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(msg.Type),
		}
		entries[i].MessageAttributes = messageAttributes
	}

	// Send batch
	result, err := p.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueURL),
		Entries:  entries,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message batch: %w", err)
	}

	response := &gateway.PublishBatchResponse{
		PublishedCount: len(result.Successful),
		FailedCount:    len(result.Failed),
		ProviderData: map[string]interface{}{
			"queue_url": queueURL,
		},
	}

	return response, nil
}

// GetTopicInfo gets AWS SQS topic information
func (p *Provider) GetTopicInfo(ctx context.Context, request *gateway.GetTopicInfoRequest) (*gateway.TopicInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("sqs provider not configured")
	}

	// Get queue URL
	queueURL, err := p.getQueueURL(ctx, request.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	// Get queue attributes
	result, err := p.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	info := &gateway.TopicInfo{
		Name: request.Topic,
		ProviderData: map[string]interface{}{
			"queue_url":  queueURL,
			"attributes": result.Attributes,
		},
	}

	return info, nil
}

// GetStats returns AWS SQS statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.MessagingStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("sqs provider not configured")
	}

	stats := &gateway.MessagingStats{
		ActiveConnections:   1, // SQS doesn't have persistent connections
		ActiveSubscriptions: len(p.queues),
		ProviderData: map[string]interface{}{
			"queues": len(p.queues),
		},
	}

	return stats, nil
}

// HealthCheck performs a health check on AWS SQS
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sqs provider not configured")
	}

	// Test connection by listing queues
	_, err := p.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return fmt.Errorf("sqs health check failed: %w", err)
	}

	return nil
}

// Close closes the AWS SQS provider
func (p *Provider) Close() error {
	// SQS doesn't require explicit closing
	p.logger.Info("AWS SQS provider closed")
	return nil
}

// getOrCreateQueueURL gets or creates a queue URL
func (p *Provider) getOrCreateQueueURL(ctx context.Context, topic string) (string, error) {
	// Check cache first
	if queueURL, exists := p.queues[topic]; exists {
		return queueURL, nil
	}

	// Try to get existing queue
	queueURL, err := p.getQueueURL(ctx, topic)
	if err == nil {
		p.queues[topic] = queueURL
		return queueURL, nil
	}

	// Create new queue
	createInput := &sqs.CreateQueueInput{
		QueueName: aws.String(topic),
	}

	// Add attributes for FIFO queue if topic ends with .fifo
	if len(topic) > 5 && topic[len(topic)-5:] == ".fifo" {
		createInput.Attributes = map[string]string{
			"FifoQueue":                 "true",
			"ContentBasedDeduplication": "true",
		}
	}

	result, err := p.client.CreateQueue(ctx, createInput)
	if err != nil {
		return "", fmt.Errorf("failed to create queue: %w", err)
	}

	queueURL = *result.QueueUrl
	p.queues[topic] = queueURL

	return queueURL, nil
}

// getQueueURL gets a queue URL
func (p *Provider) getQueueURL(ctx context.Context, topic string) (string, error) {
	result, err := p.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(topic),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL: %w", err)
	}

	return *result.QueueUrl, nil
}

// extractQueueNameFromURL extracts queue name from SQS URL
func (p *Provider) extractQueueNameFromURL(queueURL string) string {
	// SQS URL format: https://sqs.region.amazonaws.com/account-id/queue-name
	// Extract the last part after the last slash
	lastSlash := -1
	for i := len(queueURL) - 1; i >= 0; i-- {
		if queueURL[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash >= 0 && lastSlash < len(queueURL)-1 {
		return queueURL[lastSlash+1:]
	}

	return queueURL
}

// handleMessages handles messages from SQS
func (p *Provider) handleMessages(ctx context.Context, queueURL string, request *gateway.SubscribeRequest, handler gateway.MessageHandler) {
	// Receive messages
	result, err := p.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueURL),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       20, // Long polling
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		p.logger.WithError(err).Error("Failed to receive messages")
		return
	}

	// Process messages
	for _, message := range result.Messages {
		p.handleMessage(ctx, message, handler)
	}
}

// handleMessage handles a single SQS message
func (p *Provider) handleMessage(ctx context.Context, message types.Message, handler gateway.MessageHandler) {
	var gatewayMessage gateway.Message
	if err := json.Unmarshal([]byte(*message.Body), &gatewayMessage); err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"message_id": *message.MessageId,
		}).Error("Failed to unmarshal message")
		return
	}

	// Set SQS-specific fields
	gatewayMessage.ProviderData = map[string]interface{}{
		"receipt_handle": *message.ReceiptHandle,
		"message_id":     *message.MessageId,
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
		}).Error("Failed to handle message")
		return
	}

	p.logger.WithFields(logrus.Fields{
		"message_id":   gatewayMessage.ID,
		"message_type": gatewayMessage.Type,
		"source":       gatewayMessage.Source,
		"target":       gatewayMessage.Target,
	}).Debug("Message handled successfully")
}
