package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/communication"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Provider implements CommunicationProvider for WebSocket
type Provider struct {
	upgrader   websocket.Upgrader
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	config     map[string]interface{}
	logger     *logrus.Logger
	mu         sync.RWMutex
	stats      *gateway.CommunicationStats
}

// Client represents a WebSocket client
type Client struct {
	conn     *websocket.Conn
	request  *http.Request
	send     chan []byte
	server   *Provider
	ID       string
	UserID   string
	Metadata map[string]interface{}
}

// NewProvider creates a new WebSocket communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:     make(map[string]interface{}),
		logger:     logger,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
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
	return "websocket"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.CommunicationFeature {
	return []gateway.CommunicationFeature{
		gateway.FeatureWebSocket,
		gateway.FeatureRealTime,
		gateway.FeatureBroadcasting,
		gateway.FeatureMulticasting,
		gateway.FeatureUnicasting,
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
		port = 8080
	}

	return &gateway.ConnectionInfo{
		ID:          "websocket-server",
		Type:        "websocket",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the WebSocket provider
func (p *Provider) Configure(config map[string]interface{}) error {
	readBufferSize, ok := config["read_buffer_size"].(int)
	if !ok || readBufferSize == 0 {
		readBufferSize = 1024
	}

	writeBufferSize, ok := config["write_buffer_size"].(int)
	if !ok || writeBufferSize == 0 {
		writeBufferSize = 1024
	}

	checkOrigin, ok := config["check_origin"].(bool)
	if !ok {
		checkOrigin = false
	}

	pingPeriod, ok := config["ping_period"].(time.Duration)
	if !ok || pingPeriod == 0 {
		pingPeriod = 54 * time.Second
	}

	pongWait, ok := config["pong_wait"].(time.Duration)
	if !ok || pongWait == 0 {
		pongWait = 60 * time.Second
	}

	writeWait, ok := config["write_wait"].(time.Duration)
	if !ok || writeWait == 0 {
		writeWait = 10 * time.Second
	}

	maxMessageSize, ok := config["max_message_size"].(int64)
	if !ok || maxMessageSize == 0 {
		maxMessageSize = 512
	}

	// Configure upgrader
	p.upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return !checkOrigin
		},
	}

	p.config = config

	p.logger.Info("WebSocket provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Start starts the WebSocket server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("websocket provider not configured")
	}

	// Merge provided config with existing config
	for key, value := range config {
		p.config[key] = value
	}

	// Start the hub
	go p.run()

	p.logger.Info("WebSocket server started successfully")
	return nil
}

// Stop stops the WebSocket server
func (p *Provider) Stop(ctx context.Context) error {
	p.logger.Info("Stopping WebSocket server")

	// Close all clients
	p.mu.Lock()
	for client := range p.clients {
		client.conn.Close()
		close(client.send)
	}
	p.clients = make(map[*Client]bool)
	p.mu.Unlock()

	return nil
}

// IsRunning checks if the WebSocket server is running
func (p *Provider) IsRunning() bool {
	return len(p.config) > 0
}

// HandleRequest handles an HTTP request (not applicable for WebSocket provider)
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	return nil, fmt.Errorf("HTTP requests not supported by WebSocket provider")
}

// HandleWebSocket handles a WebSocket connection
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("websocket server not running")
	}

	// Create a mock HTTP request for the upgrader
	httpReq := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: request.Path},
		Header: make(http.Header),
	}

	// Set headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Create a mock response writer
	responseWriter := &mockResponseWriter{}

	// Upgrade connection
	conn, err := p.upgrader.Upgrade(responseWriter, httpReq, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade connection: %w", err)
	}

	// Create client
	client := &Client{
		conn:     conn,
		request:  httpReq,
		send:     make(chan []byte, 256),
		server:   p,
		ID:       generateClientID(),
		UserID:   request.UserID,
		Metadata: request.Metadata,
	}

	// Register client
	p.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	response := &gateway.WebSocketResponse{
		ConnectionID: client.ID,
		Status:       http.StatusSwitchingProtocols,
		Headers: map[string]string{
			"Upgrade":    "websocket",
			"Connection": "Upgrade",
		},
		Metadata: map[string]interface{}{
			"user_id": client.UserID,
		},
	}

	return response, nil
}

// SendMessage sends a message to a specific connection
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("websocket server not running")
	}

	// Find client by connection ID or user ID
	var targetClient *Client
	p.mu.RLock()
	for client := range p.clients {
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

	// Marshal message
	data, err := json.Marshal(request.Message)
	if err != nil {
		p.stats.FailedMessages++
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message
	select {
	case targetClient.send <- data:
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
	default:
		p.stats.FailedMessages++
		return nil, fmt.Errorf("failed to send message: client channel full")
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (p *Provider) BroadcastMessage(ctx context.Context, request *gateway.BroadcastRequest) (*gateway.BroadcastResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("websocket server not running")
	}

	// Marshal message
	data, err := json.Marshal(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Broadcast message
	p.broadcast <- data

	// Count clients
	p.mu.RLock()
	clientCount := len(p.clients)
	p.mu.RUnlock()

	response := &gateway.BroadcastResponse{
		SentCount:   clientCount,
		FailedCount: 0,
		Metadata: map[string]interface{}{
			"message_type": request.Message.Type,
		},
	}

	return response, nil
}

// GetConnections gets WebSocket connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]gateway.ConnectionInfo, 0, len(p.clients))
	for client := range p.clients {
		connInfo := gateway.ConnectionInfo{
			ID:          client.ID,
			Type:        "websocket",
			RemoteAddr:  client.conn.RemoteAddr().String(),
			UserAgent:   client.request.UserAgent(),
			ConnectedAt: time.Now(), // Simplified
			LastSeen:    time.Now(),
			UserID:      client.UserID,
			Metadata:    client.Metadata,
		}
		connections = append(connections, connInfo)
	}

	return connections, nil
}

// GetConnectionCount gets WebSocket connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients), nil
}

// CloseConnection closes a WebSocket connection
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	if !p.IsRunning() {
		return fmt.Errorf("websocket server not running")
	}

	p.mu.RLock()
	var targetClient *Client
	for client := range p.clients {
		if client.ID == connectionID {
			targetClient = client
			break
		}
	}
	p.mu.RUnlock()

	if targetClient == nil {
		return fmt.Errorf("connection not found")
	}

	// Unregister client
	p.unregister <- targetClient

	return nil
}

// HealthCheck performs a health check on the WebSocket server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("websocket server not running")
	}

	// Check if we can get connection count
	_, err := p.GetConnectionCount(ctx)
	if err != nil {
		return fmt.Errorf("websocket health check failed: %w", err)
	}

	return nil
}

// GetStats returns WebSocket server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ActiveConnections = len(p.clients)
	stats.ProviderData = map[string]interface{}{
		"clients": len(p.clients),
	}

	return &stats, nil
}

// Close closes the WebSocket provider
func (p *Provider) Close() error {
	return p.Stop(context.Background())
}

// run handles client connections and message broadcasting
func (p *Provider) run() {
	for {
		select {
		case client := <-p.register:
			p.mu.Lock()
			p.clients[client] = true
			p.stats.TotalConnections++
			p.stats.ActiveConnections = len(p.clients)
			p.mu.Unlock()

			p.logger.WithFields(logrus.Fields{
				"client_id":     client.ID,
				"user_id":       client.UserID,
				"total_clients": len(p.clients),
			}).Info("Client connected")

		case client := <-p.unregister:
			p.mu.Lock()
			if _, ok := p.clients[client]; ok {
				delete(p.clients, client)
				close(client.send)
				p.stats.ActiveConnections = len(p.clients)
			}
			p.mu.Unlock()

			p.logger.WithFields(logrus.Fields{
				"client_id":     client.ID,
				"user_id":       client.UserID,
				"total_clients": len(p.clients),
			}).Info("Client disconnected")

		case message := <-p.broadcast:
			p.mu.RLock()
			for client := range p.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(p.clients, client)
				}
			}
			p.mu.RUnlock()
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	maxMessageSize, _ := c.server.config["max_message_size"].(int64)
	pongWait, _ := c.server.config["pong_wait"].(time.Duration)

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.WithError(err).WithField("client_id", c.ID).Error("WebSocket read error")
			}
			break
		}

		// Parse message
		var message gateway.Message
		if err := json.Unmarshal(messageData, &message); err != nil {
			c.server.logger.WithError(err).WithField("client_id", c.ID).Error("Failed to parse message")
			continue
		}

		// Set message metadata
		message.From = c.UserID
		message.Timestamp = time.Now()

		// Handle message based on type
		c.handleMessage(&message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	pingPeriod, _ := c.server.config["ping_period"].(time.Duration)
	writeWait, _ := c.server.config["write_wait"].(time.Duration)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages
func (c *Client) handleMessage(message *gateway.Message) {
	switch message.Type {
	case "ping":
		// Respond with pong
		pongMessage := &gateway.Message{
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Type:      "pong",
			Timestamp: time.Now(),
		}
		data, _ := json.Marshal(pongMessage)
		c.send <- data

	case "broadcast":
		// Broadcast message to all clients
		c.server.BroadcastMessage(context.Background(), &gateway.BroadcastRequest{
			Message: message,
		})

	case "private":
		// Send private message to specific user
		if message.To != "" {
			c.server.SendMessage(context.Background(), &gateway.SendMessageRequest{
				UserID:  message.To,
				Message: message,
			})
		}

	default:
		c.server.logger.WithFields(logrus.Fields{
			"client_id":    c.ID,
			"message_type": message.Type,
		}).Debug("Unknown message type")
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// mockResponseWriter is a mock http.ResponseWriter for testing
type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}
