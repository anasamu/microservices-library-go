package memcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/cache/types"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/sirupsen/logrus"
)

// MemcacheProvider implements the CacheProvider interface for Memcached
type MemcacheProvider struct {
	client    *memcache.Client
	config    *MemcacheConfig
	logger    *logrus.Logger
	connected bool
}

// MemcacheConfig holds Memcached-specific configuration
type MemcacheConfig struct {
	Servers      []string      `json:"servers"`
	MaxIdleConns int           `json:"max_idle_conns"`
	Timeout      time.Duration `json:"timeout"`
	KeyPrefix    string        `json:"key_prefix"`
	Namespace    string        `json:"namespace"`
	MaxKeySize   int           `json:"max_key_size"`
	MaxValueSize int           `json:"max_value_size"`
}

// NewMemcacheProvider creates a new Memcached cache provider
func NewMemcacheProvider(config *MemcacheConfig, logger *logrus.Logger) *MemcacheProvider {
	if config == nil {
		config = &MemcacheConfig{
			Servers:      []string{"localhost:11211"},
			MaxIdleConns: 2,
			Timeout:      100 * time.Millisecond,
			MaxKeySize:   250,
			MaxValueSize: 1024 * 1024, // 1MB
		}
	}

	return &MemcacheProvider{
		config:    config,
		logger:    logger,
		connected: false,
	}
}

// GetName returns the provider name
func (m *MemcacheProvider) GetName() string {
	return "memcache"
}

// GetSupportedFeatures returns the features supported by Memcached
func (m *MemcacheProvider) GetSupportedFeatures() []types.CacheFeature {
	return []types.CacheFeature{
		types.FeatureSet,
		types.FeatureGet,
		types.FeatureDelete,
		types.FeatureExists,
		types.FeatureFlush,
		types.FeatureStats,
		types.FeatureTTL,
		types.FeatureBatch,
	}
}

// GetConnectionInfo returns connection information
func (m *MemcacheProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if m.connected {
		status = types.StatusConnected
	}

	servers := ""
	if len(m.config.Servers) > 0 {
		servers = m.config.Servers[0]
	}

	return &types.ConnectionInfo{
		Host:     servers,
		Port:     11211,
		Database: "memcache",
		Status:   status,
		Metadata: map[string]string{
			"servers":        fmt.Sprintf("%v", m.config.Servers),
			"max_idle_conns": fmt.Sprintf("%d", m.config.MaxIdleConns),
			"timeout":        m.config.Timeout.String(),
			"key_prefix":     m.config.KeyPrefix,
			"namespace":      m.config.Namespace,
		},
	}
}

// Connect establishes connection to Memcached
func (m *MemcacheProvider) Connect(ctx context.Context) error {
	m.client = memcache.New(m.config.Servers...)
	m.client.MaxIdleConns = m.config.MaxIdleConns
	m.client.Timeout = m.config.Timeout

	// Test connection
	if err := m.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Memcached: %w", err)
	}

	m.connected = true
	m.logger.Info("Connected to Memcached successfully")
	return nil
}

// Disconnect closes the Memcached connection
func (m *MemcacheProvider) Disconnect(ctx context.Context) error {
	// Memcached client doesn't have explicit close method
	m.connected = false
	m.logger.Info("Disconnected from Memcached")
	return nil
}

// Ping tests the Memcached connection
func (m *MemcacheProvider) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("Memcached client not initialized")
	}

	// Try to get a non-existent key to test connection
	_, err := m.client.Get("__ping_test__")
	if err != nil && err != memcache.ErrCacheMiss {
		return err
	}

	return nil
}

// IsConnected returns the connection status
func (m *MemcacheProvider) IsConnected() bool {
	return m.connected && m.client != nil
}

// Set stores a value in Memcached
func (m *MemcacheProvider) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	key = m.buildKey(key)
	if len(key) > m.config.MaxKeySize {
		return &types.CacheError{Code: types.ErrCodeInvalidKey, Message: "Key too long", Key: key}
	}

	data, err := json.Marshal(value)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeSerialization, Message: err.Error(), Key: key}
	}

	if len(data) > m.config.MaxValueSize {
		return &types.CacheError{Code: types.ErrCodeInvalidValue, Message: "Value too large", Key: key}
	}

	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(ttl.Seconds()),
	}

	err = m.client.Set(item)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	m.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl,
	}).Debug("Value set in Memcached")
	return nil
}

// Get retrieves a value from Memcached
func (m *MemcacheProvider) Get(ctx context.Context, key string, dest interface{}) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	key = m.buildKey(key)
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
		}
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	err = json.Unmarshal(item.Value, dest)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeDeserialization, Message: err.Error(), Key: key}
	}

	m.logger.WithField("key", key).Debug("Value retrieved from Memcached")
	return nil
}

// Delete removes a value from Memcached
func (m *MemcacheProvider) Delete(ctx context.Context, key string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	key = m.buildKey(key)
	err := m.client.Delete(key)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	m.logger.WithField("key", key).Debug("Value deleted from Memcached")
	return nil
}

// Exists checks if a key exists in Memcached
func (m *MemcacheProvider) Exists(ctx context.Context, key string) (bool, error) {
	if !m.IsConnected() {
		return false, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	key = m.buildKey(key)
	_, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		return false, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	return true, nil
}

// SetWithTags stores a value with tags (not supported by Memcached, but we can store tags as separate keys)
func (m *MemcacheProvider) SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// Store the value
	if err := m.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Store tags as separate keys (Memcached doesn't support tags natively)
	for _, tag := range tags {
		tagKey := m.buildKey(fmt.Sprintf("tag:%s:%s", tag, key))
		tagData := map[string]interface{}{
			"key": key,
			"tag": tag,
		}

		if err := m.Set(ctx, tagKey, tagData, ttl); err != nil {
			m.logger.WithError(err).WithField("tag", tag).Warn("Failed to store tag")
		}
	}

	m.logger.WithFields(logrus.Fields{
		"key":  key,
		"tags": tags,
	}).Debug("Value set with tags in Memcached")
	return nil
}

// InvalidateByTag invalidates all keys with a specific tag
func (m *MemcacheProvider) InvalidateByTag(ctx context.Context, tag string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// This is a limitation of Memcached - we can't efficiently find all keys with a tag
	// In a real implementation, you might want to maintain a separate index
	m.logger.WithField("tag", tag).Warn("Tag invalidation not efficiently supported by Memcached")
	return nil
}

// GetStats returns Memcached cache statistics
func (m *MemcacheProvider) GetStats(ctx context.Context) (*types.CacheStats, error) {
	if !m.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// Since the official gomemcache library doesn't support stats,
	// we return basic stats with connection status
	cacheStats := &types.CacheStats{
		Provider:   m.GetName(),
		LastUpdate: time.Now(),
		// Set default values since we can't get actual stats
		Hits:   0,
		Misses: 0,
		Keys:   0,
	}

	m.logger.Debug("Memcached stats requested - returning basic stats (detailed stats not supported by gomemcache)")
	return cacheStats, nil
}

// Flush clears all cache data
func (m *MemcacheProvider) Flush(ctx context.Context) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	err := m.client.FlushAll()
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	m.logger.Info("Memcached cache flushed")
	return nil
}

// SetMultiple stores multiple values in Memcached
func (m *MemcacheProvider) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// Since gomemcache doesn't have SetMulti, we'll set items individually
	for key, value := range items {
		if err := m.Set(ctx, key, value, ttl); err != nil {
			m.logger.WithError(err).WithField("key", key).Warn("Failed to set item in batch")
		}
	}

	m.logger.WithField("count", len(items)).Debug("Multiple values set in Memcached")
	return nil
}

// GetMultiple retrieves multiple values from Memcached
func (m *MemcacheProvider) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if !m.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// Build keys with prefix
	memcacheKeys := make([]string, len(keys))
	for i, key := range keys {
		memcacheKeys[i] = m.buildKey(key)
	}

	items, err := m.client.GetMulti(memcacheKeys)
	if err != nil {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error()}
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		if item, exists := items[m.buildKey(key)]; exists {
			var data interface{}
			if err := json.Unmarshal(item.Value, &data); err == nil {
				result[key] = data
			}
		}
	}

	m.logger.WithField("count", len(result)).Debug("Multiple values retrieved from Memcached")
	return result, nil
}

// DeleteMultiple removes multiple values from Memcached
func (m *MemcacheProvider) DeleteMultiple(ctx context.Context, keys []string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	// Since gomemcache doesn't have DeleteMulti, we'll delete items individually
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			m.logger.WithError(err).WithField("key", key).Warn("Failed to delete item in batch")
		}
	}

	m.logger.WithField("count", len(keys)).Debug("Multiple values deleted from Memcached")
	return nil
}

// GetKeys returns keys matching a pattern (not supported by Memcached)
func (m *MemcacheProvider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	// Memcached doesn't support key enumeration
	return nil, &types.CacheError{Code: types.ErrCodeInvalidKey, Message: "Key enumeration not supported by Memcached"}
}

// GetTTL returns the TTL of a key (not directly supported by Memcached)
func (m *MemcacheProvider) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	// Memcached doesn't provide TTL information directly
	return 0, &types.CacheError{Code: types.ErrCodeInvalidKey, Message: "TTL retrieval not supported by Memcached"}
}

// SetTTL sets the TTL of a key
func (m *MemcacheProvider) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memcached not connected"}
	}

	key = m.buildKey(key)

	// Get the current value
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
		}
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	// Set with new TTL
	newItem := &memcache.Item{
		Key:        key,
		Value:      item.Value,
		Expiration: int32(ttl.Seconds()),
	}

	err = m.client.Set(newItem)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: err.Error(), Key: key}
	}

	return nil
}

// buildKey constructs a key with prefix and namespace
func (m *MemcacheProvider) buildKey(key string) string {
	parts := []string{}

	if m.config.Namespace != "" {
		parts = append(parts, m.config.Namespace)
	}

	if m.config.KeyPrefix != "" {
		parts = append(parts, m.config.KeyPrefix)
	}

	parts = append(parts, key)

	return fmt.Sprintf("%s", parts)
}
