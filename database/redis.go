package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Redis represents Redis database connection
type Redis struct {
	Client *redis.Client
	config *RedisConfig
	logger *logrus.Logger
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// NewRedis creates new Redis connection
func NewRedis(config *RedisConfig, logger *logrus.Logger) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r := &Redis{
		Client: rdb,
		config: config,
		logger: logger,
	}

	r.logger.Info("Redis connection established successfully")
	return r, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	if r.Client != nil {
		r.logger.Info("Closing Redis connection")
		return r.Client.Close()
	}
	return nil
}

// Ping checks Redis connection
func (r *Redis) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Set sets a key-value pair with expiration
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// TTL returns time to live for a key
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// HSet sets field in hash
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet gets field from hash
func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from hash
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from hash
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// LPush pushes values to the left of list
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

// RPush pushes values to the right of list
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

// LPop pops value from the left of list
func (r *Redis) LPop(ctx context.Context, key string) (string, error) {
	return r.Client.LPop(ctx, key).Result()
}

// RPop pops value from the right of list
func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

// LLen returns length of list
func (r *Redis) LLen(ctx context.Context, key string) (int64, error) {
	return r.Client.LLen(ctx, key).Result()
}

// SAdd adds members to set
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of set
func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// SRem removes members from set
func (r *Redis) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

// ZAdd adds members to sorted set
func (r *Redis) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange returns members from sorted set by range
func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRem removes members from sorted set
func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.ZRem(ctx, key, members...).Err()
}

// Publish publishes message to channel
func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to patterns
func (r *Redis) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.PSubscribe(ctx, channels...)
}

// Incr increments key by 1
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Decr decrements key by 1
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}

// IncrBy increments key by value
func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

// DecrBy decrements key by value
func (r *Redis) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.DecrBy(ctx, key, value).Result()
}

// MSet sets multiple key-value pairs
func (r *Redis) MSet(ctx context.Context, values ...interface{}) error {
	return r.Client.MSet(ctx, values...).Err()
}

// MGet gets multiple values by keys
func (r *Redis) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return r.Client.MGet(ctx, keys...).Result()
}

// Keys returns all keys matching pattern
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()
}

// Scan iterates over keys matching pattern
func (r *Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.Client.Scan(ctx, cursor, match, count).Result()
}

// FlushDB flushes current database
func (r *Redis) FlushDB(ctx context.Context) error {
	return r.Client.FlushDB(ctx).Err()
}

// FlushAll flushes all databases
func (r *Redis) FlushAll(ctx context.Context) error {
	return r.Client.FlushAll(ctx).Err()
}

// HealthCheck performs a health check on Redis
func (r *Redis) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.Ping(ctx); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// Test basic operations
	testKey := "health_check_test"
	testValue := "test_value"

	if err := r.Set(ctx, testKey, testValue, time.Minute); err != nil {
		return fmt.Errorf("redis set operation failed: %w", err)
	}

	if _, err := r.Get(ctx, testKey); err != nil {
		return fmt.Errorf("redis get operation failed: %w", err)
	}

	if err := r.Del(ctx, testKey); err != nil {
		return fmt.Errorf("redis del operation failed: %w", err)
	}

	return nil
}

// Stats returns Redis connection statistics
func (r *Redis) Stats() *redis.PoolStats {
	return r.Client.PoolStats()
}
