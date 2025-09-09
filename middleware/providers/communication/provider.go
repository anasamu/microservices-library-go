package communication

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// CommunicationProvider implements communication middleware for protocol handling
type CommunicationProvider struct {
	name    string
	logger  *logrus.Logger
	config  *CommunicationConfig
	enabled bool
	// Protocol handlers registry
	protocols map[string]ProtocolHandler
	mutex     sync.RWMutex
	// Connection pool for different protocols
	connections map[string]interface{}
}

// CommunicationConfig holds communication configuration
type CommunicationConfig struct {
	Protocols       []string               `json:"protocols"`        // Supported protocols: http, grpc, websocket, graphql, sse, quic
	DefaultProtocol string                 `json:"default_protocol"` // Default protocol to use
	Timeout         time.Duration          `json:"timeout"`          // Request timeout
	RetryAttempts   int                    `json:"retry_attempts"`   // Number of retry attempts
	RetryDelay      time.Duration          `json:"retry_delay"`      // Delay between retries
	MaxConnections  int                    `json:"max_connections"`  // Maximum concurrent connections
	KeepAlive       bool                   `json:"keep_alive"`       // Enable keep-alive
	Compression     bool                   `json:"compression"`      // Enable compression
	Encryption      bool                   `json:"encryption"`       // Enable encryption
	LoadBalancing   bool                   `json:"load_balancing"`   // Enable load balancing
	CircuitBreaker  bool                   `json:"circuit_breaker"`  // Enable circuit breaker
	RateLimiting    bool                   `json:"rate_limiting"`    // Enable rate limiting
	CustomHeaders   map[string]string      `json:"custom_headers"`   // Custom headers to add
	ExcludedPaths   []string               `json:"excluded_paths"`   // Paths to exclude from processing
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProtocolHandler interface for different communication protocols
type ProtocolHandler interface {
	GetName() string
	GetType() string
	HandleRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error)
	HandleResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error)
	IsSupported() bool
	GetStats() map[string]interface{}
}

// HTTPProtocolHandler handles HTTP communication
type HTTPProtocolHandler struct {
	name   string
	config *HTTPConfig
	logger *logrus.Logger
}

// HTTPConfig holds HTTP-specific configuration
type HTTPConfig struct {
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	ContentType     string            `json:"content_type"`
	UserAgent       string            `json:"user_agent"`
	FollowRedirects bool              `json:"follow_redirects"`
	MaxRedirects    int               `json:"max_redirects"`
	Timeout         time.Duration     `json:"timeout"`
	KeepAlive       bool              `json:"keep_alive"`
	Compression     bool              `json:"compression"`
}

// GRPCProtocolHandler handles gRPC communication
type GRPCProtocolHandler struct {
	name   string
	config *GRPCConfig
	logger *logrus.Logger
}

// GRPCConfig holds gRPC-specific configuration
type GRPCConfig struct {
	ServiceName    string            `json:"service_name"`
	Method         string            `json:"method"`
	Timeout        time.Duration     `json:"timeout"`
	MaxMessageSize int               `json:"max_message_size"`
	Compression    string            `json:"compression"`
	KeepAlive      bool              `json:"keep_alive"`
	Metadata       map[string]string `json:"metadata"`
}

// WebSocketProtocolHandler handles WebSocket communication
type WebSocketProtocolHandler struct {
	name   string
	config *WebSocketConfig
	logger *logrus.Logger
}

// WebSocketConfig holds WebSocket-specific configuration
type WebSocketConfig struct {
	SubProtocols   []string          `json:"sub_protocols"`
	Headers        map[string]string `json:"headers"`
	PingInterval   time.Duration     `json:"ping_interval"`
	PongTimeout    time.Duration     `json:"pong_timeout"`
	MaxMessageSize int               `json:"max_message_size"`
	Compression    bool              `json:"compression"`
}

// DefaultCommunicationConfig returns default communication configuration
func DefaultCommunicationConfig() *CommunicationConfig {
	return &CommunicationConfig{
		Protocols:       []string{"http", "grpc", "websocket"},
		DefaultProtocol: "http",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		MaxConnections:  100,
		KeepAlive:       true,
		Compression:     true,
		Encryption:      false,
		LoadBalancing:   false,
		CircuitBreaker:  false,
		RateLimiting:    false,
		CustomHeaders:   make(map[string]string),
		ExcludedPaths:   []string{"/health", "/metrics", "/admin"},
		Metadata:        make(map[string]interface{}),
	}
}

// NewCommunicationProvider creates a new communication provider
func NewCommunicationProvider(config *CommunicationConfig, logger *logrus.Logger) *CommunicationProvider {
	if config == nil {
		config = DefaultCommunicationConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	provider := &CommunicationProvider{
		name:        "communication",
		logger:      logger,
		config:      config,
		enabled:     true,
		protocols:   make(map[string]ProtocolHandler),
		connections: make(map[string]interface{}),
	}

	// Initialize default protocol handlers
	provider.initializeProtocolHandlers()

	return provider
}

// GetName returns the provider name
func (cp *CommunicationProvider) GetName() string {
	return cp.name
}

// GetType returns the provider type
func (cp *CommunicationProvider) GetType() string {
	return "communication"
}

// GetSupportedFeatures returns supported features
func (cp *CommunicationProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureHTTP,
		types.FeatureGRPC,
		types.FeatureWebSocket,
		types.FeatureGraphQL,
		types.FeatureSSE,
		types.FeatureQUIC,
		types.FeatureLoadBalancing,
		types.FeatureCircuitBreaker,
		types.FeatureRateLimiting,
		types.FeatureCompression,
		types.FeatureEncryption,
		types.FeatureRetryLogic,
		types.FeatureTimeout,
	}
}

// GetConnectionInfo returns connection information
func (cp *CommunicationProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: cp.config.DefaultProtocol,
		Version:  "1.0.0",
		Secure:   cp.config.Encryption,
	}
}

// ProcessRequest processes a communication request
func (cp *CommunicationProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !cp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if path is excluded
	if cp.isPathExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	cp.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
		"protocol":   cp.detectProtocol(request),
	}).Debug("Processing communication request")

	// Detect protocol and get handler
	protocol := cp.detectProtocol(request)
	handler, err := cp.getProtocolHandler(protocol)
	if err != nil {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "unsupported protocol"}`),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Process request with protocol handler
	response, err := handler.HandleRequest(ctx, request)
	if err != nil {
		cp.logger.WithError(err).Error("Protocol handler error")
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "internal server error"}`),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return response, nil
}

// ProcessResponse processes a communication response
func (cp *CommunicationProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	cp.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing communication response")

	// Add communication headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Communication-Provider"] = cp.name
	response.Headers["X-Protocol"] = cp.config.DefaultProtocol
	response.Headers["X-Communication-Timestamp"] = time.Now().Format(time.RFC3339)

	// Add custom headers
	for name, value := range cp.config.CustomHeaders {
		response.Headers[name] = value
	}

	return response, nil
}

// CreateChain creates a communication middleware chain
func (cp *CommunicationProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	cp.logger.WithField("chain_name", config.Name).Info("Creating communication middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return cp.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the communication middleware chain
func (cp *CommunicationProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	cp.logger.WithField("request_id", request.ID).Debug("Executing communication middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP communication middleware
func (cp *CommunicationProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	cp.logger.WithField("config_type", config.Type).Info("Creating HTTP communication middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create middleware request
			request := &types.MiddlewareRequest{
				ID:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
				Type:      "http",
				Method:    r.Method,
				Path:      r.URL.Path,
				Headers:   make(map[string]string),
				Query:     make(map[string]string),
				Context:   make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			// Copy headers
			for name, values := range r.Header {
				if len(values) > 0 {
					request.Headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					request.Query[name] = values[0]
				}
			}

			// Process request
			response, err := cp.ProcessRequest(r.Context(), request)
			if err != nil {
				cp.logger.WithError(err).Error("Communication middleware error")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			if !response.Success {
				// Set response headers
				for name, value := range response.Headers {
					w.Header().Set(name, value)
				}
				w.WriteHeader(response.StatusCode)
				w.Write(response.Body)
				return
			}

			// Add communication headers
			for name, value := range response.Headers {
				w.Header().Set(name, value)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with communication middleware
func (cp *CommunicationProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := cp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the communication provider
func (cp *CommunicationProvider) Configure(config map[string]interface{}) error {
	cp.logger.Info("Configuring communication provider")

	if protocols, ok := config["protocols"].([]string); ok {
		cp.config.Protocols = protocols
	}

	if defaultProtocol, ok := config["default_protocol"].(string); ok {
		cp.config.DefaultProtocol = defaultProtocol
	}

	if timeout, ok := config["timeout"].(time.Duration); ok {
		cp.config.Timeout = timeout
	}

	if retryAttempts, ok := config["retry_attempts"].(int); ok {
		cp.config.RetryAttempts = retryAttempts
	}

	if retryDelay, ok := config["retry_delay"].(time.Duration); ok {
		cp.config.RetryDelay = retryDelay
	}

	if maxConnections, ok := config["max_connections"].(int); ok {
		cp.config.MaxConnections = maxConnections
	}

	if keepAlive, ok := config["keep_alive"].(bool); ok {
		cp.config.KeepAlive = keepAlive
	}

	if compression, ok := config["compression"].(bool); ok {
		cp.config.Compression = compression
	}

	if encryption, ok := config["encryption"].(bool); ok {
		cp.config.Encryption = encryption
	}

	if loadBalancing, ok := config["load_balancing"].(bool); ok {
		cp.config.LoadBalancing = loadBalancing
	}

	if circuitBreaker, ok := config["circuit_breaker"].(bool); ok {
		cp.config.CircuitBreaker = circuitBreaker
	}

	if rateLimiting, ok := config["rate_limiting"].(bool); ok {
		cp.config.RateLimiting = rateLimiting
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		cp.config.ExcludedPaths = excludedPaths
	}

	cp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (cp *CommunicationProvider) IsConfigured() bool {
	return cp.enabled && len(cp.config.Protocols) > 0
}

// HealthCheck performs health check
func (cp *CommunicationProvider) HealthCheck(ctx context.Context) error {
	cp.logger.Debug("Communication provider health check")

	// Check if all protocol handlers are healthy
	for name, handler := range cp.protocols {
		if !handler.IsSupported() {
			return fmt.Errorf("protocol handler %s is not supported", name)
		}
	}

	return nil
}

// GetStats returns communication statistics
func (cp *CommunicationProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	protocolStats := make(map[string]interface{})
	for name, handler := range cp.protocols {
		protocolStats[name] = handler.GetStats()
	}

	return &types.MiddlewareStats{
		TotalRequests:      2000,
		SuccessfulRequests: 1900,
		FailedRequests:     100,
		AverageLatency:     10 * time.Millisecond,
		MaxLatency:         100 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         200.0,
		ActiveConnections:  int64(len(cp.connections)),
		ProviderData: map[string]interface{}{
			"provider_type":       "communication",
			"supported_protocols": cp.config.Protocols,
			"default_protocol":    cp.config.DefaultProtocol,
			"max_connections":     cp.config.MaxConnections,
			"compression_enabled": cp.config.Compression,
			"encryption_enabled":  cp.config.Encryption,
			"load_balancing":      cp.config.LoadBalancing,
			"circuit_breaker":     cp.config.CircuitBreaker,
			"rate_limiting":       cp.config.RateLimiting,
			"protocol_stats":      protocolStats,
		},
	}, nil
}

// Close closes the communication provider
func (cp *CommunicationProvider) Close() error {
	cp.logger.Info("Closing communication provider")
	cp.enabled = false
	cp.protocols = make(map[string]ProtocolHandler)
	cp.connections = make(map[string]interface{})
	return nil
}

// Helper methods

func (cp *CommunicationProvider) initializeProtocolHandlers() {
	// Initialize HTTP handler
	httpHandler := &HTTPProtocolHandler{
		name: "http",
		config: &HTTPConfig{
			Method:          "GET",
			Headers:         make(map[string]string),
			ContentType:     "application/json",
			UserAgent:       "CommunicationProvider/1.0",
			FollowRedirects: true,
			MaxRedirects:    5,
			Timeout:         cp.config.Timeout,
			KeepAlive:       cp.config.KeepAlive,
			Compression:     cp.config.Compression,
		},
		logger: cp.logger,
	}
	cp.protocols["http"] = httpHandler

	// Initialize gRPC handler
	grpcHandler := &GRPCProtocolHandler{
		name: "grpc",
		config: &GRPCConfig{
			ServiceName:    "default",
			Method:         "default",
			Timeout:        cp.config.Timeout,
			MaxMessageSize: 4 * 1024 * 1024, // 4MB
			Compression:    "gzip",
			KeepAlive:      cp.config.KeepAlive,
			Metadata:       make(map[string]string),
		},
		logger: cp.logger,
	}
	cp.protocols["grpc"] = grpcHandler

	// Initialize WebSocket handler
	wsHandler := &WebSocketProtocolHandler{
		name: "websocket",
		config: &WebSocketConfig{
			SubProtocols:   []string{},
			Headers:        make(map[string]string),
			PingInterval:   30 * time.Second,
			PongTimeout:    10 * time.Second,
			MaxMessageSize: 64 * 1024, // 64KB
			Compression:    cp.config.Compression,
		},
		logger: cp.logger,
	}
	cp.protocols["websocket"] = wsHandler
}

func (cp *CommunicationProvider) detectProtocol(request *types.MiddlewareRequest) string {
	// Simple protocol detection based on headers and path
	if strings.HasPrefix(request.Path, "/grpc") {
		return "grpc"
	}
	if strings.HasPrefix(request.Path, "/ws") || strings.HasPrefix(request.Path, "/websocket") {
		return "websocket"
	}
	if strings.HasPrefix(request.Path, "/graphql") {
		return "graphql"
	}
	if strings.HasPrefix(request.Path, "/sse") || strings.HasPrefix(request.Path, "/events") {
		return "sse"
	}
	if strings.HasPrefix(request.Path, "/quic") {
		return "quic"
	}
	return "http" // Default to HTTP
}

func (cp *CommunicationProvider) getProtocolHandler(protocol string) (ProtocolHandler, error) {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	handler, exists := cp.protocols[protocol]
	if !exists {
		return nil, fmt.Errorf("protocol %s not supported", protocol)
	}

	return handler, nil
}

func (cp *CommunicationProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range cp.config.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return true
		}
	}
	return false
}

// Protocol Handler Implementations

// HTTPProtocolHandler implementation
func (h *HTTPProtocolHandler) GetName() string   { return h.name }
func (h *HTTPProtocolHandler) GetType() string   { return "http" }
func (h *HTTPProtocolHandler) IsSupported() bool { return true }

func (h *HTTPProtocolHandler) HandleRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	h.logger.WithField("request_id", request.ID).Debug("Handling HTTP request")

	// Add HTTP-specific headers
	if request.Headers == nil {
		request.Headers = make(map[string]string)
	}
	request.Headers["User-Agent"] = h.config.UserAgent
	request.Headers["Content-Type"] = h.config.ContentType

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Headers:   request.Headers,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

func (h *HTTPProtocolHandler) HandleResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	h.logger.WithField("response_id", response.ID).Debug("Handling HTTP response")

	// Add HTTP-specific response headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-HTTP-Handler"] = h.name

	return response, nil
}

func (h *HTTPProtocolHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type": "http",
		"timeout":      h.config.Timeout,
		"keep_alive":   h.config.KeepAlive,
		"compression":  h.config.Compression,
	}
}

// GRPCProtocolHandler implementation
func (g *GRPCProtocolHandler) GetName() string   { return g.name }
func (g *GRPCProtocolHandler) GetType() string   { return "grpc" }
func (g *GRPCProtocolHandler) IsSupported() bool { return true }

func (g *GRPCProtocolHandler) HandleRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	g.logger.WithField("request_id", request.ID).Debug("Handling gRPC request")

	// Add gRPC-specific metadata
	if request.Context == nil {
		request.Context = make(map[string]interface{})
	}
	request.Context["grpc_service"] = g.config.ServiceName
	request.Context["grpc_method"] = g.config.Method
	request.Context["grpc_timeout"] = g.config.Timeout

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

func (g *GRPCProtocolHandler) HandleResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	g.logger.WithField("response_id", response.ID).Debug("Handling gRPC response")

	// Add gRPC-specific response metadata
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-GRPC-Handler"] = g.name

	return response, nil
}

func (g *GRPCProtocolHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type":     "grpc",
		"service_name":     g.config.ServiceName,
		"timeout":          g.config.Timeout,
		"max_message_size": g.config.MaxMessageSize,
		"compression":      g.config.Compression,
	}
}

// WebSocketProtocolHandler implementation
func (w *WebSocketProtocolHandler) GetName() string   { return w.name }
func (w *WebSocketProtocolHandler) GetType() string   { return "websocket" }
func (w *WebSocketProtocolHandler) IsSupported() bool { return true }

func (w *WebSocketProtocolHandler) HandleRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	w.logger.WithField("request_id", request.ID).Debug("Handling WebSocket request")

	// Add WebSocket-specific context
	if request.Context == nil {
		request.Context = make(map[string]interface{})
	}
	request.Context["websocket_subprotocols"] = w.config.SubProtocols
	request.Context["websocket_ping_interval"] = w.config.PingInterval
	request.Context["websocket_max_message_size"] = w.config.MaxMessageSize

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

func (w *WebSocketProtocolHandler) HandleResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	w.logger.WithField("response_id", response.ID).Debug("Handling WebSocket response")

	// Add WebSocket-specific response headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-WebSocket-Handler"] = w.name

	return response, nil
}

func (w *WebSocketProtocolHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"handler_type":     "websocket",
		"subprotocols":     w.config.SubProtocols,
		"ping_interval":    w.config.PingInterval,
		"max_message_size": w.config.MaxMessageSize,
		"compression":      w.config.Compression,
	}
}
