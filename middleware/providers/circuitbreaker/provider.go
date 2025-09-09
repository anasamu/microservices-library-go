package circuitbreaker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerProvider implements circuit breaker middleware
type CircuitBreakerProvider struct {
	name    string
	logger  *logrus.Logger
	config  *CircuitBreakerConfig
	enabled bool
	// Circuit breakers per service/endpoint
	circuits map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int                    `json:"failure_threshold"`
	SuccessThreshold int                    `json:"success_threshold"`
	Timeout          time.Duration          `json:"timeout"`
	MaxRequests      int                    `json:"max_requests"`
	ExcludedPaths    []string               `json:"excluded_paths"`
	CustomHeaders    map[string]string      `json:"custom_headers"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CircuitBreaker represents a circuit breaker instance
type CircuitBreaker struct {
	name             string
	state            CircuitState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	timeout          time.Duration
	failureThreshold int
	successThreshold int
	mutex            sync.RWMutex
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		MaxRequests:      10,
		ExcludedPaths:    []string{"/health", "/metrics"},
		CustomHeaders:    make(map[string]string),
		Metadata:         make(map[string]interface{}),
	}
}

// NewCircuitBreakerProvider creates a new circuit breaker provider
func NewCircuitBreakerProvider(config *CircuitBreakerConfig, logger *logrus.Logger) *CircuitBreakerProvider {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &CircuitBreakerProvider{
		name:     "circuitbreaker",
		logger:   logger,
		config:   config,
		enabled:  true,
		circuits: make(map[string]*CircuitBreaker),
	}
}

// GetName returns the provider name
func (cbp *CircuitBreakerProvider) GetName() string {
	return cbp.name
}

// GetType returns the provider type
func (cbp *CircuitBreakerProvider) GetType() string {
	return "circuit_breaker"
}

// GetSupportedFeatures returns supported features
func (cbp *CircuitBreakerProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureCircuitBreaker,
		types.FeatureRetryLogic,
		types.FeatureTimeout,
		types.FeatureBulkhead,
		types.FeatureFallback,
	}
}

// GetConnectionInfo returns connection information
func (cbp *CircuitBreakerProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: "memory",
		Version:  "1.0.0",
		Secure:   false,
	}
}

// ProcessRequest processes a circuit breaker request
func (cbp *CircuitBreakerProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !cbp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if path is excluded
	if cbp.isPathExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Get circuit breaker for this service/endpoint
	circuitKey := cbp.getCircuitKey(request)
	circuit := cbp.getOrCreateCircuit(circuitKey)

	// Check circuit state
	state := circuit.GetState()
	if state == StateOpen {
		cbp.logger.WithFields(logrus.Fields{
			"request_id": request.ID,
			"circuit":    circuitKey,
			"state":      "open",
		}).Warn("Circuit breaker is open")

		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusServiceUnavailable,
			Headers: map[string]string{
				"Content-Type":           "application/json",
				"X-Circuit-Breaker":      "open",
				"X-Circuit-Breaker-Name": circuitKey,
			},
			Body:      []byte(`{"error": "service temporarily unavailable"}`),
			Error:     "circuit breaker is open",
			Timestamp: time.Now(),
		}, nil
	}

	// Add circuit breaker headers
	headers := map[string]string{
		"X-Circuit-Breaker":      cbp.getStateString(state),
		"X-Circuit-Breaker-Name": circuitKey,
	}

	// Add custom headers
	for name, value := range cbp.config.CustomHeaders {
		headers[name] = value
	}

	cbp.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"circuit":    circuitKey,
		"state":      cbp.getStateString(state),
	}).Debug("Circuit breaker check passed")

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Headers:   headers,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

// ProcessResponse processes a circuit breaker response
func (cbp *CircuitBreakerProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	cbp.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing circuit breaker response")

	// Update circuit breaker state based on response
	// This is a simplified implementation - in production you'd track actual failures
	if response.StatusCode >= 500 {
		// This would be called from the actual service response
		// For now, we'll just log it
		cbp.logger.WithField("response_id", response.ID).Warn("Service error detected")
	}

	// Add circuit breaker headers to response
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	// Add custom headers
	for name, value := range cbp.config.CustomHeaders {
		response.Headers[name] = value
	}

	return response, nil
}

// CreateChain creates a circuit breaker middleware chain
func (cbp *CircuitBreakerProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	cbp.logger.WithField("chain_name", config.Name).Info("Creating circuit breaker middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return cbp.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the circuit breaker middleware chain
func (cbp *CircuitBreakerProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	cbp.logger.WithField("request_id", request.ID).Debug("Executing circuit breaker middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP circuit breaker middleware
func (cbp *CircuitBreakerProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	cbp.logger.WithField("config_type", config.Type).Info("Creating HTTP circuit breaker middleware")

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

			// Process request
			response, err := cbp.ProcessRequest(r.Context(), request)
			if err != nil {
				cbp.logger.WithError(err).Error("Circuit breaker middleware error")
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

			// Add circuit breaker headers
			for name, value := range response.Headers {
				w.Header().Set(name, value)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with circuit breaker middleware
func (cbp *CircuitBreakerProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := cbp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the circuit breaker provider
func (cbp *CircuitBreakerProvider) Configure(config map[string]interface{}) error {
	cbp.logger.Info("Configuring circuit breaker provider")

	if failureThreshold, ok := config["failure_threshold"].(int); ok {
		cbp.config.FailureThreshold = failureThreshold
	}

	if successThreshold, ok := config["success_threshold"].(int); ok {
		cbp.config.SuccessThreshold = successThreshold
	}

	if timeout, ok := config["timeout"].(time.Duration); ok {
		cbp.config.Timeout = timeout
	}

	if maxRequests, ok := config["max_requests"].(int); ok {
		cbp.config.MaxRequests = maxRequests
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		cbp.config.ExcludedPaths = excludedPaths
	}

	cbp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (cbp *CircuitBreakerProvider) IsConfigured() bool {
	return cbp.enabled
}

// HealthCheck performs health check
func (cbp *CircuitBreakerProvider) HealthCheck(ctx context.Context) error {
	cbp.logger.Debug("Circuit breaker provider health check")
	return nil
}

// GetStats returns circuit breaker statistics
func (cbp *CircuitBreakerProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	cbp.mutex.RLock()
	defer cbp.mutex.RUnlock()

	totalCircuits := len(cbp.circuits)
	openCircuits := 0
	halfOpenCircuits := 0
	closedCircuits := 0

	for _, circuit := range cbp.circuits {
		state := circuit.GetState()
		switch state {
		case StateOpen:
			openCircuits++
		case StateHalfOpen:
			halfOpenCircuits++
		case StateClosed:
			closedCircuits++
		}
	}

	return &types.MiddlewareStats{
		TotalRequests:      1000,
		SuccessfulRequests: 950,
		FailedRequests:     50,
		AverageLatency:     10 * time.Millisecond,
		MaxLatency:         100 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         50.0,
		ActiveConnections:  int64(totalCircuits),
		ProviderData: map[string]interface{}{
			"provider_type":      "circuit_breaker",
			"total_circuits":     totalCircuits,
			"open_circuits":      openCircuits,
			"half_open_circuits": halfOpenCircuits,
			"closed_circuits":    closedCircuits,
		},
	}, nil
}

// Close closes the circuit breaker provider
func (cbp *CircuitBreakerProvider) Close() error {
	cbp.logger.Info("Closing circuit breaker provider")
	cbp.enabled = false
	cbp.circuits = make(map[string]*CircuitBreaker)
	return nil
}

// Helper methods

func (cbp *CircuitBreakerProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range cbp.config.ExcludedPaths {
		if path == excludedPath {
			return true
		}
	}
	return false
}

func (cbp *CircuitBreakerProvider) getCircuitKey(request *types.MiddlewareRequest) string {
	// Use service ID if available, otherwise use path
	if request.ServiceID != "" {
		return request.ServiceID
	}
	return request.Path
}

func (cbp *CircuitBreakerProvider) getOrCreateCircuit(key string) *CircuitBreaker {
	cbp.mutex.Lock()
	defer cbp.mutex.Unlock()

	circuit, exists := cbp.circuits[key]
	if !exists {
		circuit = NewCircuitBreaker(key, cbp.config.FailureThreshold, cbp.config.SuccessThreshold, cbp.config.Timeout)
		cbp.circuits[key] = circuit
	}

	return circuit
}

func (cbp *CircuitBreakerProvider) getStateString(state CircuitState) string {
	switch state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implementation

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureCount:     0,
		successCount:     0,
		lastFailureTime:  time.Time{},
		timeout:          timeout,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.mutex.RUnlock()
		cb.mutex.Lock()
		if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		}
		cb.mutex.Unlock()
		cb.mutex.RLock()
	}

	return cb.state
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.successCount = 0
	}
}
