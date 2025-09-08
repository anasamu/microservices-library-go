package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/cache/types"
	"github.com/sirupsen/logrus"
)

// MemoryProvider implements the CacheProvider interface for in-memory caching
type MemoryProvider struct {
	store     map[string]*types.CacheItem
	mu        sync.RWMutex
	config    *MemoryConfig
	logger    *logrus.Logger
	connected bool
	stats     *types.CacheStats
}

// MemoryConfig holds memory-specific configuration
type MemoryConfig struct {
	MaxSize         int           `json:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	KeyPrefix       string        `json:"key_prefix"`
	Namespace       string        `json:"namespace"`
}

// NewMemoryProvider creates a new memory cache provider
func NewMemoryProvider(config *MemoryConfig, logger *logrus.Logger) *MemoryProvider {
	if config == nil {
		config = &MemoryConfig{
			MaxSize:         10000,
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		}
	}

	provider := &MemoryProvider{
		store:     make(map[string]*types.CacheItem),
		config:    config,
		logger:    logger,
		connected: false,
		stats: &types.CacheStats{
			Provider:   "memory",
			LastUpdate: time.Now(),
		},
	}

	// Start cleanup goroutine
	go provider.startCleanup()

	return provider
}

// GetName returns the provider name
func (m *MemoryProvider) GetName() string {
	return "memory"
}

// GetSupportedFeatures returns the features supported by memory cache
func (m *MemoryProvider) GetSupportedFeatures() []types.CacheFeature {
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
	}
}

// GetConnectionInfo returns connection information
func (m *MemoryProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if m.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Database: "memory",
		Status:   status,
		Metadata: map[string]string{
			"max_size":         fmt.Sprintf("%d", m.config.MaxSize),
			"default_ttl":      m.config.DefaultTTL.String(),
			"cleanup_interval": m.config.CleanupInterval.String(),
			"key_prefix":       m.config.KeyPrefix,
			"namespace":        m.config.Namespace,
		},
	}
}

// Connect establishes connection (always succeeds for memory)
func (m *MemoryProvider) Connect(ctx context.Context) error {
	m.connected = true
	m.logger.Info("Connected to memory cache successfully")
	return nil
}

// Disconnect closes the connection
func (m *MemoryProvider) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.store = make(map[string]*types.CacheItem)
	m.logger.Info("Disconnected from memory cache")
	return nil
}

// Ping tests the connection (always succeeds for memory)
func (m *MemoryProvider) Ping(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("memory cache not connected")
	}
	return nil
}

// IsConnected returns the connection status
func (m *MemoryProvider) IsConnected() bool {
	return m.connected
}

// Set stores a value in memory
func (m *MemoryProvider) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	// Check if we need to evict items
	if len(m.store) >= m.config.MaxSize {
		m.evictOldest()
	}

	item := &types.CacheItem{
		Key:        key,
		Value:      value,
		Expiration: time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tags:       []string{},
		Metadata:   make(map[string]interface{}),
	}

	m.mu.Lock()
	m.store[key] = item
	m.mu.Unlock()

	m.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl,
	}).Debug("Value set in memory cache")
	return nil
}

// Get retrieves a value from memory
func (m *MemoryProvider) Get(ctx context.Context, key string, dest interface{}) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	m.mu.RLock()
	item, exists := m.store[key]
	m.mu.RUnlock()

	if !exists {
		m.stats.Misses++
		return &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
	}

	// Check if expired
	if time.Now().After(item.Expiration) {
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		m.stats.Misses++
		return &types.CacheError{Code: types.ErrCodeExpired, Message: "Key expired", Key: key}
	}

	// Update access time
	item.UpdatedAt = time.Now()
	m.stats.Hits++

	// Copy value to destination
	data, err := json.Marshal(item.Value)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeSerialization, Message: err.Error(), Key: key}
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return &types.CacheError{Code: types.ErrCodeDeserialization, Message: err.Error(), Key: key}
	}

	m.logger.WithField("key", key).Debug("Value retrieved from memory cache")
	return nil
}

// Delete removes a value from memory
func (m *MemoryProvider) Delete(ctx context.Context, key string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	m.mu.Lock()
	delete(m.store, key)
	m.mu.Unlock()

	m.logger.WithField("key", key).Debug("Value deleted from memory cache")
	return nil
}

// Exists checks if a key exists in memory
func (m *MemoryProvider) Exists(ctx context.Context, key string) (bool, error) {
	if !m.IsConnected() {
		return false, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	m.mu.RLock()
	item, exists := m.store[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(item.Expiration) {
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// SetWithTags stores a value with tags for easier invalidation
func (m *MemoryProvider) SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	// Check if we need to evict items
	if len(m.store) >= m.config.MaxSize {
		m.evictOldest()
	}

	item := &types.CacheItem{
		Key:        key,
		Value:      value,
		Expiration: time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tags:       tags,
		Metadata:   make(map[string]interface{}),
	}

	m.mu.Lock()
	m.store[key] = item
	m.mu.Unlock()

	m.logger.WithFields(logrus.Fields{
		"key":  key,
		"tags": tags,
	}).Debug("Value set with tags in memory cache")
	return nil
}

// InvalidateByTag invalidates all keys with a specific tag
func (m *MemoryProvider) InvalidateByTag(ctx context.Context, tag string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for key, item := range m.store {
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				delete(m.store, key)
				count++
				break
			}
		}
	}

	m.logger.WithFields(logrus.Fields{
		"tag":  tag,
		"keys": count,
	}).Info("Invalidated cache by tag in memory cache")
	return nil
}

// GetStats returns memory cache statistics
func (m *MemoryProvider) GetStats(ctx context.Context) (*types.CacheStats, error) {
	if !m.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &types.CacheStats{
		Provider:   m.GetName(),
		Hits:       m.stats.Hits,
		Misses:     m.stats.Misses,
		Keys:       int64(len(m.store)),
		Memory:     m.calculateMemoryUsage(),
		Uptime:     time.Since(m.stats.LastUpdate),
		LastUpdate: time.Now(),
	}

	return stats, nil
}

// Flush clears all cache data
func (m *MemoryProvider) Flush(ctx context.Context) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	m.mu.Lock()
	m.store = make(map[string]*types.CacheItem)
	m.mu.Unlock()

	m.logger.Info("Memory cache flushed")
	return nil
}

// SetMultiple stores multiple values in memory
func (m *MemoryProvider) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, value := range items {
		key = m.buildKey(key)

		// Check if we need to evict items
		if len(m.store) >= m.config.MaxSize {
			m.evictOldest()
		}

		item := &types.CacheItem{
			Key:        key,
			Value:      value,
			Expiration: time.Now().Add(ttl),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Tags:       []string{},
			Metadata:   make(map[string]interface{}),
		}

		m.store[key] = item
	}

	m.logger.WithField("count", len(items)).Debug("Multiple values set in memory cache")
	return nil
}

// GetMultiple retrieves multiple values from memory
func (m *MemoryProvider) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if !m.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	result := make(map[string]interface{})

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range keys {
		key = m.buildKey(key)
		item, exists := m.store[key]

		if !exists {
			continue
		}

		// Check if expired
		if time.Now().After(item.Expiration) {
			continue
		}

		result[key] = item.Value
	}

	m.logger.WithField("count", len(result)).Debug("Multiple values retrieved from memory cache")
	return result, nil
}

// DeleteMultiple removes multiple values from memory
func (m *MemoryProvider) DeleteMultiple(ctx context.Context, keys []string) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		key = m.buildKey(key)
		delete(m.store, key)
	}

	m.logger.WithField("count", len(keys)).Debug("Multiple values deleted from memory cache")
	return nil
}

// GetKeys returns keys matching a pattern
func (m *MemoryProvider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	if !m.IsConnected() {
		return nil, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	var result []string

	m.mu.RLock()
	defer m.mu.RUnlock()

	for key := range m.store {
		// Simple pattern matching (can be enhanced with regex)
		if pattern == "*" || key == pattern {
			result = append(result, m.removeKeyPrefix(key))
		}
	}

	return result, nil
}

// GetTTL returns the TTL of a key
func (m *MemoryProvider) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if !m.IsConnected() {
		return 0, &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	m.mu.RLock()
	item, exists := m.store[key]
	m.mu.RUnlock()

	if !exists {
		return 0, &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
	}

	// Check if expired
	if time.Now().After(item.Expiration) {
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return 0, &types.CacheError{Code: types.ErrCodeExpired, Message: "Key expired", Key: key}
	}

	return time.Until(item.Expiration), nil
}

// SetTTL sets the TTL of a key
func (m *MemoryProvider) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	if !m.IsConnected() {
		return &types.CacheError{Code: types.ErrCodeConnection, Message: "Memory cache not connected"}
	}

	key = m.buildKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.store[key]
	if !exists {
		return &types.CacheError{Code: types.ErrCodeNotFound, Message: "Key not found", Key: key}
	}

	item.Expiration = time.Now().Add(ttl)
	item.UpdatedAt = time.Now()

	return nil
}

// buildKey constructs a key with prefix and namespace
func (m *MemoryProvider) buildKey(key string) string {
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

// removeKeyPrefix removes the prefix from a key
func (m *MemoryProvider) removeKeyPrefix(key string) string {
	prefix := m.buildKey("")
	if len(prefix) > 0 && len(key) > len(prefix) {
		return key[len(prefix):]
	}
	return key
}

// evictOldest removes the oldest item from the cache
func (m *MemoryProvider) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range m.store {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(m.store, oldestKey)
	}
}

// calculateMemoryUsage estimates memory usage
func (m *MemoryProvider) calculateMemoryUsage() int64 {
	// Rough estimation - in real implementation, you might want to use unsafe.Sizeof
	return int64(len(m.store) * 100) // Assume ~100 bytes per item
}

// startCleanup starts the cleanup goroutine
func (m *MemoryProvider) startCleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupExpired()
	}
}

// cleanupExpired removes expired items
func (m *MemoryProvider) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0

	for key, item := range m.store {
		if now.After(item.Expiration) {
			delete(m.store, key)
			count++
		}
	}

	if count > 0 {
		m.logger.WithField("count", count).Debug("Cleaned up expired items from memory cache")
	}
}
