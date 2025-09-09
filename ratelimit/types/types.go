package types

import (
	"time"
)

// RateLimitFeature represents a rate limiting feature
type RateLimitFeature string

const (
	// Basic features
	FeatureAllow     RateLimitFeature = "allow"
	FeatureReset     RateLimitFeature = "reset"
	FeatureRemaining RateLimitFeature = "remaining"
	FeatureResetTime RateLimitFeature = "reset_time"
	FeatureStats     RateLimitFeature = "stats"

	// Advanced features
	FeatureBatch       RateLimitFeature = "batch"
	FeaturePattern     RateLimitFeature = "pattern"
	FeaturePersistence RateLimitFeature = "persistence"
	FeatureClustering  RateLimitFeature = "clustering"
	FeatureSliding     RateLimitFeature = "sliding_window"
	FeatureTokenBucket RateLimitFeature = "token_bucket"
	FeatureFixedWindow RateLimitFeature = "fixed_window"
)

// ConnectionInfo holds connection information for a rate limit provider
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

// RateLimit represents a rate limit configuration
type RateLimit struct {
	Limit     int64                  `json:"limit"`     // Maximum number of requests
	Window    time.Duration          `json:"window"`    // Time window for the limit
	Algorithm Algorithm              `json:"algorithm"` // Rate limiting algorithm
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Algorithm represents the rate limiting algorithm
type Algorithm string

const (
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmSlidingWindow Algorithm = "sliding_window"
	AlgorithmFixedWindow   Algorithm = "fixed_window"
	AlgorithmLeakyBucket   Algorithm = "leaky_bucket"
)

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool                   `json:"allowed"`               // Whether the request is allowed
	Limit      int64                  `json:"limit"`                 // Total limit
	Remaining  int64                  `json:"remaining"`             // Remaining requests
	ResetTime  time.Time              `json:"reset_time"`            // When the limit resets
	RetryAfter time.Duration          `json:"retry_after,omitempty"` // How long to wait before retry
	Key        string                 `json:"key"`                   // The rate limit key
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RateLimitRequest represents a rate limit request
type RateLimitRequest struct {
	Key   string     `json:"key"`
	Limit *RateLimit `json:"limit"`
}

// RateLimitStats represents rate limit statistics
type RateLimitStats struct {
	TotalRequests   int64         `json:"total_requests"`
	AllowedRequests int64         `json:"allowed_requests"`
	BlockedRequests int64         `json:"blocked_requests"`
	ActiveKeys      int64         `json:"active_keys"`
	Memory          int64         `json:"memory"`
	Uptime          time.Duration `json:"uptime"`
	LastUpdate      time.Time     `json:"last_update"`
	Provider        string        `json:"provider"`
}

// ProviderInfo holds information about a rate limit provider
type ProviderInfo struct {
	Name              string             `json:"name"`
	SupportedFeatures []RateLimitFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo    `json:"connection_info"`
	IsConnected       bool               `json:"is_connected"`
}

// RateLimitConfig holds general rate limit configuration
type RateLimitConfig struct {
	DefaultLimit     int64         `json:"default_limit"`
	DefaultWindow    time.Duration `json:"default_window"`
	DefaultAlgorithm Algorithm     `json:"default_algorithm"`
	KeyPrefix        string        `json:"key_prefix"`
	Namespace        string        `json:"namespace"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
}

// RateLimitEvent represents a rate limit event
type RateLimitEvent struct {
	Type      RateLimitEventType `json:"type"`
	Key       string             `json:"key"`
	Limit     *RateLimit         `json:"limit"`
	Result    *RateLimitResult   `json:"result"`
	Timestamp time.Time          `json:"timestamp"`
	Provider  string             `json:"provider"`
}

// RateLimitEventType represents the type of rate limit event
type RateLimitEventType string

const (
	EventAllow  RateLimitEventType = "allow"
	EventBlock  RateLimitEventType = "block"
	EventReset  RateLimitEventType = "reset"
	EventExpire RateLimitEventType = "expire"
)

// RateLimitError represents a rate limit-specific error
type RateLimitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// Common rate limit error codes
const (
	ErrCodeLimitExceeded   = "LIMIT_EXCEEDED"
	ErrCodeInvalidKey      = "INVALID_KEY"
	ErrCodeInvalidLimit    = "INVALID_LIMIT"
	ErrCodeConnection      = "CONNECTION_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeQuota           = "QUOTA_EXCEEDED"
	ErrCodeSerialization   = "SERIALIZATION_ERROR"
	ErrCodeDeserialization = "DESERIALIZATION_ERROR"
)

// TokenBucketConfig represents token bucket configuration
type TokenBucketConfig struct {
	Capacity     int64         `json:"capacity"`      // Maximum number of tokens
	RefillRate   int64         `json:"refill_rate"`   // Tokens per second
	RefillPeriod time.Duration `json:"refill_period"` // Refill period
}

// SlidingWindowConfig represents sliding window configuration
type SlidingWindowConfig struct {
	WindowSize  time.Duration `json:"window_size"` // Size of the sliding window
	Granularity time.Duration `json:"granularity"` // Granularity of time slots
}

// FixedWindowConfig represents fixed window configuration
type FixedWindowConfig struct {
	WindowSize time.Duration `json:"window_size"` // Size of the fixed window
}

// LeakyBucketConfig represents leaky bucket configuration
type LeakyBucketConfig struct {
	Capacity int64 `json:"capacity"`  // Maximum number of requests
	LeakRate int64 `json:"leak_rate"` // Requests per second that can be processed
}
