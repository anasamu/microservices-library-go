package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit/types"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Provider implements rate limiting using Redis
type Provider struct {
	client    *redis.Client
	config    *Config
	logger    *logrus.Logger
	connected bool
}

// Config holds Redis provider configuration
type Config struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     int    `json:"database"`
	Password     string `json:"password"`
	KeyPrefix    string `json:"key_prefix"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

// NewProvider creates a new Redis rate limit provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			KeyPrefix:    "ratelimit:",
			PoolSize:     10,
			MinIdleConns: 5,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	provider := &Provider{
		client:    client,
		config:    config,
		logger:    logrus.New(),
		connected: false,
	}

	return provider, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "redis"
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
		types.FeaturePattern,
		types.FeaturePersistence,
		types.FeatureClustering,
		types.FeatureSliding,
		types.FeatureTokenBucket,
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
		Host:     p.config.Host,
		Port:     p.config.Port,
		Database: strconv.Itoa(p.config.Database),
		Status:   status,
		Metadata: map[string]string{
			"key_prefix": p.config.KeyPrefix,
		},
	}
}

// Connect establishes connection to Redis
func (p *Provider) Connect(ctx context.Context) error {
	if err := p.client.Ping(ctx).Err(); err != nil {
		p.connected = false
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	p.connected = true
	p.logger.Info("Connected to Redis")
	return nil
}

// Disconnect closes the Redis connection
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.client != nil {
		err := p.client.Close()
		p.connected = false
		p.logger.Info("Disconnected from Redis")
		return err
	}
	return nil
}

// Ping checks if Redis is reachable
func (p *Provider) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

// IsConnected returns connection status
func (p *Provider) IsConnected() bool {
	return p.connected
}

// Allow checks if a request is allowed
func (p *Provider) Allow(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
	if !p.connected {
		if err := p.Connect(ctx); err != nil {
			return nil, err
		}
	}

	fullKey := p.config.KeyPrefix + key
	now := time.Now()

	switch limit.Algorithm {
	case types.AlgorithmTokenBucket:
		return p.allowTokenBucket(ctx, fullKey, limit, now)
	case types.AlgorithmSlidingWindow:
		return p.allowSlidingWindow(ctx, fullKey, limit, now)
	case types.AlgorithmFixedWindow:
		return p.allowFixedWindow(ctx, fullKey, limit, now)
	default:
		return p.allowTokenBucket(ctx, fullKey, limit, now)
	}
}

// allowTokenBucket implements token bucket algorithm
func (p *Provider) allowTokenBucket(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (*types.RateLimitResult, error) {
	// Use Lua script for atomic operations
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or limit
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add based on time passed
		local time_passed = now - last_refill
		local tokens_to_add = math.floor(time_passed / (window / limit))
		tokens = math.min(limit, tokens + tokens_to_add)
		
		local allowed = tokens > 0
		if allowed then
			tokens = tokens - 1
		end
		
		-- Update bucket
		redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
		redis.call('EXPIRE', key, window)
		
		return {allowed and 1 or 0, tokens, now + window}
	`

	result, err := p.client.Eval(ctx, script, []string{key}, limit.Limit, limit.Window.Milliseconds(), now.UnixMilli()).Result()
	if err != nil {
		return nil, fmt.Errorf("token bucket script failed: %w", err)
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remaining := results[1].(int64)
	resetTime := time.UnixMilli(results[2].(int64))

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      limit.Limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		Key:        key,
	}, nil
}

// allowSlidingWindow implements sliding window algorithm
func (p *Provider) allowSlidingWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (*types.RateLimitResult, error) {
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		-- Remove expired entries
		redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
		
		-- Count current entries
		local current = redis.call('ZCARD', key)
		
		local allowed = current < limit
		if allowed then
			-- Add current request
			redis.call('ZADD', key, now, now .. ':' .. math.random())
			redis.call('EXPIRE', key, math.ceil(window / 1000))
		end
		
		return {allowed and 1 or 0, limit - current - (allowed and 1 or 0), now + window}
	`

	result, err := p.client.Eval(ctx, script, []string{key}, limit.Limit, limit.Window.Milliseconds(), now.UnixMilli()).Result()
	if err != nil {
		return nil, fmt.Errorf("sliding window script failed: %w", err)
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remaining := results[1].(int64)
	resetTime := time.UnixMilli(results[2].(int64))

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      limit.Limit,
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

	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		
		local current = redis.call('GET', key)
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end
		
		local allowed = current < limit
		if allowed then
			redis.call('INCR', key)
			redis.call('EXPIRE', key, math.ceil(window / 1000))
		end
		
		return {allowed and 1 or 0, limit - current - (allowed and 1 or 0), tonumber(ARGV[3]) + window}
	`

	result, err := p.client.Eval(ctx, script, []string{windowKey}, limit.Limit, limit.Window.Milliseconds(), windowStart.UnixMilli()).Result()
	if err != nil {
		return nil, fmt.Errorf("fixed window script failed: %w", err)
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remaining := results[1].(int64)
	resetTime := time.UnixMilli(results[2].(int64))

	return &types.RateLimitResult{
		Allowed:    allowed,
		Limit:      limit.Limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		Key:        key,
	}, nil
}

// Reset resets the rate limit for a key
func (p *Provider) Reset(ctx context.Context, key string) error {
	if !p.connected {
		if err := p.Connect(ctx); err != nil {
			return err
		}
	}

	fullKey := p.config.KeyPrefix + key
	return p.client.Del(ctx, fullKey).Err()
}

// GetRemaining returns the remaining requests for a key
func (p *Provider) GetRemaining(ctx context.Context, key string, limit *types.RateLimit) (int64, error) {
	if !p.connected {
		if err := p.Connect(ctx); err != nil {
			return 0, err
		}
	}

	fullKey := p.config.KeyPrefix + key
	now := time.Now()

	switch limit.Algorithm {
	case types.AlgorithmTokenBucket:
		return p.getRemainingTokenBucket(ctx, fullKey, limit, now)
	case types.AlgorithmSlidingWindow:
		return p.getRemainingSlidingWindow(ctx, fullKey, limit, now)
	case types.AlgorithmFixedWindow:
		return p.getRemainingFixedWindow(ctx, fullKey, limit, now)
	default:
		return p.getRemainingTokenBucket(ctx, fullKey, limit, now)
	}
}

// getRemainingTokenBucket gets remaining tokens for token bucket
func (p *Provider) getRemainingTokenBucket(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	bucket, err := p.client.HMGet(ctx, key, "tokens", "last_refill").Result()
	if err != nil {
		return 0, err
	}

	if bucket[0] == nil {
		return limit.Limit, nil
	}

	tokens, _ := strconv.ParseInt(bucket[0].(string), 10, 64)
	lastRefill, _ := strconv.ParseInt(bucket[1].(string), 10, 64)

	// Calculate tokens to add based on time passed
	timePassed := now.UnixMilli() - lastRefill
	tokensToAdd := timePassed / (limit.Window.Milliseconds() / limit.Limit)
	tokens = min(limit.Limit, tokens+tokensToAdd)

	return tokens, nil
}

// getRemainingSlidingWindow gets remaining requests for sliding window
func (p *Provider) getRemainingSlidingWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	// Remove expired entries
	p.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.UnixMilli()-limit.Window.Milliseconds(), 10))

	// Count current entries
	current, err := p.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	remaining := limit.Limit - current
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// getRemainingFixedWindow gets remaining requests for fixed window
func (p *Provider) getRemainingFixedWindow(ctx context.Context, key string, limit *types.RateLimit, now time.Time) (int64, error) {
	windowStart := now.Truncate(limit.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	current, err := p.client.Get(ctx, windowKey).Int64()
	if err == redis.Nil {
		return limit.Limit, nil
	}
	if err != nil {
		return 0, err
	}

	remaining := limit.Limit - current
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
		// For token bucket, reset time is when the bucket will be full
		return now.Add(limit.Window), nil
	case types.AlgorithmSlidingWindow:
		// For sliding window, reset time is when the oldest entry expires
		fullKey := p.config.KeyPrefix + key
		oldest, err := p.client.ZRange(ctx, fullKey, 0, 0).Result()
		if err != nil || len(oldest) == 0 {
			return now, nil
		}
		// Parse timestamp from the oldest entry
		if len(oldest[0]) > 13 {
			timestamp, _ := strconv.ParseInt(oldest[0][:13], 10, 64)
			return time.UnixMilli(timestamp).Add(limit.Window), nil
		}
		return now, nil
	case types.AlgorithmFixedWindow:
		// For fixed window, reset time is when the current window ends
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
		if err := p.Connect(ctx); err != nil {
			return err
		}
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = p.config.KeyPrefix + key
	}

	return p.client.Del(ctx, fullKeys...).Err()
}

// GetStats returns rate limit statistics
func (p *Provider) GetStats(ctx context.Context) (*types.RateLimitStats, error) {
	if !p.connected {
		if err := p.Connect(ctx); err != nil {
			return nil, err
		}
	}

	info, err := p.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse Redis info to extract basic stats
	stats := &types.RateLimitStats{
		Provider:   "redis",
		LastUpdate: time.Now(),
	}

	// Count keys with our prefix
	keys, err := p.client.Keys(ctx, p.config.KeyPrefix+"*").Result()
	if err == nil {
		stats.ActiveKeys = int64(len(keys))
	}

	return stats, nil
}

// GetKeys returns keys matching a pattern
func (p *Provider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	if !p.connected {
		if err := p.Connect(ctx); err != nil {
			return nil, err
		}
	}

	fullPattern := p.config.KeyPrefix + pattern
	keys, err := p.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, err
	}

	// Remove prefix from keys
	for i, key := range keys {
		keys[i] = key[len(p.config.KeyPrefix):]
	}

	return keys, nil
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
