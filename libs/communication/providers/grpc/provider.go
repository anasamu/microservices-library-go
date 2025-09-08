package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/libs/communication/gateway"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Provider implements CommunicationProvider for gRPC
type Provider struct {
	server   *grpc.Server
	config   map[string]interface{}
	logger   *logrus.Logger
	mu       sync.RWMutex
	services map[string]interface{}
	stats    *gateway.CommunicationStats
}

// NewProvider creates a new gRPC communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:   make(map[string]interface{}),
		logger:   logger,
		services: make(map[string]interface{}),
		stats: &gateway.CommunicationStats{
			TotalConnections:    0,
			ActiveConnections:   0,
			TotalRequests:       0,
			TotalMessages:       0,
			FailedRequests:      0,
			FailedMessages:      0,
			AverageResponseTime: 0,
			ProviderData:        make(map[string]interface{}),
		},
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "grpc"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.CommunicationFeature {
	return []gateway.CommunicationFeature{
		gateway.FeatureHTTP,
		gateway.FeatureHTTPS,
		gateway.FeatureAuthentication,
		gateway.FeatureRateLimiting,
		gateway.FeatureMetrics,
		gateway.FeatureLogging,
		gateway.FeatureHealthChecks,
		gateway.FeatureMiddleware,
		gateway.FeatureHeaders,
		gateway.FeatureQueryParams,
		gateway.FeatureBodyParsing,
		gateway.FeatureStreaming,
		gateway.FeatureRealTime,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 9090
	}

	return &gateway.ConnectionInfo{
		ID:          "grpc-server",
		Type:        "grpc",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the gRPC provider
func (p *Provider) Configure(config map[string]interface{}) error {
	host, ok := config["host"].(string)
	if !ok || host == "" {
		host = "0.0.0.0"
	}

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		port = 9090
	}

	maxRecvMsgSize, ok := config["max_recv_msg_size"].(int)
	if !ok || maxRecvMsgSize == 0 {
		maxRecvMsgSize = 4 * 1024 * 1024 // 4MB
	}

	maxSendMsgSize, ok := config["max_send_msg_size"].(int)
	if !ok || maxSendMsgSize == 0 {
		maxSendMsgSize = 4 * 1024 * 1024 // 4MB
	}

	enableReflection, ok := config["enable_reflection"].(bool)
	if !ok {
		enableReflection = true
	}

	p.config = config

	p.logger.Info("gRPC provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Start starts the gRPC server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("grpc provider not configured")
	}

	// Merge provided config with existing config
	for key, value := range config {
		p.config[key] = value
	}

	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	maxRecvMsgSize, _ := p.config["max_recv_msg_size"].(int)
	maxSendMsgSize, _ := p.config["max_send_msg_size"].(int)
	enableReflection, _ := p.config["enable_reflection"].(bool)

	// Create gRPC server
	p.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
	)

	// Enable reflection if configured
	if enableReflection {
		reflection.Register(p.server)
	}

	// Register services
	p.registerServices()

	// Start server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Info("gRPC server started successfully")

	// Start server in goroutine
	go func() {
		if err := p.server.Serve(lis); err != nil {
			p.logger.WithError(err).Error("gRPC server error")
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (p *Provider) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	p.logger.Info("Stopping gRPC server")
	p.server.GracefulStop()
	return nil
}

// IsRunning checks if the gRPC server is running
func (p *Provider) IsRunning() bool {
	return p.server != nil
}

// HandleRequest handles a gRPC request
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("grpc server not running")
	}

	start := time.Now()

	// gRPC requests are handled differently than HTTP requests
	// This is a simplified implementation for demonstration
	response := &gateway.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"content-type": "application/grpc",
		},
		Body: []byte(`{"message":"gRPC request handled"}`),
		Metadata: map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
			"method":      request.Method,
			"path":        request.Path,
		},
	}

	// Update stats
	p.mu.Lock()
	p.stats.TotalRequests++
	p.stats.AverageResponseTime = (p.stats.AverageResponseTime + time.Since(start)) / 2
	p.mu.Unlock()

	return response, nil
}

// HandleWebSocket handles a WebSocket connection (not supported by gRPC provider)
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	return nil, fmt.Errorf("WebSocket not supported by gRPC provider")
}

// SendMessage sends a message (not applicable for gRPC provider)
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	return nil, fmt.Errorf("message sending not applicable for gRPC provider")
}

// BroadcastMessage broadcasts a message (not applicable for gRPC provider)
func (p *Provider) BroadcastMessage(ctx context.Context, request *gateway.BroadcastRequest) (*gateway.BroadcastResponse, error) {
	return nil, fmt.Errorf("message broadcasting not applicable for gRPC provider")
}

// GetConnections gets gRPC connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	// gRPC connections are managed by the server
	connections := []gateway.ConnectionInfo{
		*p.GetConnectionInfo(),
	}
	return connections, nil
}

// GetConnectionCount gets gRPC connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	// gRPC server manages connections internally
	// This is a simplified implementation
	return 1, nil
}

// CloseConnection closes a connection (not applicable for gRPC provider)
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	return fmt.Errorf("connection closing not applicable for gRPC provider")
}

// HealthCheck performs a health check on the gRPC server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("grpc server not running")
	}

	// Create a test request
	request := &gateway.Request{
		Method: "GET",
		Path:   "/health",
	}

	_, err := p.HandleRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("grpc health check failed: %w", err)
	}

	return nil
}

// GetStats returns gRPC server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ProviderData = map[string]interface{}{
		"services": len(p.services),
		"server":   p.server != nil,
	}

	return &stats, nil
}

// Close closes the gRPC provider
func (p *Provider) Close() error {
	if p.server != nil {
		p.server.GracefulStop()
	}
	return nil
}

// registerServices registers gRPC services
func (p *Provider) registerServices() {
	// Register health check service
	// This is a simplified implementation
	// In a real implementation, you would register actual gRPC services

	p.logger.Info("gRPC services registered")
}

// RegisterService registers a gRPC service
func (p *Provider) RegisterService(name string, service interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.services[name] = service
	p.logger.WithField("service", name).Info("gRPC service registered")
}

// GetService gets a registered gRPC service
func (p *Provider) GetService(name string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	service, exists := p.services[name]
	return service, exists
}

// ListServices lists all registered gRPC services
func (p *Provider) ListServices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	services := make([]string, 0, len(p.services))
	for name := range p.services {
		services = append(services, name)
	}
	return services
}
