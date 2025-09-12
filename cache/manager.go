package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/cache/types"
	"github.com/sirupsen/logrus"
)

// CacheManager manages multiple cache providers
type CacheManager struct {
	providers map[string]CacheProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds cache manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// CacheProvider interface for cache backends
type CacheProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.CacheFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Advanced operations
	SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error
	InvalidateByTag(ctx context.Context, tag string) error
	GetStats(ctx context.Context) (*types.CacheStats, error)
	Flush(ctx context.Context) error

	// Batch operations
	SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteMultiple(ctx context.Context, keys []string) error

	// Key management
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *ManagerConfig, logger *logrus.Logger) *CacheManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &CacheManager{
		providers: make(map[string]CacheProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a cache provider
func (cm *CacheManager) RegisterProvider(provider CacheProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := cm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	cm.providers[name] = provider
	cm.logger.WithField("provider", name).Info("Cache provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (cm *CacheManager) GetProvider(name string) (CacheProvider, error) {
	provider, exists := cm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (cm *CacheManager) GetDefaultProvider() (CacheProvider, error) {
	if cm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range cm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return cm.GetProvider(cm.config.DefaultProvider)
}

// Set stores a value in the cache using the default provider
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return cm.SetWithProvider(ctx, "", key, value, ttl)
}

// SetWithProvider stores a value in the cache using a specific provider
func (cm *CacheManager) SetWithProvider(ctx context.Context, providerName string, key string, value interface{}, ttl time.Duration) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.Set(ctx, key, value, ttl)
	})
}

// Get retrieves a value from the cache using the default provider
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	return cm.GetWithProvider(ctx, "", key, dest)
}

// GetWithProvider retrieves a value from the cache using a specific provider
func (cm *CacheManager) GetWithProvider(ctx context.Context, providerName string, key string, dest interface{}) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.Get(ctx, key, dest)
	})
}

// Delete removes a value from the cache using the default provider
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.DeleteWithProvider(ctx, "", key)
}

// DeleteWithProvider removes a value from the cache using a specific provider
func (cm *CacheManager) DeleteWithProvider(ctx context.Context, providerName string, key string) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.Delete(ctx, key)
	})
}

// Exists checks if a key exists in the cache using the default provider
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	return cm.ExistsWithProvider(ctx, "", key)
}

// ExistsWithProvider checks if a key exists in the cache using a specific provider
func (cm *CacheManager) ExistsWithProvider(ctx context.Context, providerName string, key string) (bool, error) {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return false, err
	}

	var result bool
	err = cm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.Exists(ctx, key)
		return err
	})

	return result, err
}

// SetWithTags stores a value with tags for easier invalidation
func (cm *CacheManager) SetWithTags(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	return cm.SetWithTagsAndProvider(ctx, "", key, value, ttl, tags)
}

// SetWithTagsAndProvider stores a value with tags using a specific provider
func (cm *CacheManager) SetWithTagsAndProvider(ctx context.Context, providerName string, key string, value interface{}, ttl time.Duration, tags []string) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.SetWithTags(ctx, key, value, ttl, tags)
	})
}

// InvalidateByTag invalidates all keys with a specific tag
func (cm *CacheManager) InvalidateByTag(ctx context.Context, tag string) error {
	return cm.InvalidateByTagWithProvider(ctx, "", tag)
}

// InvalidateByTagWithProvider invalidates all keys with a specific tag using a specific provider
func (cm *CacheManager) InvalidateByTagWithProvider(ctx context.Context, providerName string, tag string) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.InvalidateByTag(ctx, tag)
	})
}

// GetStats returns cache statistics from all providers
func (cm *CacheManager) GetStats(ctx context.Context) (map[string]*types.CacheStats, error) {
	stats := make(map[string]*types.CacheStats)

	for name, provider := range cm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			cm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// Flush clears all cache data from all providers
func (cm *CacheManager) Flush(ctx context.Context) error {
	var lastErr error

	for name, provider := range cm.providers {
		if err := provider.Flush(ctx); err != nil {
			cm.logger.WithError(err).WithField("provider", name).Error("Failed to flush provider")
			lastErr = err
		}
	}

	return lastErr
}

// SetMultiple stores multiple values in the cache
func (cm *CacheManager) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return cm.SetMultipleWithProvider(ctx, "", items, ttl)
}

// SetMultipleWithProvider stores multiple values using a specific provider
func (cm *CacheManager) SetMultipleWithProvider(ctx context.Context, providerName string, items map[string]interface{}, ttl time.Duration) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.SetMultiple(ctx, items, ttl)
	})
}

// GetMultiple retrieves multiple values from the cache
func (cm *CacheManager) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return cm.GetMultipleWithProvider(ctx, "", keys)
}

// GetMultipleWithProvider retrieves multiple values using a specific provider
func (cm *CacheManager) GetMultipleWithProvider(ctx context.Context, providerName string, keys []string) (map[string]interface{}, error) {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = cm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetMultiple(ctx, keys)
		return err
	})

	return result, err
}

// DeleteMultiple removes multiple values from the cache
func (cm *CacheManager) DeleteMultiple(ctx context.Context, keys []string) error {
	return cm.DeleteMultipleWithProvider(ctx, "", keys)
}

// DeleteMultipleWithProvider removes multiple values using a specific provider
func (cm *CacheManager) DeleteMultipleWithProvider(ctx context.Context, providerName string, keys []string) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.DeleteMultiple(ctx, keys)
	})
}

// GetKeys returns keys matching a pattern
func (cm *CacheManager) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	return cm.GetKeysWithProvider(ctx, "", pattern)
}

// GetKeysWithProvider returns keys matching a pattern using a specific provider
func (cm *CacheManager) GetKeysWithProvider(ctx context.Context, providerName string, pattern string) ([]string, error) {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []string
	err = cm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetKeys(ctx, pattern)
		return err
	})

	return result, err
}

// GetTTL returns the TTL of a key
func (cm *CacheManager) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return cm.GetTTLWithProvider(ctx, "", key)
}

// GetTTLWithProvider returns the TTL of a key using a specific provider
func (cm *CacheManager) GetTTLWithProvider(ctx context.Context, providerName string, key string) (time.Duration, error) {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return 0, err
	}

	var result time.Duration
	err = cm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetTTL(ctx, key)
		return err
	})

	return result, err
}

// SetTTL sets the TTL of a key
func (cm *CacheManager) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return cm.SetTTLWithProvider(ctx, "", key, ttl)
}

// SetTTLWithProvider sets the TTL of a key using a specific provider
func (cm *CacheManager) SetTTLWithProvider(ctx context.Context, providerName string, key string, ttl time.Duration) error {
	provider, err := cm.getProvider(providerName)
	if err != nil {
		return err
	}

	return cm.executeWithRetry(ctx, func() error {
		return provider.SetTTL(ctx, key, ttl)
	})
}

// Close closes all providers
func (cm *CacheManager) Close() error {
	var lastErr error

	for name, provider := range cm.providers {
		if err := provider.Disconnect(context.Background()); err != nil {
			cm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (cm *CacheManager) getProvider(name string) (CacheProvider, error) {
	if name == "" {
		return cm.GetDefaultProvider()
	}
	return cm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (cm *CacheManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cm.config.RetryDelay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			cm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Cache operation failed, retrying")
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", cm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (cm *CacheManager) ListProviders() []string {
	providers := make([]string, 0, len(cm.providers))
	for name := range cm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (cm *CacheManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range cm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
