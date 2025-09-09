package types

import (
	"time"
)

// CircuitBreakerFeature represents a circuit breaker feature
type CircuitBreakerFeature string

const (
	// Basic features
	FeatureExecute  CircuitBreakerFeature = "execute"
	FeatureState    CircuitBreakerFeature = "state"
	FeatureMetrics  CircuitBreakerFeature = "metrics"
	FeatureFallback CircuitBreakerFeature = "fallback"
	FeatureTimeout  CircuitBreakerFeature = "timeout"
	FeatureRetry    CircuitBreakerFeature = "retry"

	// Advanced features
	FeatureCustomState CircuitBreakerFeature = "custom_state"
	FeatureBulkhead    CircuitBreakerFeature = "bulkhead"
	FeatureRateLimit   CircuitBreakerFeature = "rate_limit"
	FeatureHealthCheck CircuitBreakerFeature = "health_check"
	FeatureMonitoring  CircuitBreakerFeature = "monitoring"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"   // Normal operation
	StateOpen     CircuitState = "open"     // Circuit is open, failing fast
	StateHalfOpen CircuitState = "halfopen" // Testing if service is back
)

// ConnectionInfo holds connection information for a circuit breaker provider
type ConnectionInfo struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Database string            `json:"database"`
	Username string            `json:"username"`
	Status   ConnectionStatus  `json:"status"`
	Metadata map[string]string `json:"metadata"`
}

// ConnectionStatus represents the connection status
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusError        ConnectionStatus = "error"
)

// CircuitBreakerStats represents circuit breaker statistics
type CircuitBreakerStats struct {
	Requests             int64         `json:"requests"`
	Successes            int64         `json:"successes"`
	Failures             int64         `json:"failures"`
	Timeouts             int64         `json:"timeouts"`
	Rejects              int64         `json:"rejects"`
	State                CircuitState  `json:"state"`
	LastFailureTime      time.Time     `json:"last_failure_time"`
	LastSuccessTime      time.Time     `json:"last_success_time"`
	ConsecutiveFailures  int64         `json:"consecutive_failures"`
	ConsecutiveSuccesses int64         `json:"consecutive_successes"`
	Uptime               time.Duration `json:"uptime"`
	LastUpdate           time.Time     `json:"last_update"`
	Provider             string        `json:"provider"`
}

// ProviderInfo holds information about a circuit breaker provider
type ProviderInfo struct {
	Name              string                  `json:"name"`
	SupportedFeatures []CircuitBreakerFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo         `json:"connection_info"`
	IsConnected       bool                    `json:"is_connected"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Name                string                                                `json:"name"`
	MaxRequests         uint32                                                `json:"max_requests"`
	Interval            time.Duration                                         `json:"interval"`
	Timeout             time.Duration                                         `json:"timeout"`
	ReadyToTrip         func(counts Counts) bool                              `json:"-"`
	OnStateChange       func(name string, from CircuitState, to CircuitState) `json:"-"`
	IsSuccessful        func(err error) bool                                  `json:"-"`
	MaxConsecutiveFails uint32                                                `json:"max_consecutive_fails"`
	FailureThreshold    float64                                               `json:"failure_threshold"`
	SuccessThreshold    uint32                                                `json:"success_threshold"`
	FallbackEnabled     bool                                                  `json:"fallback_enabled"`
	RetryEnabled        bool                                                  `json:"retry_enabled"`
	RetryAttempts       int                                                   `json:"retry_attempts"`
	RetryDelay          time.Duration                                         `json:"retry_delay"`
	Metadata            map[string]string                                     `json:"metadata"`
}

// Counts holds the number of requests and their outcomes
type Counts struct {
	Requests             uint32    `json:"requests"`
	TotalSuccesses       uint32    `json:"total_successes"`
	TotalFailures        uint32    `json:"total_failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures  uint32    `json:"consecutive_failures"`
	LastFailureTime      time.Time `json:"last_failure_time"`
	LastSuccessTime      time.Time `json:"last_success_time"`
}

// ExecutionResult represents the result of a circuit breaker execution
type ExecutionResult struct {
	Result   interface{}   `json:"result"`
	Error    error         `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	State    CircuitState  `json:"state"`
	Fallback bool          `json:"fallback"`
	Retry    bool          `json:"retry"`
	Attempts int           `json:"attempts"`
}

// CircuitBreakerEvent represents a circuit breaker event
type CircuitBreakerEvent struct {
	Type      CircuitBreakerEventType `json:"type"`
	Name      string                  `json:"name"`
	State     CircuitState            `json:"state"`
	Error     error                   `json:"error,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
	Provider  string                  `json:"provider"`
	Metadata  map[string]interface{}  `json:"metadata"`
}

// CircuitBreakerEventType represents the type of circuit breaker event
type CircuitBreakerEventType string

const (
	EventStateChange CircuitBreakerEventType = "state_change"
	EventRequest     CircuitBreakerEventType = "request"
	EventSuccess     CircuitBreakerEventType = "success"
	EventFailure     CircuitBreakerEventType = "failure"
	EventTimeout     CircuitBreakerEventType = "timeout"
	EventReject      CircuitBreakerEventType = "reject"
	EventFallback    CircuitBreakerEventType = "fallback"
	EventRetry       CircuitBreakerEventType = "retry"
)

// CircuitBreakerError represents a circuit breaker-specific error
type CircuitBreakerError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	State   CircuitState `json:"state,omitempty"`
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}

// Common circuit breaker error codes
const (
	ErrCodeCircuitOpen    = "CIRCUIT_OPEN"
	ErrCodeTimeout        = "TIMEOUT"
	ErrCodeMaxRequests    = "MAX_REQUESTS_EXCEEDED"
	ErrCodeInvalidConfig  = "INVALID_CONFIG"
	ErrCodeConnection     = "CONNECTION_ERROR"
	ErrCodeFallbackFailed = "FALLBACK_FAILED"
	ErrCodeRetryExhausted = "RETRY_EXHAUSTED"
	ErrCodeInvalidState   = "INVALID_STATE"
)

// Default configurations
var (
	DefaultConfig = &CircuitBreakerConfig{
		MaxRequests:         10,
		Interval:            10 * time.Second,
		Timeout:             60 * time.Second,
		MaxConsecutiveFails: 5,
		FailureThreshold:    0.5,
		SuccessThreshold:    3,
		FallbackEnabled:     true,
		RetryEnabled:        true,
		RetryAttempts:       3,
		RetryDelay:          time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
)
