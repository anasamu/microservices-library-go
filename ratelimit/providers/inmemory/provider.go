package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit/types"
	"github.com/sirupsen/logrus"
)

// Provider implements rate limiting using in-memory storage
type Provider struct {
	config         *Config
	logger         *logrus.Logger
	connected      bool
	mu             sync.RWMutex
	buckets        map[string]*TokenBucket
	slidingWindows map[string]*SlidingWindow
	fixedWindows   map[string]*FixedWindow
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
}

// Config holds in-memory provider configuration
type Config struct {
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxBuckets      int           `json:"max_buckets"`
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	Capacity   int64         `json:"capacity"`
	Tokens     int64         `json:"tokens"`
	LastRefill time.Time     `json:"last_refill"`
	RefillRate int64         `json:"refill_rate"`
	Window     time.Duration `json:"window"`
	mu         sync.Mutex
}

// SlidingWindow represents a sliding window for rate limiting
type SlidingWindow struct {
	Limit    int64         `json:"limit"`
	Window   time.Duration `json:"window"`
	Requests []time.Time   `json:"requests"`
	mu       sync.Mutex
}

// FixedWindow represents a fixed window for rate limiting
type FixedWindow struct {
	Limit  int64         `json:"limit"`
	Window time.Duration `json:"window"`
	Start  time.Time     `json:"start"`
	Count  int64         `json:"count"`
	mu     sync.Mutex
}

// NewProvider creates a new in-memory rate limit provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			CleanupInterval: time.Minute,
			MaxBuckets:      10000,
		}
	}

	provider := &Provider{
		config:         config,
		logger:         logrus.New(),
		connected:      true,
		buckets:        make(map[string]*TokenBucket),
		slidingWindows: make(map[string]*SlidingWindow),
		fixedWindows:   make(map[string]*FixedWindow),
		stopCleanup:    make(chan struct{}),
	}

	// Start cleanup routine
	provider.startCleanup()

	return provider, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "inmemory"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []types.RateLimitFeature {
	return []types.RateLimitFeature{
		types.FeatureAllow,
		types.FeatureReset,
		types.FeatureRemaining,
		types.FeatureResetTime,
		types.FeatureStats,
		types.FeatureBatch,
		types.FeatureTokenBucket,
		types.FeatureSlidingWindow,
		types.FeatureFixedWindow,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if p.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:   "localhost",
		Port:   0,
		Status: status,
		Metadata: map[string]string{
			"storage": "in-memory",
		},
	}
}

// Connect establishes connection (no-op for in-memory)
func (p *Provider) Connect(ctx context.Context) error {
	p.connected = true
	p.logger.Info("Connected to in-memory storage")
	return nil
}

// Disconnect closes the provider
func (p *Provider) Disconnect(ctx context.Context) error {
	p.connected = false

	// Stop cleanup routine
	close(p.stopCleanup)
	if p.cleanupTicker != nil {
		p.cleanupTicker.Stop()
	}

	p.logger.Info("Disconnected from in-memory storage")
	return nil
}

// Ping checks if provider is available (no-op for in-memory)
func (p *Provider) Ping(ctx context.Context) error {
	if !p.connected {
		return fmt.Errorf("provider not connected")
	}
	return nil
}

// IsConnected returns connection status
func (p *Provider) IsConnected() bool {
	return p.connected
}

// Allow checks if a request is allowed
func (p *Provider) Allow(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
	if !p.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	now := time.Now()

	switch limit.Algorithm {
	case types.AlgorithmTokenBucket:
		return p.allowTokenBucket(ctx, key, limit, now)
	case types.AlgorithmSlidingWindow:
		return p.allowSlidingWindow(ctx, key, limit, now)
	case types.AlgorithmFixedWindow:
		return p.allowFixedWindow(ctx, key, limit, now)
	default:
		return p.allowTokenBucket(ctx, key, limit, now)
	}
}

// allowTokenBucket implements token bucket algorithm
func (p *Provider) allowTokenBucket(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (*types.RateLimitResult, error) {
	p.mu.Lock()
	bucket, exists := p.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			Capacity:   limit.Limit,
			Tokens:     limit.Limit,
			LastRefill: now,
			RefillRate: limit.Limit,
			Window:     limit.Window,
		}
		p.buckets[key] = bucket
	}
	p.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate tokens to add based on time passed
	timePassed := now.Sub(bucket.LastRefill)
	tokensToAdd := int64(timePassed / (bucket.Window / time.Duration(bucket.RefillRate)))
	bucket.Tokens = min(bucket.Capacity, bucket.Tokens+tokensToAdd)
	bucket.LastRefill = now

	allowed := bucket.Tokens > 0
	if allowed {
		bucket.Tokens--
	}

	resetTime := now.Add(bucket.Window)

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      bucket.Capacity,
		Remaining:  bucket.Tokens,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		Key:        key,
	}, nil
}

// allowSlidingWindow implements sliding window algorithm
func (p *Provider) allowSlidingWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (*types.RateLimitResult, error) {
	p.mu.Lock()
	window, exists := p.slidingWindows[key]
	if !exists {
		window = &SlidingWindow{
			Limit:    limit.Limit,
			Window:   limit.Window,
			Requests: make([]time.Time, 0),
		}
		p.slidingWindows[key] = window
	}
	p.mu.Unlock()

	window.mu.Lock()
	defer window.mu.Unlock()

	// Remove expired requests
	cutoff := now.Add(-window.Window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range window.Requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	window.Requests = validRequests

	allowed := int64(len(window.Requests)) < window.Limit
	if allowed {
		window.Requests = append(window.Requests, now)
	}

	remaining := window.Limit - int64(len(window.Requests))
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (when oldest request expires)
	var resetTime time.Time
	if len(window.Requests) > 0 {
		resetTime = window.Requests[0].Add(window.Window)
	} else {
		resetTime = now.Add(window.Window)
	}

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      window.Limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		Key:        key,
	}, nil
}

// allowFixedWindow implements fixed window algorithm
func (p *Provider) allowFixedWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (*types.RateLimitResult, error) {
	// Calculate window start
	windowStart := now.Truncate(limit.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	p.mu.Lock()
	window, exists := p.fixedWindows[windowKey]
	if !exists {
		window = &FixedWindow{
			Limit:  limit.Limit,
			Window: limit.Window,
			Start:  windowStart,
			Count:  0,
		}
		p.fixedWindows[windowKey] = window
	}
	p.mu.Unlock()

	window.mu.Lock()
	defer window.mu.Unlock()

	allowed := window.Count < window.Limit
	if allowed {
		window.Count++
	}

	remaining := window.Limit - window.Count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := windowStart.Add(window.Window)

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      window.Limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		Key:        key,
	}, nil
}

// Reset resets the rate limit for a key
func (p *Provider) Reset(ctx context.Context, key string) error {
	if !p.connected {
		return fmt.Errorf("provider not connected")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.buckets, key)
	delete(p.slidingWindows, key)

	// Remove all fixed windows for this key
	for windowKey := range p.fixedWindows {
		if len(windowKey) > len(key) && windowKey[:len(key)] == key {
			delete(p.fixedWindows, windowKey)
		}
	}

	return nil
}

// GetRemaining returns the remaining requests for a key
func (p *Provider) GetRemaining(ctx context.Context, key string, limit *types.RateLimit) (int64, error) {
	if !p.connected {
		return 0, fmt.Errorf("provider not connected")
	}

	now := time.Now()

	switch limit.Algorithm {
	case types.AlgorithmTokenBucket:
		return p.getRemainingTokenBucket(ctx, key, limit, now)
	case types.AlgorithmSlidingWindow:
		return p.getRemainingSlidingWindow(ctx, key, limit, now)
	case types.AlgorithmFixedWindow:
		return p.getRemainingFixedWindow(ctx, key, limit, now)
	default:
		return p.getRemainingTokenBucket(ctx, key, limit, now)
	}
}

// getRemainingTokenBucket gets remaining tokens for token bucket
func (p *Provider) getRemainingTokenBucket(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	p.mu.RLock()
	bucket, exists := p.buckets[key]
	p.mu.RUnlock()

	if !exists {
		return limit.Limit, nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate tokens to add based on time passed
	timePassed := now.Sub(bucket.LastRefill)
	tokensToAdd := int64(timePassed / (bucket.Window / time.Duration(bucket.RefillRate)))
	tokens := min(bucket.Capacity, bucket.Tokens+tokensToAdd)

	return tokens, nil
}

// getRemainingSlidingWindow gets remaining requests for sliding window
func (p *Provider) getRemainingSlidingWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	p.mu.RLock()
	window, exists := p.slidingWindows[key]
	p.mu.RUnlock()

	if !exists {
		return limit.Limit, nil
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	// Remove expired requests
	cutoff := now.Add(-window.Window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range window.Requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	remaining := window.Limit - int64(len(validRequests))
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// getRemainingFixedWindow gets remaining requests for fixed window
func (p *Provider) getRemainingFixedWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	windowStart := now.Truncate(limit.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	p.mu.RLock()
	window, exists := p.fixedWindows[windowKey]
	p.mu.RUnlock()

	if !exists {
		return limit.Limit, nil
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	remaining := window.Limit - window.Count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// GetResetTime returns the reset time for a key
func (p *Provider) GetResetTime(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error) {
	now := time.Now()

	switch limit.Algorithm {
	case types.AlgorithmTokenBucket:
		return now.Add(limit.Window), nil
	case types.AlgorithmSlidingWindow:
		p.mu.RLock()
		window, exists := p.slidingWindows[key]
		p.mu.RUnlock()

		if !exists || len(window.Requests) == 0 {
			return now, nil
		}

		window.mu.Lock()
		defer window.mu.Unlock()

		if len(window.Requests) > 0 {
			return window.Requests[0].Add(window.Window), nil
		}
		return now, nil
	case types.AlgorithmFixedWindow:
		windowStart := now.Truncate(limit.Window)
		return windowStart.Add(limit.Window), nil
	default:
		return now.Add(limit.Window), nil
	}
}

// AllowMultiple checks multiple rate limit requests
func (p *Provider) AllowMultiple(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error) {
	results := make([]*types.RateLimitResult, len(requests))

	for i, req := range requests {
		result, err := p.Allow(ctx, req.Key, req.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to check rate limit for key %s: %w", req.Key, err)
		}
		results[i] = result
	}

	return results, nil
}

// ResetMultiple resets multiple rate limits
func (p *Provider) ResetMultiple(ctx context.Context, keys []string) error {
	if !p.connected {
		return fmt.Errorf("provider not connected")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, key := range keys {
		delete(p.buckets, key)
		delete(p.slidingWindows, key)

		// Remove all fixed windows for this key
		for windowKey := range p.fixedWindows {
			if len(windowKey) > len(key) && windowKey[:len(key)] == key {
				delete(p.fixedWindows, windowKey)
			}
		}
	}

	return nil
}

// GetStats returns rate limit statistics
func (p *Provider) GetStats(ctx context.Context) (*types.RateLimitStats, error) {
	if !p.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &types.RateLimitStats{
		Provider:   "inmemory",
		LastUpdate: time.Now(),
		ActiveKeys: int64(len(p.buckets) + len(p.slidingWindows) + len(p.fixedWindows)),
	}

	return stats, nil
}

// GetKeys returns keys matching a pattern
func (p *Provider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	if !p.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0)

	// Add bucket keys
	for key := range p.buckets {
		if pattern == "" || matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	// Add sliding window keys
	for key := range p.slidingWindows {
		if pattern == "" || matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	// Add fixed window keys (without timestamp suffix)
	seen := make(map[string]bool)
	for windowKey := range p.fixedWindows {
		// Extract base key (remove timestamp suffix)
		baseKey := windowKey
		if idx := len(windowKey) - 11; idx > 0 && windowKey[idx-1] == ':' {
			baseKey = windowKey[:idx-1]
		}

		if !seen[baseKey] && (pattern == "" || matchPattern(baseKey, pattern)) {
			keys = append(keys, baseKey)
			seen[baseKey] = true
		}
	}

	return keys, nil
}

// startCleanup starts the cleanup routine
func (p *Provider) startCleanup() {
	p.cleanupTicker = time.NewTicker(p.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-p.cleanupTicker.C:
				p.cleanup()
			case <-p.stopCleanup:
				return
			}
		}
	}()
}

// cleanup removes expired entries
func (p *Provider) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	// Clean up sliding windows
	for key, window := range p.slidingWindows {
		window.mu.Lock()
		cutoff := now.Add(-window.Window)
		validRequests := make([]time.Time, 0)
		for _, reqTime := range window.Requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		window.Requests = validRequests

		// Remove empty windows
		if len(window.Requests) == 0 {
			delete(p.slidingWindows, key)
		}
		window.mu.Unlock()
	}

	// Clean up fixed windows (remove old windows)
	for windowKey, window := range p.fixedWindows {
		window.mu.Lock()
		if now.After(window.Start.Add(window.Window)) {
			delete(p.fixedWindows, windowKey)
		}
		window.mu.Unlock()
	}

	// Limit total number of buckets
	if len(p.buckets) > p.config.MaxBuckets {
		// Remove oldest buckets (simple FIFO)
		count := 0
		for key := range p.buckets {
			if count >= len(p.buckets)-p.config.MaxBuckets {
				break
			}
			delete(p.buckets, key)
			count++
		}
	}
}

// matchPattern checks if a key matches a pattern (simple wildcard matching)
func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple prefix matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return key == pattern
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
