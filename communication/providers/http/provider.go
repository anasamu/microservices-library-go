package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/libs/communication/gateway"
	"github.com/sirupsen/logrus"
)

// Provider implements CommunicationProvider for HTTP
type Provider struct {
	server   *http.Server
	config   map[string]interface{}
	logger   *logrus.Logger
	mu       sync.RWMutex
	handlers map[string]http.HandlerFunc
	stats    *gateway.CommunicationStats
}

// NewProvider creates a new HTTP communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:   make(map[string]interface{}),
		logger:   logger,
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
	return "http"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.CommunicationFeature {
	return []gateway.CommunicationFeature{
		gateway.FeatureHTTP,
		gateway.FeatureHTTPS,
		gateway.FeatureCORS,
		gateway.FeatureCompression,
		gateway.FeatureAuthentication,
		gateway.FeatureRateLimiting,
		gateway.FeatureMetrics,
		gateway.FeatureLogging,
		gateway.FeatureHealthChecks,
		gateway.FeatureMiddleware,
		gateway.FeatureStaticFiles,
		gateway.FeatureTemplates,
		gateway.FeatureSessions,
		gateway.FeatureCookies,
		gateway.FeatureHeaders,
		gateway.FeatureQueryParams,
		gateway.FeaturePathParams,
		gateway.FeatureBodyParsing,
		gateway.FeatureFileUpload,
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
		ID:          "http-server",
		Type:        "http",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the HTTP provider
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

	p.config = config

	p.logger.Info("HTTP provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Start starts the HTTP server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("http provider not configured")
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
	}).Info("HTTP server started successfully")

	// Start server in goroutine
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.WithError(err).Error("HTTP server error")
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (p *Provider) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	p.logger.Info("Stopping HTTP server")
	return p.server.Shutdown(ctx)
}

// IsRunning checks if the HTTP server is running
func (p *Provider) IsRunning() bool {
	return p.server != nil
}

// HandleRequest handles an HTTP request
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("http server not running")
	}

	start := time.Now()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, request.Path, nil)
	if err != nil {
		p.stats.FailedRequests++
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set query parameters
	if len(request.QueryParams) > 0 {
		query := httpReq.URL.Query()
		for key, value := range request.QueryParams {
			query.Add(key, value)
		}
		httpReq.URL.RawQuery = query.Encode()
	}

	// Set body
	if len(request.Body) > 0 {
		httpReq.Body = http.NoBody // Simplified for this example
	}

	// Create response recorder
	recorder := &responseRecorder{
		statusCode: http.StatusOK,
		headers:    make(map[string]string),
		body:       make([]byte, 0),
	}

	// Handle request
	handler := p.getHandler(request.Path)
	if handler == nil {
		handler = p.defaultHandler
	}

	handler(recorder, httpReq)

	// Update stats
	duration := time.Since(start)
	p.mu.Lock()
	p.stats.TotalRequests++
	if recorder.statusCode >= 400 {
		p.stats.FailedRequests++
	}
	p.stats.AverageResponseTime = (p.stats.AverageResponseTime + duration) / 2
	p.mu.Unlock()

	response := &gateway.Response{
		StatusCode: recorder.statusCode,
		Headers:    recorder.headers,
		Body:       recorder.body,
		Metadata: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		},
	}

	return response, nil
}

// HandleWebSocket handles a WebSocket connection (not supported by HTTP provider)
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	return nil, fmt.Errorf("WebSocket not supported by HTTP provider")
}

// SendMessage sends a message (not applicable for HTTP provider)
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	return nil, fmt.Errorf("message sending not applicable for HTTP provider")
}

// BroadcastMessage broadcasts a message (not applicable for HTTP provider)
func (p *Provider) BroadcastMessage(ctx context.Context, request *gateway.BroadcastRequest) (*gateway.BroadcastResponse, error) {
	return nil, fmt.Errorf("message broadcasting not applicable for HTTP provider")
}

// GetConnections gets HTTP connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	// HTTP is stateless, so we return server info
	connections := []gateway.ConnectionInfo{
		*p.GetConnectionInfo(),
	}
	return connections, nil
}

// GetConnectionCount gets HTTP connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	// HTTP is stateless, return 1 for the server
	return 1, nil
}

// CloseConnection closes a connection (not applicable for HTTP provider)
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	return fmt.Errorf("connection closing not applicable for HTTP provider")
}

// HealthCheck performs a health check on the HTTP server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("http server not running")
	}

	// Create a test request
	request := &gateway.Request{
		Method: "GET",
		Path:   "/health",
	}

	_, err := p.HandleRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("http health check failed: %w", err)
	}

	return nil
}

// GetStats returns HTTP server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ProviderData = map[string]interface{}{
		"server_addr": p.server.Addr,
		"handlers":    len(p.handlers),
	}

	return &stats, nil
}

// Close closes the HTTP provider
func (p *Provider) Close() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}
	return nil
}

// setupMiddleware sets up HTTP middleware
func (p *Provider) setupMiddleware() http.Handler {
	mux := http.NewServeMux()

	// Add health check endpoint
	mux.HandleFunc("/health", p.healthHandler)

	// Add registered handlers
	for path, handler := range p.handlers {
		mux.HandleFunc(path, handler)
	}

	// Add default handler for unmatched routes
	mux.HandleFunc("/", p.defaultHandler)

	// Wrap with middleware
	handler := p.corsMiddleware(mux)
	handler = p.loggingMiddleware(handler)
	handler = p.metricsMiddleware(handler)

	return handler
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
		}).Info("HTTP request")
	})
}

// metricsMiddleware adds metrics collection
func (p *Provider) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		p.logger.WithFields(logrus.Fields{
			"metric_type": "http_request_duration",
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": duration.Milliseconds(),
		}).Debug("HTTP metrics")
	})
}

// healthHandler handles health check requests
func (p *Provider) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","provider":"http"}`, time.Now().UTC().Format(time.RFC3339))
}

// defaultHandler handles unmatched routes
func (p *Provider) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error":"not found","path":"%s","method":"%s"}`, r.URL.Path, r.Method)
}

// getHandler gets a handler for a specific path
func (p *Provider) getHandler(path string) http.HandlerFunc {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.handlers[path]
}

// RegisterHandler registers a handler for a specific path
func (p *Provider) RegisterHandler(path string, handler http.HandlerFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[path] = handler
}

// responseRecorder records HTTP response data
type responseRecorder struct {
	statusCode int
	headers    map[string]string
	body       []byte
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
