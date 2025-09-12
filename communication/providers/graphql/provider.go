package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/communication"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/sirupsen/logrus"
)

// Provider implements CommunicationProvider for GraphQL
type Provider struct {
	server   *http.Server
	config   map[string]interface{}
	logger   *logrus.Logger
	mu       sync.RWMutex
	schema   *graphql.Schema
	handlers map[string]*handler.Handler
	stats    *gateway.CommunicationStats
	clients  map[string]*GraphQLClient
}

// GraphQLClient represents a GraphQL client connection
type GraphQLClient struct {
	ID          string
	UserID      string
	ConnectedAt time.Time
	LastSeen    time.Time
	Metadata    map[string]interface{}
}

// NewProvider creates a new GraphQL communication provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config:   make(map[string]interface{}),
		logger:   logger,
		handlers: make(map[string]*handler.Handler),
		clients:  make(map[string]*GraphQLClient),
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
	return "graphql"
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
		ID:          "graphql-server",
		Type:        "graphql",
		RemoteAddr:  fmt.Sprintf("%s:%d", host, port),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Configure configures the GraphQL provider
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

	// Create default schema if not provided
	if p.schema == nil {
		schema, err := p.createDefaultSchema()
		if err != nil {
			return fmt.Errorf("failed to create default schema: %w", err)
		}
		p.schema = schema
	}

	p.config = config

	p.logger.Info("GraphQL provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0 && p.schema != nil
}

// Start starts the GraphQL server
func (p *Provider) Start(ctx context.Context, config map[string]interface{}) error {
	if !p.IsConfigured() {
		return fmt.Errorf("graphql provider not configured")
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
	}).Info("GraphQL server started successfully")

	// Start server in goroutine
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.WithError(err).Error("GraphQL server error")
		}
	}()

	return nil
}

// Stop stops the GraphQL server
func (p *Provider) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	p.logger.Info("Stopping GraphQL server")
	return p.server.Shutdown(ctx)
}

// IsRunning checks if the GraphQL server is running
func (p *Provider) IsRunning() bool {
	return p.server != nil
}

// HandleRequest handles a GraphQL request
func (p *Provider) HandleRequest(ctx context.Context, request *gateway.Request) (*gateway.Response, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("graphql server not running")
	}

	start := time.Now()

	// Parse GraphQL request
	var graphqlRequest struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operationName"`
	}

	if len(request.Body) > 0 {
		if err := json.Unmarshal(request.Body, &graphqlRequest); err != nil {
			p.stats.FailedRequests++
			return nil, fmt.Errorf("failed to parse GraphQL request: %w", err)
		}
	}

	// Execute GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:         *p.schema,
		RequestString:  graphqlRequest.Query,
		VariableValues: graphqlRequest.Variables,
		OperationName:  graphqlRequest.OperationName,
		Context:        ctx,
	})

	// Marshal response
	responseBody, err := json.Marshal(result)
	if err != nil {
		p.stats.FailedRequests++
		return nil, fmt.Errorf("failed to marshal GraphQL response: %w", err)
	}

	// Update stats
	duration := time.Since(start)
	p.mu.Lock()
	p.stats.TotalRequests++
	if len(result.Errors) > 0 {
		p.stats.FailedRequests++
	}
	p.stats.AverageResponseTime = (p.stats.AverageResponseTime + duration) / 2
	p.mu.Unlock()

	response := &gateway.Response{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: responseBody,
		Metadata: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"errors":      len(result.Errors),
		},
	}

	return response, nil
}

// HandleWebSocket handles a WebSocket connection (not supported by GraphQL provider)
func (p *Provider) HandleWebSocket(ctx context.Context, request *gateway.WebSocketRequest) (*gateway.WebSocketResponse, error) {
	return nil, fmt.Errorf("WebSocket not supported by GraphQL provider")
}

// SendMessage sends a message (not applicable for GraphQL provider)
func (p *Provider) SendMessage(ctx context.Context, request *gateway.SendMessageRequest) (*gateway.SendMessageResponse, error) {
	return nil, fmt.Errorf("message sending not applicable for GraphQL provider")
}

// BroadcastMessage broadcasts a message (not applicable for GraphQL provider)
func (p *Provider) BroadcastMessage(ctx context.Context, request *gateway.BroadcastRequest) (*gateway.BroadcastResponse, error) {
	return nil, fmt.Errorf("message broadcasting not applicable for GraphQL provider")
}

// GetConnections gets GraphQL connections
func (p *Provider) GetConnections(ctx context.Context) ([]gateway.ConnectionInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]gateway.ConnectionInfo, 0, len(p.clients))
	for _, client := range p.clients {
		connInfo := gateway.ConnectionInfo{
			ID:          client.ID,
			Type:        "graphql",
			ConnectedAt: client.ConnectedAt,
			LastSeen:    client.LastSeen,
			UserID:      client.UserID,
			Metadata:    client.Metadata,
		}
		connections = append(connections, connInfo)
	}

	return connections, nil
}

// GetConnectionCount gets GraphQL connection count
func (p *Provider) GetConnectionCount(ctx context.Context) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients), nil
}

// CloseConnection closes a GraphQL connection
func (p *Provider) CloseConnection(ctx context.Context, connectionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.clients[connectionID]; !exists {
		return fmt.Errorf("connection not found")
	}

	delete(p.clients, connectionID)
	p.stats.ActiveConnections = len(p.clients)

	p.logger.WithField("connection_id", connectionID).Info("GraphQL connection closed")
	return nil
}

// HealthCheck performs a health check on the GraphQL server
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsRunning() {
		return fmt.Errorf("graphql server not running")
	}

	// Create a test GraphQL request
	request := &gateway.Request{
		Method: "POST",
		Path:   "/graphql",
		Body:   []byte(`{"query":"{ __schema { queryType { name } } }"}`),
	}

	_, err := p.HandleRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("graphql health check failed: %w", err)
	}

	return nil
}

// GetStats returns GraphQL server statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.CommunicationStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy of stats
	stats := *p.stats
	stats.ActiveConnections = len(p.clients)
	stats.ProviderData = map[string]interface{}{
		"server_addr": p.server.Addr,
		"clients":     len(p.clients),
		"handlers":    len(p.handlers),
	}

	return &stats, nil
}

// Close closes the GraphQL provider
func (p *Provider) Close() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}
	return nil
}

// setupMiddleware sets up GraphQL middleware
func (p *Provider) setupMiddleware() http.Handler {
	mux := http.NewServeMux()

	// Add GraphQL endpoint
	graphqlHandler := handler.New(&handler.Config{
		Schema:   p.schema,
		Pretty:   true,
		GraphiQL: true,
	})
	mux.Handle("/graphql", graphqlHandler)

	// Add health check endpoint
	mux.HandleFunc("/health", p.healthHandler)

	// Add registered handlers
	for path, handler := range p.handlers {
		mux.Handle(path, handler)
	}

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
		}).Info("GraphQL request")
	})
}

// metricsMiddleware adds metrics collection
func (p *Provider) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		p.logger.WithFields(logrus.Fields{
			"metric_type": "graphql_request_duration",
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": duration.Milliseconds(),
		}).Debug("GraphQL metrics")
	})
}

// healthHandler handles health check requests
func (p *Provider) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","provider":"graphql"}`, time.Now().UTC().Format(time.RFC3339))
}

// createDefaultSchema creates a default GraphQL schema
func (p *Provider) createDefaultSchema() (*graphql.Schema, error) {
	// Define a simple query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hello": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Hello, GraphQL!", nil
				},
			},
			"health": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "healthy", nil
				},
			},
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL schema: %w", err)
	}

	return &schema, nil
}

// SetSchema sets a custom GraphQL schema
func (p *Provider) SetSchema(schema *graphql.Schema) {
	p.schema = schema
}

// RegisterHandler registers a custom handler for a specific path
func (p *Provider) RegisterHandler(path string, handler *handler.Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[path] = handler
}

// RegisterClient registers a new GraphQL client
func (p *Provider) RegisterClient(clientID, userID string, metadata map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	client := &GraphQLClient{
		ID:          clientID,
		UserID:      userID,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Metadata:    metadata,
	}

	p.clients[clientID] = client
	p.stats.TotalConnections++
	p.stats.ActiveConnections = len(p.clients)

	p.logger.WithFields(logrus.Fields{
		"client_id": clientID,
		"user_id":   userID,
	}).Info("GraphQL client registered")
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
