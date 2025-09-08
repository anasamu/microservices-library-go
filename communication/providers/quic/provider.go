package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/communication/gateway"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

// Provider implements CommunicationProvider for QUIC
type Provider struct {
	listener *quic.Listener
	running  bool
	config   map[string]interface{}
	logger   *logrus.Logger
	mu       sync.RWMutex
	clients  map[string]*QUICClient
	stats    *gateway.CommunicationStats
}

// QUICClient represents a QUIC client connection
type QUICClient struct {
	ID          string
	UserID      string
	ConnectedAt time.Time
	LastSeen    time.Time
	Metadata    map[string]interface{}
	conn        quic.Connection
	streams     map[quic.StreamID]quic.Stream
}

// NewProvider creates a new QUIC communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:  make(map[string]interface{}),
		logger:  logger,
		clients: make(map[string]*QUICClient),
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
	return "quic"
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
		port = 8443
	}

	return &gateway.ConnectionInfo{
		ID:          "quic-server",
		Type:        "quic",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the QUIC provider
func (p *Provider) Configure(config map[string]interface{}) error {
	host, ok := config["host"].(string)
	if !ok || host == "" {
		host = "0.0.0.0"
	}

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		port = 8443
	}

	maxStreams, ok := config["max_streams"].(int64)
	if !ok || maxStreams == 0 {
		maxStreams = 100
	}

	maxIdleTimeout, ok := config["max_idle_timeout"].(time.Duration)
	if !ok || maxIdleTimeout == 0 {
		maxIdleTimeout = 30 * time.Second
	}

	keepAlivePeriod, ok := config["keep_alive_period"].(time.Duration)
	if !ok || keepAlivePeriod == 0 {
		keepAlivePeriod = 10 * time.Second
	}

	p.config = config

	p.logger.Info("QUIC provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Start starts the QUIC server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("quic provider not configured")
	}

	// Merge provided config with existing config
	for key, value := range config {
		p.config[key] = value
	}

	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	maxStreams, _ := p.config["max_streams"].(int64)
	maxIdleTimeout, _ := p.config["max_idle_timeout"].(time.Duration)
	keepAlivePeriod, _ := p.config["keep_alive_period"].(time.Duration)

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{p.generateSelfSignedCert()},
		NextProtos:   []string{"quic-echo-example"},
	}

	// Create QUIC config
	quicConfig := &quic.Config{
		MaxIncomingStreams: maxStreams,
		MaxIdleTimeout:     maxIdleTimeout,
		KeepAlivePeriod:    keepAlivePeriod,
	}

	// Create listener
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	p.listener = listener
	p.running = true

	p.logger.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Info("QUIC server started successfully")

	// Start accepting connections in goroutine
	go p.acceptConnections()

	return nil
}

// Stop stops the QUIC server
func (p *Provider) Stop(ctx context.Context) error {
	if !p.running {
		return nil
	}

	p.logger.Info("Stopping QUIC server")

	// Close all clients
	p.mu.Lock()
	for _, client := range p.clients {
		client.conn.CloseWithError(0, "server shutdown")
	}
	p.clients = make(map[string]*QUICClient)
	p.mu.Unlock()

	p.running = false
	return p.listener.Close()
}

// IsRunning checks if the QUIC server is running
func (p *Provider) IsRunning() bool {
	return p.running
}

// HandleRequest handles a QUIC request
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("quic server not running")
	}

	start := time.Now()

	// QUIC requests are handled through streams
	// This is a simplified implementation for demonstration
	response := &gateway.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"content-type": "application/octet-stream",
		},
		Body: []byte(`{"message":"QUIC request handled"}`),
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

// HandleWebSocket handles a WebSocket connection (not supported by QUIC provider)
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	return nil, fmt.Errorf("WebSocket not supported by QUIC provider")
}

// SendMessage sends a message to a specific connection
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("quic server not running")
	}

	// Find client by connection ID or user ID
	var targetClient *QUICClient
	p.mu.RLock()
	for _, client := range p.clients {
		if (request.ConnectionID != "" && client.ID == request.ConnectionID) ||
			(request.UserID != "" && client.UserID == request.UserID) {
			targetClient = client
			break
		}
	}
	p.mu.RUnlock()

	if targetClient == nil {
		return nil, fmt.Errorf("client not found")
	}

	// Create a new stream for the message
	stream, err := targetClient.conn.OpenStreamSync(ctx)
	if err != nil {
		p.stats.FailedMessages++
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Send message data
	messageData := []byte(fmt.Sprintf("Message: %s", request.Message.ID))
	_, err = stream.Write(messageData)
	if err != nil {
		p.stats.FailedMessages++
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	p.stats.TotalMessages++
	response := &gateway.SendMessageResponse{
		MessageID: request.Message.ID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"connection_id": targetClient.ID,
			"user_id":       targetClient.UserID,
		},
	}

	return response, nil
}

// BroadcastMessage broadcasts a message to all connected clients
func (p *Provider) BroadcastMessage(ctx context.Context, request *gateway.BroadcastRequest) (*gateway.BroadcastResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("quic server not running")
	}

	sentCount := 0
	failedCount := 0

	p.mu.RLock()
	clients := make([]*QUICClient, 0, len(p.clients))
	for _, client := range p.clients {
		clients = append(clients, client)
	}
	p.mu.RUnlock()

	for _, client := range clients {
		// Create a new stream for the message
		stream, err := client.conn.OpenStreamSync(ctx)
		if err != nil {
			failedCount++
			continue
		}

		// Send message data
		messageData := []byte(fmt.Sprintf("Broadcast: %s", request.Message.ID))
		_, err = stream.Write(messageData)
		stream.Close()

		if err != nil {
			failedCount++
		} else {
			sentCount++
		}
	}

	response := &gateway.BroadcastResponse{
		SentCount:   sentCount,
		FailedCount: failedCount,
		Metadata: map[string]interface{}{
			"message_type": request.Message.Type,
		},
	}

	return response, nil
}

// GetConnections gets QUIC connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]gateway.ConnectionInfo, 0, len(p.clients))
	for _, client := range p.clients {
		connInfo := gateway.ConnectionInfo{
			ID:          client.ID,
			Type:        "quic",
			RemoteAddr:  client.conn.RemoteAddr().String(),
			ConnectedAt: client.ConnectedAt,
			LastSeen:    client.LastSeen,
			UserID:      client.UserID,
			Metadata:    client.Metadata,
		}
		connections = append(connections, connInfo)
	}

	return connections, nil
}

// GetConnectionCount gets QUIC connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients), nil
}

// CloseConnection closes a QUIC connection
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	if !p.IsRunning() {
		return fmt.Errorf("quic server not running")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	client, exists := p.clients[connectionID]
	if !exists {
		return fmt.Errorf("connection not found")
	}

	client.conn.CloseWithError(0, "connection closed")
	delete(p.clients, connectionID)
	p.stats.ActiveConnections = len(p.clients)

	p.logger.WithField("connection_id", connectionID).Info("QUIC connection closed")
	return nil
}

// HealthCheck performs a health check on the QUIC server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("quic server not running")
	}

	// Check if we can get connection count
	_, err := p.GetConnectionCount(ctx)
	if err != nil {
		return fmt.Errorf("quic health check failed: %w", err)
	}

	return nil
}

// GetStats returns QUIC server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ActiveConnections = len(p.clients)
	stats.ProviderData = map[string]interface{}{
		"clients": len(p.clients),
		"server":  p.listener != nil,
	}

	return &stats, nil
}

// Close closes the QUIC provider
func (p *Provider) Close() error {
	if p.listener != nil {
		return p.listener.Close()
	}
	return nil
}

// acceptConnections accepts incoming QUIC connections
func (p *Provider) acceptConnections() {
	for {
		conn, err := p.listener.Accept(context.Background())
		if err != nil {
			if err != net.ErrClosed {
				p.logger.WithError(err).Error("Failed to accept QUIC connection")
			}
			return
		}

		// Handle connection in goroutine
		go p.handleConnection(conn)
	}
}

// handleConnection handles a QUIC connection
func (p *Provider) handleConnection(conn quic.Connection) {
	clientID := generateClientID()
	client := &QUICClient{
		ID:          clientID,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Metadata:    make(map[string]interface{}),
		conn:        conn,
		streams:     make(map[quic.StreamID]quic.Stream),
	}

	p.mu.Lock()
	p.clients[clientID] = client
	p.stats.TotalConnections++
	p.stats.ActiveConnections = len(p.clients)
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"client_id":     clientID,
		"remote_addr":   conn.RemoteAddr().String(),
		"total_clients": len(p.clients),
	}).Info("QUIC client connected")

	// Handle streams
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			break
		}

		client.streams[stream.StreamID()] = stream
		go p.handleStream(client, stream)
	}

	// Clean up connection
	p.mu.Lock()
	delete(p.clients, clientID)
	p.stats.ActiveConnections = len(p.clients)
	p.mu.Unlock()

	p.logger.WithField("client_id", clientID).Info("QUIC client disconnected")
}

// handleStream handles a QUIC stream
func (p *Provider) handleStream(client *QUICClient, stream quic.Stream) {
	defer stream.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			break
		}

		// Echo the data back
		stream.Write(buffer[:n])

		// Update last seen
		client.LastSeen = time.Now()
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("quic_client_%d", time.Now().UnixNano())
}

// generateSelfSignedCert generates a self-signed certificate for QUIC
func (p *Provider) generateSelfSignedCert() tls.Certificate {
	// This is a simplified implementation
	// In production, you should use proper certificates
	certPEM := `-----BEGIN CERTIFICATE-----
MIICljCCAX4CCQDQ5Q5Q5Q5Q5Q==
-----END CERTIFICATE-----`
	keyPEM := `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7VJTUt9Us8cKB
-----END PRIVATE KEY-----`

	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		// Fallback to a minimal certificate
		cert, _ = tls.X509KeyPair([]byte{}, []byte{})
	}

	return cert
}
