package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/cache/v9"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CacheManager manages different types of caches
type CacheManager struct {
	redis    *redis.Client
	memcache *memcache.Client
	memory   *cache.Cache
	config   *CacheConfig
	logger   *logrus.Logger
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Redis      RedisConfig    `json:"redis"`
	Memcache   MemcacheConfig `json:"memcache"`
	Memory     MemoryConfig   `json:"memory"`
	DefaultTTL time.Duration  `json:"default_ttl"`
	Enabled    bool           `json:"enabled"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
	Enabled  bool   `json:"enabled"`
}

// MemcacheConfig holds Memcache configuration
type MemcacheConfig struct {
	Servers []string `json:"servers"`
	Enabled bool     `json:"enabled"`
}

// MemoryConfig holds in-memory cache configuration
type MemoryConfig struct {
	DefaultExpiration time.Duration `json:"default_expiration"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	Enabled           bool          `json:"enabled"`
}

// CacheItem represents a cached item
type CacheItem struct {
	Key        string                 `json:"key"`
	Value      interface{}            `json:"value"`
	Expiration time.Time              `json:"expiration"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits       int64         `json:"hits"`
	Misses     int64         `json:"misses"`
	Keys       int64         `json:"keys"`
	Memory     int64         `json:"memory"`
	Uptime     time.Duration `json:"uptime"`
	LastUpdate time.Time     `json:"last_update"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *CacheConfig, logger *logrus.Logger) (*CacheManager, error) {
	// Set default values
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}

	manager := &CacheManager{
		config: config,
		logger: logger,
	}

	// Initialize Redis if enabled
	if config.Redis.Enabled {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
			PoolSize: config.Redis.PoolSize,
		})

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}

		manager.redis = redisClient
		logger.Info("Redis cache initialized successfully")
	}

	// Initialize Memcache if enabled
	if config.Memcache.Enabled {
		memcacheClient := memcache.New(config.Memcache.Servers...)
		manager.memcache = memcacheClient
		logger.Info("Memcache initialized successfully")
	}

	// Initialize in-memory cache if enabled
	if config.Memory.Enabled {
		if config.Memory.DefaultExpiration == 0 {
			config.Memory.DefaultExpiration = 5 * time.Minute
		}
		if config.Memory.CleanupInterval == 0 {
			config.Memory.CleanupInterval = 10 * time.Minute
		}

		manager.memory = cache.New(config.Memory.DefaultExpiration, config.Memory.CleanupInterval)
		logger.Info("In-memory cache initialized successfully")
	}

	logger.WithFields(logrus.Fields{
		"redis_enabled":    config.Redis.Enabled,
		"memcache_enabled": config.Memcache.Enabled,
		"memory_enabled":   config.Memory.Enabled,
		"default_ttl":      config.DefaultTTL,
	}).Info("Cache manager initialized successfully")

	return manager, nil
}

// Set stores a value in the cache
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cm.config.Enabled {
		return nil
	}

	if ttl == 0 {
		ttl = cm.config.DefaultTTL
	}

	// Try Redis first
	if cm.redis != nil {
		if err := cm.setRedis(ctx, key, value, ttl); err != nil {
			cm.logger.WithError(err).WithField("key", key).Warn("Failed to set value in Redis")
		} else {
			cm.logger.WithField("key", key).Debug("Value set in Redis cache")
			return nil
		}
	}

	// Try Memcache
	if cm.memcache != nil {
		if err := cm.setMemcache(ctx, key, value, ttl); err != nil {
			cm.logger.WithError(err).WithField("key", key).Warn("Failed to set value in Memcache")
		} else {
			cm.logger.WithField("key", key).Debug("Value set in Memcache")
			return nil
		}
	}

	// Fallback to in-memory cache
	if cm.memory != nil {
		cm.setMemory(key, value, ttl)
		cm.logger.WithField("key", key).Debug("Value set in memory cache")
		return nil
	}

	return fmt.Errorf("no cache backend available")
}

// Get retrieves a value from the cache
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	if !cm.config.Enabled {
		return fmt.Errorf("cache not enabled")
	}

	// Try Redis first
	if cm.redis != nil {
		if err := cm.getRedis(ctx, key, dest); err == nil {
			cm.logger.WithField("key", key).Debug("Value retrieved from Redis cache")
			return nil
		}
	}

	// Try Memcache
	if cm.memcache != nil {
		if err := cm.getMemcache(ctx, key, dest); err == nil {
			cm.logger.WithField("key", key).Debug("Value retrieved from Memcache")
			return nil
		}
	}

	// Try in-memory cache
	if cm.memory != nil {
		if err := cm.getMemory(key, dest); err == nil {
			cm.logger.WithField("key", key).Debug("Value retrieved from memory cache")
			return nil
		}
	}

	return fmt.Errorf("key not found: %s", key)
}

// Delete removes a value from the cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	if !cm.config.Enabled {
		return nil
	}

	var lastErr error

	// Delete from Redis
	if cm.redis != nil {
		if err := cm.redis.Del(ctx, key).Err(); err != nil {
			cm.logger.WithError(err).WithField("key", key).Warn("Failed to delete from Redis")
			lastErr = err
		}
	}

	// Delete from Memcache
	if cm.memcache != nil {
		if err := cm.memcache.Delete(key); err != nil {
			cm.logger.WithError(err).WithField("key", key).Warn("Failed to delete from Memcache")
			lastErr = err
		}
	}

	// Delete from in-memory cache
	if cm.memory != nil {
		cm.memory.Delete(key)
	}

	cm.logger.WithField("key", key).Debug("Value deleted from cache")
	return lastErr
}

// Exists checks if a key exists in the cache
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	if !cm.config.Enabled {
		return false, nil
	}

	// Check Redis first
	if cm.redis != nil {
		exists, err := cm.redis.Exists(ctx, key).Result()
		if err == nil && exists > 0 {
			return true, nil
		}
	}

	// Check Memcache
	if cm.memcache != nil {
		_, err := cm.memcache.Get(key)
		if err == nil {
			return true, nil
		}
	}

	// Check in-memory cache
	if cm.memory != nil {
		_, exists := cm.memory.Get(key)
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// SetWithTags stores a value with tags for easier invalidation
func (cm *CacheManager) SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	if !cm.config.Enabled {
		return nil
	}

	// Store the value
	if err := cm.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Store tags
	if cm.redis != nil {
		for _, tag := range tags {
			tagKey := fmt.Sprintf("tag:%s", tag)
			cm.redis.SAdd(ctx, tagKey, key)
			cm.redis.Expire(ctx, tagKey, ttl)
		}
	}

	return nil
}

// InvalidateByTag invalidates all keys with a specific tag
func (cm *CacheManager) InvalidateByTag(ctx context.Context, tag string) error {
	if !cm.config.Enabled {
		return nil
	}

	if cm.redis == nil {
		return fmt.Errorf("tag invalidation requires Redis")
	}

	tagKey := fmt.Sprintf("tag:%s", tag)
	keys, err := cm.redis.SMembers(ctx, tagKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for tag: %w", err)
	}

	// Delete all keys with this tag
	for _, key := range keys {
		cm.Delete(ctx, key)
	}

	// Delete the tag itself
	cm.redis.Del(ctx, tagKey)

	cm.logger.WithFields(logrus.Fields{
		"tag":  tag,
		"keys": len(keys),
	}).Info("Invalidated cache by tag")

	return nil
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{
		LastUpdate: time.Now(),
	}

	if cm.redis != nil {
		info, err := cm.redis.Info(ctx, "stats").Result()
		if err == nil {
			// Parse Redis stats (simplified)
			stats.Keys, _ = cm.redis.DBSize(ctx).Result()
		}
	}

	if cm.memory != nil {
		// Get memory cache stats
		items := cm.memory.Items()
		stats.Keys = int64(len(items))
	}

	return stats, nil
}

// Flush clears all cache data
func (cm *CacheManager) Flush(ctx context.Context) error {
	if !cm.config.Enabled {
		return nil
	}

	var lastErr error

	// Flush Redis
	if cm.redis != nil {
		if err := cm.redis.FlushDB(ctx).Err(); err != nil {
			cm.logger.WithError(err).Error("Failed to flush Redis")
			lastErr = err
		}
	}

	// Flush Memcache
	if cm.memcache != nil {
		if err := cm.memcache.FlushAll(); err != nil {
			cm.logger.WithError(err).Error("Failed to flush Memcache")
			lastErr = err
		}
	}

	// Flush in-memory cache
	if cm.memory != nil {
		cm.memory.Flush()
	}

	cm.logger.Info("Cache flushed")
	return lastErr
}

// Close closes the cache manager
func (cm *CacheManager) Close() error {
	if cm.redis != nil {
		return cm.redis.Close()
	}
	return nil
}

// setRedis stores a value in Redis
func (cm *CacheManager) setRedis(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return cm.redis.Set(ctx, key, data, ttl).Err()
}

// getRedis retrieves a value from Redis
func (cm *CacheManager) getRedis(ctx context.Context, key string, dest interface{}) error {
	data, err := cm.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// setMemcache stores a value in Memcache
func (cm *CacheManager) setMemcache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(ttl.Seconds()),
	}

	return cm.memcache.Set(item)
}

// getMemcache retrieves a value from Memcache
func (cm *CacheManager) getMemcache(ctx context.Context, key string, dest interface{}) error {
	item, err := cm.memcache.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(item.Value, dest)
}

// setMemory stores a value in in-memory cache
func (cm *CacheManager) setMemory(key string, value interface{}, ttl time.Duration) {
	cm.memory.Set(key, value, ttl)
}

// getMemory retrieves a value from in-memory cache
func (cm *CacheManager) getMemory(key string, dest interface{}) error {
	value, found := cm.memory.Get(key)
	if !found {
		return fmt.Errorf("key not found: %s", key)
	}

	// Type assertion
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, dest)
	default:
		// Direct assignment if types match
		if destPtr, ok := dest.(*interface{}); ok {
			*destPtr = v
			return nil
		}
		return fmt.Errorf("type mismatch for key: %s", key)
	}
}

// CreateCacheKey creates a standardized cache key
func CreateCacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// CreateUserCacheKey creates a cache key for user data
func CreateUserCacheKey(userID string) string {
	return CreateCacheKey("user", userID)
}

// CreateTenantCacheKey creates a cache key for tenant data
func CreateTenantCacheKey(tenantID string) string {
	return CreateCacheKey("tenant", tenantID)
}

// CreateSessionCacheKey creates a cache key for session data
func CreateSessionCacheKey(sessionID string) string {
	return CreateCacheKey("session", sessionID)
}

// CreateAPICacheKey creates a cache key for API responses
func CreateAPICacheKey(endpoint string, params map[string]string) string {
	key := CreateCacheKey("api", endpoint)
	for k, v := range params {
		key += ":" + k + "=" + v
	}
	return key
}
