package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// CommunicationManager manages multiple communication providers
type CommunicationManager struct {
	providers map[string]CommunicationProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds communication manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	MaxConnections  int               `json:"max_connections"`
	Metadata        map[string]string `json:"metadata"`
}

// CommunicationProvider interface for communication backends
type CommunicationProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []CommunicationFeature
	GetConnectionInfo() *ConnectionInfo

	// Server management
	Start(ctx context.Context, config map[string]interface{}) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Request handling
	HandleRequest(ctx context.Context, request *Request) (*Response, error)
	HandleWebSocket(ctx context.Context, request *WebSocketRequest) (*WebSocketResponse, error)

	// Message operations
	SendMessage(ctx context.Context, request *SendMessageRequest) (*SendMessageResponse, error)
	BroadcastMessage(ctx context.Context, request *BroadcastRequest) (*BroadcastResponse, error)

	// Connection management
	GetConnections(ctx context.Context) ([]ConnectionInfo, error)
	GetConnectionCount(ctx context.Context) (int, error)
	CloseConnection(ctx context.Context, connectionID string) error

	// Health and monitoring
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*CommunicationStats, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// CommunicationFeature represents a communication feature
type CommunicationFeature string

const (
	FeatureHTTP             CommunicationFeature = "http"
	FeatureHTTPS            CommunicationFeature = "https"
	FeatureWebSocket        CommunicationFeature = "websocket"
	FeatureServerSentEvents CommunicationFeature = "server_sent_events"
	FeatureLongPolling      CommunicationFeature = "long_polling"
	FeatureCORS             CommunicationFeature = "cors"
	FeatureCompression      CommunicationFeature = "compression"
	FeatureAuthentication   CommunicationFeature = "authentication"
	FeatureRateLimiting     CommunicationFeature = "rate_limiting"
	FeatureLoadBalancing    CommunicationFeature = "load_balancing"
	FeatureMetrics          CommunicationFeature = "metrics"
	FeatureLogging          CommunicationFeature = "logging"
	FeatureHealthChecks     CommunicationFeature = "health_checks"
	FeatureMiddleware       CommunicationFeature = "middleware"
	FeatureStaticFiles      CommunicationFeature = "static_files"
	FeatureTemplates        CommunicationFeature = "templates"
	FeatureSessions         CommunicationFeature = "sessions"
	FeatureCookies          CommunicationFeature = "cookies"
	FeatureHeaders          CommunicationFeature = "headers"
	FeatureQueryParams      CommunicationFeature = "query_params"
	FeaturePathParams       CommunicationFeature = "path_params"
	FeatureBodyParsing      CommunicationFeature = "body_parsing"
	FeatureFileUpload       CommunicationFeature = "file_upload"
	FeatureStreaming        CommunicationFeature = "streaming"
	FeatureRealTime         CommunicationFeature = "real_time"
	FeatureBroadcasting     CommunicationFeature = "broadcasting"
	FeatureMulticasting     CommunicationFeature = "multicasting"
	FeatureUnicasting       CommunicationFeature = "unicasting"
)

// ConnectionInfo represents connection information
type ConnectionInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	RemoteAddr  string                 `json:"remote_addr"`
	UserAgent   string                 `json:"user_agent"`
	ConnectedAt time.Time              `json:"connected_at"`
	LastSeen    time.Time              `json:"last_seen"`
	UserID      string                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Request represents an HTTP request
type Request struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Headers     map[string]string      `json:"headers"`
	QueryParams map[string]string      `json:"query_params"`
	PathParams  map[string]string      `json:"path_params"`
	Body        []byte                 `json:"body"`
	RemoteAddr  string                 `json:"remote_addr"`
	UserAgent   string                 `json:"user_agent"`
	UserID      string                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketRequest represents a WebSocket connection request
type WebSocketRequest struct {
	Path        string                 `json:"path"`
	Headers     map[string]string      `json:"headers"`
	QueryParams map[string]string      `json:"query_params"`
	RemoteAddr  string                 `json:"remote_addr"`
	UserAgent   string                 `json:"user_agent"`
	UserID      string                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketResponse represents a WebSocket connection response
type WebSocketResponse struct {
	ConnectionID string                 `json:"connection_id"`
	Status       int                    `json:"status"`
	Headers      map[string]string      `json:"headers"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageRequest represents a message sending request
type SendMessageRequest struct {
	ConnectionID string                 `json:"connection_id"`
	UserID       string                 `json:"user_id,omitempty"`
	Message      *Message               `json:"message"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageResponse represents a message sending response
type SendMessageResponse struct {
	MessageID string                 `json:"message_id"`
	Status    string                 `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BroadcastRequest represents a broadcast request
type BroadcastRequest struct {
	Message  *Message               `json:"message"`
	Filter   map[string]interface{} `json:"filter,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BroadcastResponse represents a broadcast response
type BroadcastResponse struct {
	SentCount   int                    `json:"sent_count"`
	FailedCount int                    `json:"failed_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a communication message
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   interface{}            `json:"content"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CommunicationStats represents communication statistics
type CommunicationStats struct {
	TotalConnections    int                    `json:"total_connections"`
	ActiveConnections   int                    `json:"active_connections"`
	TotalRequests       int64                  `json:"total_requests"`
	TotalMessages       int64                  `json:"total_messages"`
	FailedRequests      int64                  `json:"failed_requests"`
	FailedMessages      int64                  `json:"failed_messages"`
	AverageResponseTime time.Duration          `json:"average_response_time"`
	ProviderData        map[string]interface{} `json:"provider_data"`
}

// DefaultManagerConfig returns default communication manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "http",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		MaxConnections:  1000,
		Metadata:        make(map[string]string),
	}
}

// NewCommunicationManager creates a new communication manager
func NewCommunicationManager(config *ManagerConfig, logger *logrus.Logger) *CommunicationManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &CommunicationManager{
		providers: make(map[string]CommunicationProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a communication provider
func (cm *CommunicationManager) RegisterProvider(provider CommunicationProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	cm.providers[name] = provider
	cm.logger.WithField("provider", name).Info("Communication provider registered")

	return nil
}

// GetProvider returns a communication provider by name
func (cm *CommunicationManager) GetProvider(name string) (CommunicationProvider, error) {
	provider, exists := cm.providers[name]
	if !exists {
		return nil, fmt.Errorf("communication provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default communication provider
func (cm *CommunicationManager) GetDefaultProvider() (CommunicationProvider, error) {
	return cm.GetProvider(cm.config.DefaultProvider)
}

// Start starts a communication provider
func (cm *CommunicationManager) Start(ctx context.Context, providerName string, config map[string]interface{}) error {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Start with retry logic
	for attempt := 1; attempt <= cm.config.RetryAttempts; attempt++ {
		err = provider.Start(ctx, config)
		if err == nil {
			break
		}

		cm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
		}).Warn("Communication provider start failed, retrying")

		if attempt < cm.config.RetryAttempts {
			time.Sleep(cm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to start communication provider after %d attempts: %w", cm.config.RetryAttempts, err)
	}

	cm.logger.WithField("provider", providerName).Info("Communication provider started successfully")
	return nil
}

// Stop stops a communication provider
func (cm *CommunicationManager) Stop(ctx context.Context, providerName string) error {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Stop(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop communication provider: %w", err)
	}

	cm.logger.WithField("provider", providerName).Info("Communication provider stopped successfully")
	return nil
}

// HandleRequest handles an HTTP request using the specified provider
func (cm *CommunicationManager) HandleRequest(ctx context.Context, providerName string, request *Request) (*Response, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := cm.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	response, err := provider.HandleRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to handle request: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"method":      request.Method,
		"path":        request.Path,
		"status_code": response.StatusCode,
	}).Debug("Request handled successfully")

	return response, nil
}

// HandleWebSocket handles a WebSocket connection using the specified provider
func (cm *CommunicationManager) HandleWebSocket(ctx context.Context, providerName string, request *WebSocketRequest) (*WebSocketResponse, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := cm.validateWebSocketRequest(request); err != nil {
		return nil, fmt.Errorf("invalid WebSocket request: %w", err)
	}

	response, err := provider.HandleWebSocket(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to handle WebSocket request: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"provider":      providerName,
		"path":          request.Path,
		"connection_id": response.ConnectionID,
		"status":        response.Status,
	}).Debug("WebSocket request handled successfully")

	return response, nil
}

// SendMessage sends a message using the specified provider
func (cm *CommunicationManager) SendMessage(ctx context.Context, providerName string, request *SendMessageRequest) (*SendMessageResponse, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := cm.validateSendMessageRequest(request); err != nil {
		return nil, fmt.Errorf("invalid send message request: %w", err)
	}

	response, err := provider.SendMessage(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"message_id": response.MessageID,
		"status":     response.Status,
	}).Debug("Message sent successfully")

	return response, nil
}

// BroadcastMessage broadcasts a message using the specified provider
func (cm *CommunicationManager) BroadcastMessage(ctx context.Context, providerName string, request *BroadcastRequest) (*BroadcastResponse, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := cm.validateBroadcastRequest(request); err != nil {
		return nil, fmt.Errorf("invalid broadcast request: %w", err)
	}

	response, err := provider.BroadcastMessage(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast message: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"provider":     providerName,
		"sent_count":   response.SentCount,
		"failed_count": response.FailedCount,
	}).Debug("Message broadcast successfully")

	return response, nil
}

// GetConnections gets connections from a provider
func (cm *CommunicationManager) GetConnections(ctx context.Context, providerName string) ([]ConnectionInfo, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	connections, err := provider.GetConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	return connections, nil
}

// GetConnectionCount gets connection count from a provider
func (cm *CommunicationManager) GetConnectionCount(ctx context.Context, providerName string) (int, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return 0, err
	}

	count, err := provider.GetConnectionCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection count: %w", err)
	}

	return count, nil
}

// CloseConnection closes a connection using the specified provider
func (cm *CommunicationManager) CloseConnection(ctx context.Context, providerName, connectionID string) error {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.CloseConnection(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"provider":      providerName,
		"connection_id": connectionID,
	}).Debug("Connection closed successfully")

	return nil
}

// HealthCheck performs health check on all providers
func (cm *CommunicationManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range cm.providers {
		err := provider.HealthCheck(ctx)
		results[name] = err
	}

	return results
}

// GetStats gets statistics from a provider
func (cm *CommunicationManager) GetStats(ctx context.Context, providerName string) (*CommunicationStats, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	stats, err := provider.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get communication stats: %w", err)
	}

	return stats, nil
}

// GetSupportedProviders returns a list of registered providers
func (cm *CommunicationManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(cm.providers))
	for name := range cm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (cm *CommunicationManager) GetProviderCapabilities(providerName string) ([]CommunicationFeature, *ConnectionInfo, error) {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	return provider.GetSupportedFeatures(), provider.GetConnectionInfo(), nil
}

// Close closes all communication connections
func (cm *CommunicationManager) Close() error {
	var lastErr error

	for name, provider := range cm.providers {
		if err := provider.Close(); err != nil {
			cm.logger.WithError(err).WithField("provider", name).Error("Failed to close communication provider")
			lastErr = err
		}
	}

	return lastErr
}

// IsProviderRunning checks if a provider is running
func (cm *CommunicationManager) IsProviderRunning(providerName string) bool {
	provider, err := cm.GetProvider(providerName)
	if err != nil {
		return false
	}
	return provider.IsRunning()
}

// GetRunningProviders returns a list of running providers
func (cm *CommunicationManager) GetRunningProviders() []string {
	running := make([]string, 0)
	for name, provider := range cm.providers {
		if provider.IsRunning() {
			running = append(running, name)
		}
	}
	return running
}

// validateRequest validates an HTTP request
func (cm *CommunicationManager) validateRequest(request *Request) error {
	if request.Method == "" {
		return fmt.Errorf("method is required")
	}

	if request.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

// validateWebSocketRequest validates a WebSocket request
func (cm *CommunicationManager) validateWebSocketRequest(request *WebSocketRequest) error {
	if request.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

// validateSendMessageRequest validates a send message request
func (cm *CommunicationManager) validateSendMessageRequest(request *SendMessageRequest) error {
	if request.ConnectionID == "" && request.UserID == "" {
		return fmt.Errorf("connection_id or user_id is required")
	}

	if request.Message == nil {
		return fmt.Errorf("message is required")
	}

	return nil
}

// validateBroadcastRequest validates a broadcast request
func (cm *CommunicationManager) validateBroadcastRequest(request *BroadcastRequest) error {
	if request.Message == nil {
		return fmt.Errorf("message is required")
	}

	return nil
}

// CreateMessage creates a new message with default values
func CreateMessage(messageType string, content interface{}) *Message {
	return &Message{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Type:      messageType,
		Content:   content,
		Timestamp: time.Now(),
		Headers:   make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
}

// SetFrom sets the message sender
func (m *Message) SetFrom(from string) {
	m.From = from
}

// SetTo sets the message recipient
func (m *Message) SetTo(to string) {
	m.To = to
}

// AddHeader adds a header to the message
func (m *Message) AddHeader(key string, value interface{}) {
	if m.Headers == nil {
		m.Headers = make(map[string]interface{})
	}
	m.Headers[key] = value
}

// GetHeader retrieves a header from the message
func (m *Message) GetHeader(key string) (interface{}, bool) {
	if m.Headers == nil {
		return nil, false
	}
	value, exists := m.Headers[key]
	return value, exists
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
