package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/communication"
	"github.com/sirupsen/logrus"
)

// Provider implements CommunicationProvider for Server-Sent Events
type Provider struct {
	server   *http.Server
	config   map[string]interface{}
	logger   *logrus.Logger
	mu       sync.RWMutex
	clients  map[string]*SSEClient
	stats    *gateway.CommunicationStats
	handlers map[string]http.HandlerFunc
}

// SSEClient represents an SSE client connection
type SSEClient struct {
	ID          string
	UserID      string
	ConnectedAt time.Time
	LastSeen    time.Time
	Metadata    map[string]interface{}
	writer      http.ResponseWriter
	flusher     http.Flusher
	done        chan bool
}

// NewProvider creates a new SSE communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:   make(map[string]interface{}),
		logger:   logger,
		clients:  make(map[string]*SSEClient),
		handlers: make(map[string]http.HandlerFunc),
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
	return "sse"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.CommunicationFeature {
	return []gateway.CommunicationFeature{
		gateway.FeatureHTTP,
		gateway.FeatureHTTPS,
		gateway.FeatureServerSentEvents,
		gateway.FeatureCORS,
		gateway.FeatureCompression,
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
		gateway.FeatureBroadcasting,
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
		ID:          "sse-server",
		Type:        "sse",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the SSE provider
func (p *Provider) Configure(config map[string]interface{}) error {
	host, ok := config["host"].(string)
	if !ok || host == "" {
		host = "0.0.0.0"
	}

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		port = 8080
	}

	readTimeout, ok := config["read_timeout"].(time.Duration)
	if !ok || readTimeout == 0 {
		readTimeout = 30 * time.Second
	}

	writeTimeout, ok := config["write_timeout"].(time.Duration)
	if !ok || writeTimeout == 0 {
		writeTimeout = 30 * time.Second
	}

	idleTimeout, ok := config["idle_timeout"].(time.Duration)
	if !ok || idleTimeout == 0 {
		idleTimeout = 120 * time.Second
	}

	heartbeatInterval, ok := config["heartbeat_interval"].(time.Duration)
	if !ok || heartbeatInterval == 0 {
		heartbeatInterval = 30 * time.Second
	}

	p.config = config

	p.logger.Info("SSE provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Start starts the SSE server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("sse provider not configured")
	}

	// Merge provided config with existing config
	for key, value := range config {
		p.config[key] = value
	}

	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	readTimeout, _ := p.config["read_timeout"].(time.Duration)
	writeTimeout, _ := p.config["write_timeout"].(time.Duration)
	idleTimeout, _ := p.config["idle_timeout"].(time.Duration)

	// Create HTTP server
	p.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      p.setupMiddleware(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	p.logger.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Info("SSE server started successfully")

	// Start server in goroutine
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.WithError(err).Error("SSE server error")
		}
	}()

	// Start heartbeat goroutine
	go p.startHeartbeat()

	return nil
}

// Stop stops the SSE server
func (p *Provider) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	p.logger.Info("Stopping SSE server")

	// Close all clients
	p.mu.Lock()
	for _, client := range p.clients {
		client.done <- true
	}
	p.clients = make(map[string]*SSEClient)
	p.mu.Unlock()

	return p.server.Shutdown(ctx)
}

// IsRunning checks if the SSE server is running
func (p *Provider) IsRunning() bool {
	return p.server != nil
}

// HandleRequest handles an HTTP request
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("sse server not running")
	}

	start := time.Now()

	// Handle SSE connection request
	if request.Path == "/events" {
		// This would typically be handled by the HTTP server
		// For now, return a simple response
		response := &gateway.Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":                "text/event-stream",
				"Cache-Control":               "no-cache",
				"Connection":                  "keep-alive",
				"Access-Control-Allow-Origin": "*",
			},
			Body: []byte("data: SSE connection established\n\n"),
			Metadata: map[string]interface{}{
				"duration_ms": time.Since(start).Milliseconds(),
			},
		}

		// Update stats
		p.mu.Lock()
		p.stats.TotalRequests++
		p.stats.AverageResponseTime = (p.stats.AverageResponseTime + time.Since(start)) / 2
		p.mu.Unlock()

		return response, nil
	}

	// Handle other requests
	response := &gateway.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message":"SSE request handled"}`),
		Metadata: map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
		},
	}

	// Update stats
	p.mu.Lock()
	p.stats.TotalRequests++
	p.stats.AverageResponseTime = (p.stats.AverageResponseTime + time.Since(start)) / 2
	p.mu.Unlock()

	return response, nil
}

// HandleWebSocket handles a WebSocket connection (not supported by SSE provider)
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	return nil, fmt.Errorf("WebSocket not supported by SSE provider")
}

// SendMessage sends a message to a specific connection
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("sse server not running")
	}

	// Find client by connection ID or user ID
	var targetClient *SSEClient
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

	// Send SSE message
	err := p.sendSSEMessage(targetClient, request.Message)
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
		return nil, fmt.Errorf("sse server not running")
	}

	sentCount := 0
	failedCount := 0

	p.mu.RLock()
	clients := make([]*SSEClient, 0, len(p.clients))
	for _, client := range p.clients {
		clients = append(clients, client)
	}
	p.mu.RUnlock()

	for _, client := range clients {
		err := p.sendSSEMessage(client, request.Message)
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

// GetConnections gets SSE connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]gateway.ConnectionInfo, 0, len(p.clients))
	for _, client := range p.clients {
		connInfo := gateway.ConnectionInfo{
			ID:          client.ID,
			Type:        "sse",
			ConnectedAt: client.ConnectedAt,
			LastSeen:    client.LastSeen,
			UserID:      client.UserID,
			Metadata:    client.Metadata,
		}
		connections = append(connections, connInfo)
	}

	return connections, nil
}

// GetConnectionCount gets SSE connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients), nil
}

// CloseConnection closes an SSE connection
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	if !p.IsRunning() {
		return fmt.Errorf("sse server not running")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	client, exists := p.clients[connectionID]
	if !exists {
		return fmt.Errorf("connection not found")
	}

	client.done <- true
	delete(p.clients, connectionID)
	p.stats.ActiveConnections = len(p.clients)

	p.logger.WithField("connection_id", connectionID).Info("SSE connection closed")
	return nil
}

// HealthCheck performs a health check on the SSE server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("sse server not running")
	}

	// Check if we can get connection count
	_, err := p.GetConnectionCount(ctx)
	if err != nil {
		return fmt.Errorf("sse health check failed: %w", err)
	}

	return nil
}

// GetStats returns SSE server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ActiveConnections = len(p.clients)
	stats.ProviderData = map[string]interface{}{
		"clients": len(p.clients),
		"server":  p.server != nil,
	}

	return &stats, nil
}

// Close closes the SSE provider
func (p *Provider) Close() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}
	return nil
}

// setupMiddleware sets up SSE middleware
func (p *Provider) setupMiddleware() http.Handler {
	mux := http.NewServeMux()

	// Add SSE endpoint
	mux.HandleFunc("/events", p.sseHandler)

	// Add health check endpoint
	mux.HandleFunc("/health", p.healthHandler)

	// Add registered handlers
	for path, handler := range p.handlers {
		mux.HandleFunc(path, handler)
	}

	// Wrap with middleware
	handler := p.corsMiddleware(mux)
	handler = p.loggingMiddleware(handler)
	handler = p.metricsMiddleware(handler)

	return handler
}

// sseHandler handles SSE connections
func (p *Provider) sseHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Create client
	clientID := generateClientID()
	client := &SSEClient{
		ID:          clientID,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Metadata:    make(map[string]interface{}),
		writer:      w,
		flusher:     flusher,
		done:        make(chan bool),
	}

	// Register client
	p.mu.Lock()
	p.clients[clientID] = client
	p.stats.TotalConnections++
	p.stats.ActiveConnections = len(p.clients)
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"client_id":     clientID,
		"remote_addr":   r.RemoteAddr,
		"total_clients": len(p.clients),
	}).Info("SSE client connected")

	// Send initial connection message
	fmt.Fprintf(w, "data: %s\n\n", `{"type":"connected","client_id":"`+clientID+`"}`)
	flusher.Flush()

	// Wait for client to disconnect
	<-client.done

	// Unregister client
	p.mu.Lock()
	delete(p.clients, clientID)
	p.stats.ActiveConnections = len(p.clients)
	p.mu.Unlock()

	p.logger.WithField("client_id", clientID).Info("SSE client disconnected")
}

// sendSSEMessage sends an SSE message to a client
func (p *Provider) sendSSEMessage(client *SSEClient, message *gateway.Message) error {
	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send SSE message
	fmt.Fprintf(client.writer, "data: %s\n\n", string(data))
	client.flusher.Flush()

	// Update last seen
	client.LastSeen = time.Now()

	return nil
}

// startHeartbeat starts the heartbeat goroutine
func (p *Provider) startHeartbeat() {
	heartbeatInterval, _ := p.config["heartbeat_interval"].(time.Duration)
	if heartbeatInterval == 0 {
		heartbeatInterval = 30 * time.Second
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.RLock()
			clients := make([]*SSEClient, 0, len(p.clients))
			for _, client := range p.clients {
				clients = append(clients, client)
			}
			p.mu.RUnlock()

			// Send heartbeat to all clients
			heartbeatMessage := &gateway.Message{
				ID:        fmt.Sprintf("heartbeat_%d", time.Now().Unix()),
				Type:      "heartbeat",
				Content:   map[string]interface{}{"timestamp": time.Now().Unix()},
				Timestamp: time.Now(),
			}

			for _, client := range clients {
				p.sendSSEMessage(client, heartbeatMessage)
			}
		}
	}
}

// corsMiddleware adds CORS support
func (p *Provider) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware adds request logging
func (p *Provider) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		p.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      wrapped.statusCode,
			"duration":    duration,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("SSE request")
	})
}

// metricsMiddleware adds metrics collection
func (p *Provider) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		p.logger.WithFields(logrus.Fields{
			"metric_type": "sse_request_duration",
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": duration.Milliseconds(),
		}).Debug("SSE metrics")
	})
}

// healthHandler handles health check requests
func (p *Provider) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","provider":"sse"}`, time.Now().UTC().Format(time.RFC3339))
}

// RegisterHandler registers a handler for a specific path
func (p *Provider) RegisterHandler(path string, handler http.HandlerFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[path] = handler
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("sse_client_%d", time.Now().UnixNano())
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
