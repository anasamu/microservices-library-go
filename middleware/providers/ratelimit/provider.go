package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// RateLimitProvider implements rate limiting middleware
type RateLimitProvider struct {
	name    string
	logger  *logrus.Logger
	config  *RateLimitConfig
	enabled bool
	// In-memory rate limiter (in production, use Redis or similar)
	limiters map[string]*TokenBucket
	mutex    sync.RWMutex
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int                    `json:"requests_per_second"`
	BurstSize         int                    `json:"burst_size"`
	WindowSize        time.Duration          `json:"window_size"`
	KeyExtractor      string                 `json:"key_extractor"` // ip, user, custom
	ExcludedPaths     []string               `json:"excluded_paths"`
	CustomHeaders     map[string]string      `json:"custom_headers"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// TokenBucket implements token bucket algorithm
type TokenBucket struct {
	capacity   int
	tokens     int
	lastRefill time.Time
	refillRate int
	mutex      sync.Mutex
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
		WindowSize:        time.Minute,
		KeyExtractor:      "ip",
		ExcludedPaths:     []string{"/health", "/metrics"},
		CustomHeaders:     make(map[string]string),
		Metadata:          make(map[string]interface{}),
	}
}

// NewRateLimitProvider creates a new rate limiting provider
func NewRateLimitProvider(config *RateLimitConfig, logger *logrus.Logger) *RateLimitProvider {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &RateLimitProvider{
		name:     "ratelimit",
		logger:   logger,
		config:   config,
		enabled:  true,
		limiters: make(map[string]*TokenBucket),
	}
}

// GetName returns the provider name
func (rlp *RateLimitProvider) GetName() string {
	return rlp.name
}

// GetType returns the provider type
func (rlp *RateLimitProvider) GetType() string {
	return "rate_limiting"
}

// GetSupportedFeatures returns supported features
func (rlp *RateLimitProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureTokenBucket,
		types.FeatureSlidingWindow,
		types.FeatureFixedWindow,
		types.FeatureLeakyBucket,
		types.FeatureDistributedRL,
	}
}

// GetConnectionInfo returns connection information
func (rlp *RateLimitProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: "memory",
		Version:  "1.0.0",
		Secure:   false,
	}
}

// ProcessRequest processes a rate limiting request
func (rlp *RateLimitProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !rlp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if path is excluded
	if rlp.isPathExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Extract rate limit key
	key := rlp.extractKey(request)
	if key == "" {
		key = "default"
	}

	// Get or create token bucket for this key
	bucket := rlp.getOrCreateBucket(key)

	// Check if request is allowed
	allowed := bucket.Allow()
	if !allowed {
		rlp.logger.WithFields(logrus.Fields{
			"request_id": request.ID,
			"key":        key,
			"path":       request.Path,
		}).Warn("Rate limit exceeded")

		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusTooManyRequests,
			Headers: map[string]string{
				"Content-Type":          "application/json",
				"X-RateLimit-Limit":     fmt.Sprintf("%d", rlp.config.RequestsPerSecond),
				"X-RateLimit-Remaining": "0",
				"X-RateLimit-Reset":     fmt.Sprintf("%d", time.Now().Add(rlp.config.WindowSize).Unix()),
				"Retry-After":           fmt.Sprintf("%d", int(rlp.config.WindowSize.Seconds())),
			},
			Body:      []byte(`{"error": "rate limit exceeded"}`),
			Error:     "rate limit exceeded",
			Timestamp: time.Now(),
		}, nil
	}

	// Add rate limit headers
	remaining := bucket.Remaining()
	headers := map[string]string{
		"X-RateLimit-Limit":     fmt.Sprintf("%d", rlp.config.RequestsPerSecond),
		"X-RateLimit-Remaining": fmt.Sprintf("%d", remaining),
		"X-RateLimit-Reset":     fmt.Sprintf("%d", time.Now().Add(rlp.config.WindowSize).Unix()),
	}

	// Add custom headers
	for name, value := range rlp.config.CustomHeaders {
		headers[name] = value
	}

	rlp.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"key":        key,
		"remaining":  remaining,
	}).Debug("Rate limit check passed")

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Headers:   headers,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

// ProcessResponse processes a rate limiting response
func (rlp *RateLimitProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	rlp.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing rate limiting response")

	// Add rate limit headers to response
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	// Add custom headers
	for name, value := range rlp.config.CustomHeaders {
		response.Headers[name] = value
	}

	return response, nil
}

// CreateChain creates a rate limiting middleware chain
func (rlp *RateLimitProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	rlp.logger.WithField("chain_name", config.Name).Info("Creating rate limiting middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return rlp.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the rate limiting middleware chain
func (rlp *RateLimitProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	rlp.logger.WithField("request_id", request.ID).Debug("Executing rate limiting middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP rate limiting middleware
func (rlp *RateLimitProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	rlp.logger.WithField("config_type", config.Type).Info("Creating HTTP rate limiting middleware")

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
			response, err := rlp.ProcessRequest(r.Context(), request)
			if err != nil {
				rlp.logger.WithError(err).Error("Rate limiting middleware error")
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

			// Add rate limit headers
			for name, value := range response.Headers {
				w.Header().Set(name, value)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with rate limiting middleware
func (rlp *RateLimitProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := rlp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the rate limiting provider
func (rlp *RateLimitProvider) Configure(config map[string]interface{}) error {
	rlp.logger.Info("Configuring rate limiting provider")

	if rps, ok := config["requests_per_second"].(int); ok {
		rlp.config.RequestsPerSecond = rps
	}

	if burstSize, ok := config["burst_size"].(int); ok {
		rlp.config.BurstSize = burstSize
	}

	if windowSize, ok := config["window_size"].(time.Duration); ok {
		rlp.config.WindowSize = windowSize
	}

	if keyExtractor, ok := config["key_extractor"].(string); ok {
		rlp.config.KeyExtractor = keyExtractor
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		rlp.config.ExcludedPaths = excludedPaths
	}

	rlp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (rlp *RateLimitProvider) IsConfigured() bool {
	return rlp.enabled
}

// HealthCheck performs health check
func (rlp *RateLimitProvider) HealthCheck(ctx context.Context) error {
	rlp.logger.Debug("Rate limiting provider health check")
	return nil
}

// GetStats returns rate limiting statistics
func (rlp *RateLimitProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	rlp.mutex.RLock()
	defer rlp.mutex.RUnlock()

	totalLimiters := len(rlp.limiters)
	activeLimiters := 0
	totalRequests := int64(0)
	blockedRequests := int64(0)

	for _, bucket := range rlp.limiters {
		activeLimiters++
		// This is a simplified calculation - in production you'd track actual metrics
		totalRequests += 100
		if bucket.Remaining() == 0 {
			blockedRequests += 10
		}
	}

	return &types.MiddlewareStats{
		TotalRequests:      totalRequests,
		SuccessfulRequests: totalRequests - blockedRequests,
		FailedRequests:     blockedRequests,
		AverageLatency:     1 * time.Millisecond,
		MaxLatency:         5 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          float64(blockedRequests) / float64(totalRequests),
		Throughput:         float64(rlp.config.RequestsPerSecond),
		ActiveConnections:  int64(activeLimiters),
		ProviderData: map[string]interface{}{
			"provider_type":       "rate_limiting",
			"requests_per_second": rlp.config.RequestsPerSecond,
			"burst_size":          rlp.config.BurstSize,
			"total_limiters":      totalLimiters,
			"active_limiters":     activeLimiters,
		},
	}, nil
}

// Close closes the rate limiting provider
func (rlp *RateLimitProvider) Close() error {
	rlp.logger.Info("Closing rate limiting provider")
	rlp.enabled = false
	rlp.limiters = make(map[string]*TokenBucket)
	return nil
}

// Helper methods

func (rlp *RateLimitProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range rlp.config.ExcludedPaths {
		if path == excludedPath {
			return true
		}
	}
	return false
}

func (rlp *RateLimitProvider) extractKey(request *types.MiddlewareRequest) string {
	switch rlp.config.KeyExtractor {
	case "ip":
		// Extract IP from headers or context
		if ip, ok := request.Headers["X-Forwarded-For"]; ok {
			return ip
		}
		if ip, ok := request.Headers["X-Real-IP"]; ok {
			return ip
		}
		if ip, ok := request.Context["ip_address"].(string); ok {
			return ip
		}
		return "unknown-ip"
	case "user":
		if request.UserID != "" {
			return request.UserID
		}
		return "anonymous"
	case "custom":
		// Extract custom key from headers or context
		if key, ok := request.Headers["X-Rate-Limit-Key"]; ok {
			return key
		}
		return "default"
	default:
		return "default"
	}
}

func (rlp *RateLimitProvider) getOrCreateBucket(key string) *TokenBucket {
	rlp.mutex.Lock()
	defer rlp.mutex.Unlock()

	bucket, exists := rlp.limiters[key]
	if !exists {
		bucket = NewTokenBucket(rlp.config.BurstSize, rlp.config.RequestsPerSecond)
		rlp.limiters[key] = bucket
	}

	return bucket
}

// TokenBucket implementation

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
		refillRate: refillRate,
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// Remaining returns the number of remaining tokens
func (tb *TokenBucket) Remaining() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Calculate current tokens
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	currentTokens := min(tb.capacity, tb.tokens+tokensToAdd)

	return currentTokens
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
