package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/cache/types"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisProvider implements the CacheProvider interface for Redis
type RedisProvider struct {
	client    *redis.Client
	config    *RedisConfig
	logger    *logrus.Logger
	connected bool
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	KeyPrefix    string        `json:"key_prefix"`
	Namespace    string        `json:"namespace"`
}

// NewRedisProvider creates a new Redis cache provider
func NewRedisProvider(config *RedisConfig, logger *logrus.Logger) *RedisProvider {
	if config == nil {
		config = &RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     10,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}
	}

	return &RedisProvider{
		config:    config,
		logger:    logger,
		connected: false,
	}
}

// GetName returns the provider name
func (r *RedisProvider) GetName() string {
	return "redis"
}

// GetSupportedFeatures returns the features supported by Redis
func (r *RedisProvider) GetSupportedFeatures() []types.CacheFeature {
	return []types.CacheFeature{
		types.FeatureSet,
		types.FeatureGet,
		types.FeatureDelete,
		types.FeatureExists,
		types.FeatureFlush,
		types.FeatureStats,
		types.FeatureTags,
		types.FeatureTTL,
		types.FeatureBatch,
		types.FeaturePattern,
		types.FeaturePersistence,
		types.FeaturePubSub,
		types.FeatureLuaScripts,
	}
}

// GetConnectionInfo returns connection information
func (r *RedisProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if r.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     r.config.Host,
		Port:     r.config.Port,
		Database: strconv.Itoa(r.config.DB),
		Status:   status,
		Metadata: map[string]string{
			"pool_size":      strconv.Itoa(r.config.PoolSize),
			"min_idle_conns": strconv.Itoa(r.config.MinIdleConns),
			"key_prefix":     r.config.KeyPrefix,
			"namespace":      r.config.Namespace,
		},
	}
}

// Connect establishes connection to Redis
func (r *RedisProvider) Connect(ctx context.Context) error {
	r.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password:     r.config.Password,
		DB:           r.config.DB,
		PoolSize:     r.config.PoolSize,
		MinIdleConns: r.config.MinIdleConns,
		MaxRetries:   r.config.MaxRetries,
		DialTimeout:  r.config.DialTimeout,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	})

	// Test connection
	if err := r.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.connected = true
	r.logger.Info("Connected to Redis successfully")
	return nil
}

// Disconnect closes the Redis connection
func (r *RedisProvider) Disconnect(ctx context.Context) error {
	if r.client != nil {
		err := r.client.Close()
		r.connected = false
		r.logger.Info("Disconnected from Redis")
		return err
	}
	return nil
}

// Ping tests the Redis connection
func (r *RedisProvider) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return r.client.Ping(ctx).Err()
}

// IsConnected returns the connection status
func (r *RedisProvider) IsConnected() bool {
	return r.connected && r.client != nil
}

// Set stores a value in Redis
func (r *RedisProvider) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeSerialization, Message: err.Error(), Key: key}
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	r.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl,
	}).Debug("Value set in Redis")
	return nil
}

// Get retrieves a value from Redis
func (r *RedisProvider) Get(ctx context.Context, key string, dest interface{}) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
		}
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeDeserialization, Message: err.Error(), Key: key}
	}

	r.logger.WithField("key", key).Debug("Value retrieved from Redis")
	return nil
}

// Delete removes a value from Redis
func (r *RedisProvider) Delete(ctx context.Context, key string) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	r.logger.WithField("key", key).Debug("Value deleted from Redis")
	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisProvider) Exists(ctx context.Context, key string) (bool, error) {
	if !r.IsConnected() {
		return false, &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	return count > 0, nil
}

// SetWithTags stores a value with tags for easier invalidation
func (r *RedisProvider) SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	// Store the value
	if err := r.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Store tags
	key = r.buildKey(key)
	pipe := r.client.Pipeline()
	for _, tag := range tags {
		tagKey := r.buildKey(fmt.Sprintf("tag:%s", tag))
		pipe.SAdd(ctx, tagKey, key)
		pipe.Expire(ctx, tagKey, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	r.logger.WithFields(logrus.Fields{
		"key":  key,
		"tags": tags,
	}).Debug("Value set with tags in Redis")
	return nil
}

// InvalidateByTag invalidates all keys with a specific tag
func (r *RedisProvider) InvalidateByTag(ctx context.Context, tag string) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	tagKey := r.buildKey(fmt.Sprintf("tag:%s", tag))
	keys, err := r.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all keys with this tag
	pipe := r.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, tagKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	r.logger.WithFields(logrus.Fields{
		"tag":  tag,
		"keys": len(keys),
	}).Info("Invalidated cache by tag in Redis")
	return nil
}

// GetStats returns Redis cache statistics
func (r *RedisProvider) GetStats(ctx context.Context) (*types.CacheStats, error) {
	if !r.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	info, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	stats := &types.CacheStats{
		Provider:   r.GetName(),
		LastUpdate: time.Now(),
	}

	// Parse Redis info
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "keyspace_hits:") {
			fmt.Sscanf(line, "keyspace_hits:%d", &stats.Hits)
		} else if strings.HasPrefix(line, "keyspace_misses:") {
			fmt.Sscanf(line, "keyspace_misses:%d", &stats.Misses)
		}
	}

	// Get database size
	stats.Keys, _ = r.client.DBSize(ctx).Result()

	return stats, nil
}

// Flush clears all cache data
func (r *RedisProvider) Flush(ctx context.Context) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	r.logger.Info("Redis cache flushed")
	return nil
}

// SetMultiple stores multiple values in Redis
func (r *RedisProvider) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	pipe := r.client.Pipeline()
	for key, value := range items {
		key = r.buildKey(key)
		data, err := json.Marshal(value)
		if err != nil {
			return &types.CacheError{Code: types.ErrCodeSerialization, Message: err.Error(), Key: key}
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	r.logger.WithField("count", len(items)).Debug("Multiple values set in Redis")
	return nil
}

// GetMultiple retrieves multiple values from Redis
func (r *RedisProvider) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if !r.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	// Build keys with prefix
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.buildKey(key)
	}

	values, err := r.client.MGet(ctx, redisKeys...).Result()
	if err != nil {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(value.(string)), &data); err == nil {
				result[keys[i]] = data
			}
		}
	}

	r.logger.WithField("count", len(result)).Debug("Multiple values retrieved from Redis")
	return result, nil
}

// DeleteMultiple removes multiple values from Redis
func (r *RedisProvider) DeleteMultiple(ctx context.Context, keys []string) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	// Build keys with prefix
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = r.buildKey(key)
	}

	err := r.client.Del(ctx, redisKeys...).Err()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	r.logger.WithField("count", len(keys)).Debug("Multiple values deleted from Redis")
	return nil
}

// GetKeys returns keys matching a pattern
func (r *RedisProvider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	if !r.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	pattern = r.buildKey(pattern)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	// Remove prefix from keys
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = r.removeKeyPrefix(key)
	}

	return result, nil
}

// GetTTL returns the TTL of a key
func (r *RedisProvider) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if !r.IsConnected() {
		return 0, &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	return ttl, nil
}

// SetTTL sets the TTL of a key
func (r *RedisProvider) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	if !r.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Redis not connected"}
	}

	key = r.buildKey(key)
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	return nil
}

// buildKey constructs a key with prefix and namespace
func (r *RedisProvider) buildKey(key string) string {
	parts := []string{}

	if r.config.Namespace != "" {
		parts = append(parts, r.config.Namespace)
	}

	if r.config.KeyPrefix != "" {
		parts = append(parts, r.config.KeyPrefix)
	}

	parts = append(parts, key)

	return strings.Join(parts, ":")
}

// removeKeyPrefix removes the prefix from a key
func (r *RedisProvider) removeKeyPrefix(key string) string {
	prefix := r.buildKey("")
	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}
	return key
}
