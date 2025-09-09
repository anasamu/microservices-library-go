package types

import (
	"time"
)

// FailoverFeature represents a failover feature
type FailoverFeature string

const (
	// Basic features
	FeatureHealthCheck    FailoverFeature = "health_check"
	FeatureLoadBalancing  FailoverFeature = "load_balancing"
	FeatureCircuitBreaker FailoverFeature = "circuit_breaker"
	FeatureRetry          FailoverFeature = "retry"
	FeatureTimeout        FailoverFeature = "timeout"
	FeatureMonitoring     FailoverFeature = "monitoring"

	// Advanced features
	FeatureCustomStrategy    FailoverFeature = "custom_strategy"
	FeatureWeightedRouting   FailoverFeature = "weighted_routing"
	FeatureGeographicRouting FailoverFeature = "geographic_routing"
	FeatureServiceMesh       FailoverFeature = "service_mesh"
	FeatureAutoScaling       FailoverFeature = "auto_scaling"
	FeatureTrafficShifting   FailoverFeature = "traffic_shifting"
)

// FailoverStrategy represents the failover strategy
type FailoverStrategy string

const (
	StrategyRoundRobin  FailoverStrategy = "round_robin"
	StrategyLeastConn   FailoverStrategy = "least_connections"
	StrategyWeighted    FailoverStrategy = "weighted"
	StrategyRandom      FailoverStrategy = "random"
	StrategyIPHash      FailoverStrategy = "ip_hash"
	StrategyGeographic  FailoverStrategy = "geographic"
	StrategyHealthBased FailoverStrategy = "health_based"
	StrategyCustom      FailoverStrategy = "custom"
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnknown   HealthStatus = "unknown"
)

// ConnectionInfo holds connection information for a failover provider
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

// ServiceEndpoint represents a service endpoint
type ServiceEndpoint struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Address      string            `json:"address"`
	Port         int               `json:"port"`
	Protocol     string            `json:"protocol"`
	Health       HealthStatus      `json:"health"`
	Weight       int               `json:"weight"`
	Priority     int               `json:"priority"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]string `json:"metadata"`
	LastCheck    time.Time         `json:"last_check"`
	ResponseTime time.Duration     `json:"response_time"`
	Failures     int               `json:"failures"`
	Successes    int               `json:"successes"`
}

// FailoverConfig holds failover configuration
type FailoverConfig struct {
	Name                string                `json:"name"`
	Strategy            FailoverStrategy      `json:"strategy"`
	HealthCheckInterval time.Duration         `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration         `json:"health_check_timeout"`
	MaxFailures         int                   `json:"max_failures"`
	RecoveryTime        time.Duration         `json:"recovery_time"`
	RetryAttempts       int                   `json:"retry_attempts"`
	RetryDelay          time.Duration         `json:"retry_delay"`
	Timeout             time.Duration         `json:"timeout"`
	CircuitBreaker      *CircuitBreakerConfig `json:"circuit_breaker,omitempty"`
	LoadBalancer        *LoadBalancerConfig   `json:"load_balancer,omitempty"`
	Metadata            map[string]string     `json:"metadata"`
}

// CircuitBreakerConfig holds circuit breaker configuration for failover
type CircuitBreakerConfig struct {
	Enabled          bool          `json:"enabled"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
	MaxRequests      int           `json:"max_requests"`
	Interval         time.Duration `json:"interval"`
}

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	Algorithm         FailoverStrategy `json:"algorithm"`
	StickySession     bool             `json:"sticky_session"`
	SessionTimeout    time.Duration    `json:"session_timeout"`
	MaxConnections    int              `json:"max_connections"`
	ConnectionTimeout time.Duration    `json:"connection_timeout"`
	KeepAliveTimeout  time.Duration    `json:"keep_alive_timeout"`
}

// FailoverStats represents failover statistics
type FailoverStats struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	HealthChecks        int64         `json:"health_checks"`
	Failovers           int64         `json:"failovers"`
	Recoveries          int64         `json:"recoveries"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastFailover        time.Time     `json:"last_failover"`
	LastRecovery        time.Time     `json:"last_recovery"`
	Uptime              time.Duration `json:"uptime"`
	LastUpdate          time.Time     `json:"last_update"`
	Provider            string        `json:"provider"`
}

// ProviderInfo holds information about a failover provider
type ProviderInfo struct {
	Name              string            `json:"name"`
	SupportedFeatures []FailoverFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo   `json:"connection_info"`
	IsConnected       bool              `json:"is_connected"`
}

// FailoverResult represents the result of a failover operation
type FailoverResult struct {
	Endpoint *ServiceEndpoint       `json:"endpoint"`
	Error    error                  `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Attempts int                    `json:"attempts"`
	Fallback bool                   `json:"fallback"`
	Retry    bool                   `json:"retry"`
	Strategy FailoverStrategy       `json:"strategy"`
	Metadata map[string]interface{} `json:"metadata"`
}

// FailoverEvent represents a failover event
type FailoverEvent struct {
	Type      FailoverEventType      `json:"type"`
	Name      string                 `json:"name"`
	Endpoint  *ServiceEndpoint       `json:"endpoint"`
	Error     error                  `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Provider  string                 `json:"provider"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// FailoverEventType represents the type of failover event
type FailoverEventType string

const (
	EventHealthCheck  FailoverEventType = "health_check"
	EventFailover     FailoverEventType = "failover"
	EventRecovery     FailoverEventType = "recovery"
	EventEndpointDown FailoverEventType = "endpoint_down"
	EventEndpointUp   FailoverEventType = "endpoint_up"
	EventCircuitOpen  FailoverEventType = "circuit_open"
	EventCircuitClose FailoverEventType = "circuit_close"
	EventLoadBalance  FailoverEventType = "load_balance"
	EventRetry        FailoverEventType = "retry"
	EventTimeout      FailoverEventType = "timeout"
)

// FailoverError represents a failover-specific error
type FailoverError struct {
	Code     string           `json:"code"`
	Message  string           `json:"message"`
	Strategy FailoverStrategy `json:"strategy,omitempty"`
}

func (e *FailoverError) Error() string {
	return e.Message
}

// Common failover error codes
const (
	ErrCodeNoEndpoints        = "NO_ENDPOINTS"
	ErrCodeAllEndpointsDown   = "ALL_ENDPOINTS_DOWN"
	ErrCodeHealthCheckFailed  = "HEALTH_CHECK_FAILED"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeInvalidConfig      = "INVALID_CONFIG"
	ErrCodeConnection         = "CONNECTION_ERROR"
	ErrCodeCircuitOpen        = "CIRCUIT_OPEN"
	ErrCodeRetryExhausted     = "RETRY_EXHAUSTED"
	ErrCodeInvalidStrategy    = "INVALID_STRATEGY"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Default configurations
var (
	DefaultConfig = &FailoverConfig{
		Strategy:            StrategyRoundRobin,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxFailures:         3,
		RecoveryTime:        60 * time.Second,
		RetryAttempts:       3,
		RetryDelay:          time.Second,
		Timeout:             30 * time.Second,
		CircuitBreaker: &CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          60 * time.Second,
			MaxRequests:      10,
			Interval:         10 * time.Second,
		},
		LoadBalancer: &LoadBalancerConfig{
			Algorithm:         StrategyRoundRobin,
			StickySession:     false,
			SessionTimeout:    30 * time.Minute,
			MaxConnections:    1000,
			ConnectionTimeout: 10 * time.Second,
			KeepAliveTimeout:  30 * time.Second,
		},
	}
)
