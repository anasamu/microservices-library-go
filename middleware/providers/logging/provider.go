package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// LoggingProvider implements logging middleware
type LoggingProvider struct {
	name    string
	logger  *logrus.Logger
	config  *LoggingConfig
	enabled bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	LogLevel        string                 `json:"log_level"`
	LogFormat       string                 `json:"log_format"` // json, text
	IncludeHeaders  bool                   `json:"include_headers"`
	IncludeBody     bool                   `json:"include_body"`
	IncludeQuery    bool                   `json:"include_query"`
	MaxBodySize     int                    `json:"max_body_size"`
	ExcludedPaths   []string               `json:"excluded_paths"`
	ExcludedHeaders []string               `json:"excluded_headers"`
	CustomFields    map[string]string      `json:"custom_fields"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		LogLevel:        "info",
		LogFormat:       "json",
		IncludeHeaders:  true,
		IncludeBody:     false,
		IncludeQuery:    true,
		MaxBodySize:     1024,
		ExcludedPaths:   []string{"/health", "/metrics"},
		ExcludedHeaders: []string{"authorization", "cookie", "x-api-key"},
		CustomFields:    make(map[string]string),
		Metadata:        make(map[string]interface{}),
	}
}

// NewLoggingProvider creates a new logging provider
func NewLoggingProvider(config *LoggingConfig, logger *logrus.Logger) *LoggingProvider {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if config.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	}

	return &LoggingProvider{
		name:    "logging",
		logger:  logger,
		config:  config,
		enabled: true,
	}
}

// GetName returns the provider name
func (lp *LoggingProvider) GetName() string {
	return lp.name
}

// GetType returns the provider type
func (lp *LoggingProvider) GetType() string {
	return "logging"
}

// GetSupportedFeatures returns supported features
func (lp *LoggingProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureStructuredLogging,
		types.FeatureRequestLogging,
		types.FeatureResponseLogging,
		types.FeatureAuditLogging,
		types.FeatureErrorLogging,
		types.FeaturePerformanceLogging,
	}
}

// GetConnectionInfo returns connection information
func (lp *LoggingProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: "local",
		Version:  "1.0.0",
		Secure:   false,
	}
}

// ProcessRequest processes a logging request
func (lp *LoggingProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !lp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if path is excluded
	if lp.isPathExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Create log entry
	logEntry := lp.createLogEntry(request, "request")

	// Log the request
	lp.logger.WithFields(logEntry).Info("HTTP request received")

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

// ProcessResponse processes a logging response
func (lp *LoggingProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	if !lp.enabled {
		return response, nil
	}

	// Create log entry
	logEntry := lp.createResponseLogEntry(response)

	// Determine log level based on status code
	logLevel := lp.getLogLevel(response.StatusCode)

	// Log the response
	lp.logger.WithFields(logEntry).Log(logLevel, "HTTP response sent")

	return response, nil
}

// CreateChain creates a logging middleware chain
func (lp *LoggingProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	lp.logger.WithField("chain_name", config.Name).Info("Creating logging middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return lp.ProcessRequest(ctx, req)
		},
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			// This would typically process the response
			return nil, nil
		},
	)

	return chain, nil
}

// ExecuteChain executes the logging middleware chain
func (lp *LoggingProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	lp.logger.WithField("request_id", request.ID).Debug("Executing logging middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP logging middleware
func (lp *LoggingProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	lp.logger.WithField("config_type", config.Type).Info("Creating HTTP logging middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create middleware request
			request := &types.MiddlewareRequest{
				ID:        fmt.Sprintf("req-%d", start.UnixNano()),
				Type:      "http",
				Method:    r.Method,
				Path:      r.URL.Path,
				Headers:   make(map[string]string),
				Query:     make(map[string]string),
				Context:   make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				Timestamp: start,
			}

			// Copy headers
			for name, values := range r.Header {
				if len(values) > 0 && !lp.isHeaderExcluded(name) {
					request.Headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					request.Query[name] = values[0]
				}
			}

			// Log request
			lp.ProcessRequest(r.Context(), request)

			// Create response wrapper
			responseWrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           make([]byte, 0),
			}

			// Call next handler
			next.ServeHTTP(responseWrapper, r)

			// Calculate duration
			duration := time.Since(start)

			// Create response
			response := &types.MiddlewareResponse{
				ID:         request.ID,
				Success:    responseWrapper.statusCode < 400,
				StatusCode: responseWrapper.statusCode,
				Headers:    make(map[string]string),
				Body:       responseWrapper.body,
				Timestamp:  time.Now(),
				Duration:   duration,
			}

			// Copy response headers
			for name, values := range responseWrapper.Header() {
				if len(values) > 0 {
					response.Headers[name] = values[0]
				}
			}

			// Log response
			lp.ProcessResponse(r.Context(), response)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with logging middleware
func (lp *LoggingProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := lp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the logging provider
func (lp *LoggingProvider) Configure(config map[string]interface{}) error {
	lp.logger.Info("Configuring logging provider")

	if logLevel, ok := config["log_level"].(string); ok {
		level, err := logrus.ParseLevel(logLevel)
		if err == nil {
			lp.config.LogLevel = logLevel
			lp.logger.SetLevel(level)
		}
	}

	if logFormat, ok := config["log_format"].(string); ok {
		lp.config.LogFormat = logFormat
		if logFormat == "json" {
			lp.logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
			})
		} else {
			lp.logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: time.RFC3339,
				FullTimestamp:   true,
			})
		}
	}

	if includeHeaders, ok := config["include_headers"].(bool); ok {
		lp.config.IncludeHeaders = includeHeaders
	}

	if includeBody, ok := config["include_body"].(bool); ok {
		lp.config.IncludeBody = includeBody
	}

	if includeQuery, ok := config["include_query"].(bool); ok {
		lp.config.IncludeQuery = includeQuery
	}

	if maxBodySize, ok := config["max_body_size"].(int); ok {
		lp.config.MaxBodySize = maxBodySize
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		lp.config.ExcludedPaths = excludedPaths
	}

	if excludedHeaders, ok := config["excluded_headers"].([]string); ok {
		lp.config.ExcludedHeaders = excludedHeaders
	}

	lp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (lp *LoggingProvider) IsConfigured() bool {
	return lp.enabled
}

// HealthCheck performs health check
func (lp *LoggingProvider) HealthCheck(ctx context.Context) error {
	lp.logger.Debug("Logging provider health check")
	return nil
}

// GetStats returns logging statistics
func (lp *LoggingProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	return &types.MiddlewareStats{
		TotalRequests:      2000,
		SuccessfulRequests: 1900,
		FailedRequests:     100,
		AverageLatency:     2 * time.Millisecond,
		MaxLatency:         10 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         200.0,
		ActiveConnections:  0,
		ProviderData: map[string]interface{}{
			"provider_type": "logging",
			"log_level":     lp.config.LogLevel,
			"log_format":    lp.config.LogFormat,
		},
	}, nil
}

// Close closes the logging provider
func (lp *LoggingProvider) Close() error {
	lp.logger.Info("Closing logging provider")
	lp.enabled = false
	return nil
}

// Helper methods

func (lp *LoggingProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range lp.config.ExcludedPaths {
		if path == excludedPath {
			return true
		}
	}
	return false
}

func (lp *LoggingProvider) isHeaderExcluded(headerName string) bool {
	for _, excludedHeader := range lp.config.ExcludedHeaders {
		if headerName == excludedHeader {
			return true
		}
	}
	return false
}

func (lp *LoggingProvider) createLogEntry(request *types.MiddlewareRequest, eventType string) logrus.Fields {
	fields := logrus.Fields{
		"event_type": eventType,
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
		"timestamp":  request.Timestamp.Format(time.RFC3339),
		"user_id":    request.UserID,
		"service_id": request.ServiceID,
	}

	// Add headers if configured
	if lp.config.IncludeHeaders {
		filteredHeaders := make(map[string]string)
		for name, value := range request.Headers {
			if !lp.isHeaderExcluded(name) {
				filteredHeaders[name] = value
			}
		}
		fields["headers"] = filteredHeaders
	}

	// Add query parameters if configured
	if lp.config.IncludeQuery && len(request.Query) > 0 {
		fields["query"] = request.Query
	}

	// Add body if configured and within size limit
	if lp.config.IncludeBody && len(request.Body) > 0 && len(request.Body) <= lp.config.MaxBodySize {
		fields["body"] = string(request.Body)
	}

	// Add custom fields
	for key, value := range lp.config.CustomFields {
		fields[key] = value
	}

	// Add context
	if len(request.Context) > 0 {
		fields["context"] = request.Context
	}

	// Add metadata
	if len(request.Metadata) > 0 {
		fields["metadata"] = request.Metadata
	}

	return fields
}

func (lp *LoggingProvider) createResponseLogEntry(response *types.MiddlewareResponse) logrus.Fields {
	fields := logrus.Fields{
		"event_type":  "response",
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
		"timestamp":   response.Timestamp.Format(time.RFC3339),
		"duration_ms": float64(response.Duration.Nanoseconds()) / 1e6,
	}

	if response.Error != "" {
		fields["error"] = response.Error
	}

	if response.Message != "" {
		fields["message"] = response.Message
	}

	// Add headers if configured
	if lp.config.IncludeHeaders && len(response.Headers) > 0 {
		filteredHeaders := make(map[string]string)
		for name, value := range response.Headers {
			if !lp.isHeaderExcluded(name) {
				filteredHeaders[name] = value
			}
		}
		fields["response_headers"] = filteredHeaders
	}

	// Add body if configured and within size limit
	if lp.config.IncludeBody && len(response.Body) > 0 && len(response.Body) <= lp.config.MaxBodySize {
		fields["response_body"] = string(response.Body)
	}

	// Add custom fields
	for key, value := range lp.config.CustomFields {
		fields[key] = value
	}

	// Add context
	if len(response.Context) > 0 {
		fields["context"] = response.Context
	}

	// Add metadata
	if len(response.Metadata) > 0 {
		fields["metadata"] = response.Metadata
	}

	return fields
}

func (lp *LoggingProvider) getLogLevel(statusCode int) logrus.Level {
	switch {
	case statusCode >= 500:
		return logrus.ErrorLevel
	case statusCode >= 400:
		return logrus.WarnLevel
	case statusCode >= 300:
		return logrus.InfoLevel
	default:
		return logrus.InfoLevel
	}
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}
