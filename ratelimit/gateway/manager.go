package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit/types"
	"github.com/sirupsen/logrus"
)

// RateLimitManager manages multiple rate limiting providers
type RateLimitManager struct {
	providers map[string]RateLimitProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds rate limit manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// RateLimitProvider interface for rate limiting backends
type RateLimitProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.RateLimitFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Rate limiting operations
	Allow(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error)
	Reset(ctx context.Context, key string) error
	GetRemaining(ctx context.Context, key string, limit *types.RateLimit) (int64, error)
	GetResetTime(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error)

	// Batch operations
	AllowMultiple(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error)
	ResetMultiple(ctx context.Context, keys []string) error

	// Statistics and monitoring
	GetStats(ctx context.Context) (*types.RateLimitStats, error)
	GetKeys(ctx context.Context, pattern string) ([]string, error)
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(config *ManagerConfig, logger *logrus.Logger) *RateLimitManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &RateLimitManager{
		providers: make(map[string]RateLimitProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a rate limit provider
func (rlm *RateLimitManager) RegisterProvider(provider RateLimitProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := rlm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	rlm.providers[name] = provider
	rlm.logger.WithField("provider", name).Info("Rate limit provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (rlm *RateLimitManager) GetProvider(name string) (RateLimitProvider, error) {
	provider, exists := rlm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (rlm *RateLimitManager) GetDefaultProvider() (RateLimitProvider, error) {
	if rlm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range rlm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return rlm.GetProvider(rlm.config.DefaultProvider)
}

// Allow checks if a request is allowed using the default provider
func (rlm *RateLimitManager) Allow(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
	return rlm.AllowWithProvider(ctx, "", key, limit)
}

// AllowWithProvider checks if a request is allowed using a specific provider
func (rlm *RateLimitManager) AllowWithProvider(ctx context.Context, providerName string, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.RateLimitResult
	err = rlm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.Allow(ctx, key, limit)
		return err
	})

	return result, err
}

// Reset resets the rate limit for a key using the default provider
func (rlm *RateLimitManager) Reset(ctx context.Context, key string) error {
	return rlm.ResetWithProvider(ctx, "", key)
}

// ResetWithProvider resets the rate limit for a key using a specific provider
func (rlm *RateLimitManager) ResetWithProvider(ctx context.Context, providerName string, key string) error {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return err
	}

	return rlm.executeWithRetry(ctx, func() error {
		return provider.Reset(ctx, key)
	})
}

// GetRemaining returns the remaining requests for a key using the default provider
func (rlm *RateLimitManager) GetRemaining(ctx context.Context, key string, limit *types.RateLimit) (int64, error) {
	return rlm.GetRemainingWithProvider(ctx, "", key, limit)
}

// GetRemainingWithProvider returns the remaining requests for a key using a specific provider
func (rlm *RateLimitManager) GetRemainingWithProvider(ctx context.Context, providerName string, key string, limit *types.RateLimit) (int64, error) {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return 0, err
	}

	var result int64
	err = rlm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetRemaining(ctx, key, limit)
		return err
	})

	return result, err
}

// GetResetTime returns the reset time for a key using the default provider
func (rlm *RateLimitManager) GetResetTime(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error) {
	return rlm.GetResetTimeWithProvider(ctx, "", key, limit)
}

// GetResetTimeWithProvider returns the reset time for a key using a specific provider
func (rlm *RateLimitManager) GetResetTimeWithProvider(ctx context.Context, providerName string, key string, limit *types.RateLimit) (time.Time, error) {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return time.Time{}, err
	}

	var result time.Time
	err = rlm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetResetTime(ctx, key, limit)
		return err
	})

	return result, err
}

// AllowMultiple checks multiple rate limit requests using the default provider
func (rlm *RateLimitManager) AllowMultiple(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error) {
	return rlm.AllowMultipleWithProvider(ctx, "", requests)
}

// AllowMultipleWithProvider checks multiple rate limit requests using a specific provider
func (rlm *RateLimitManager) AllowMultipleWithProvider(ctx context.Context, providerName string, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error) {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []*types.RateLimitResult
	err = rlm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.AllowMultiple(ctx, requests)
		return err
	})

	return result, err
}

// ResetMultiple resets multiple rate limits using the default provider
func (rlm *RateLimitManager) ResetMultiple(ctx context.Context, keys []string) error {
	return rlm.ResetMultipleWithProvider(ctx, "", keys)
}

// ResetMultipleWithProvider resets multiple rate limits using a specific provider
func (rlm *RateLimitManager) ResetMultipleWithProvider(ctx context.Context, providerName string, keys []string) error {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return err
	}

	return rlm.executeWithRetry(ctx, func() error {
		return provider.ResetMultiple(ctx, keys)
	})
}

// GetStats returns rate limit statistics from all providers
func (rlm *RateLimitManager) GetStats(ctx context.Context) (map[string]*types.RateLimitStats, error) {
	stats := make(map[string]*types.RateLimitStats)

	for name, provider := range rlm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			rlm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// GetKeys returns keys matching a pattern using the default provider
func (rlm *RateLimitManager) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	return rlm.GetKeysWithProvider(ctx, "", pattern)
}

// GetKeysWithProvider returns keys matching a pattern using a specific provider
func (rlm *RateLimitManager) GetKeysWithProvider(ctx context.Context, providerName string, pattern string) ([]string, error) {
	provider, err := rlm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []string
	err = rlm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetKeys(ctx, pattern)
		return err
	})

	return result, err
}

// Close closes all providers
func (rlm *RateLimitManager) Close() error {
	var lastErr error

	for name, provider := range rlm.providers {
		if err := provider.Disconnect(context.Background()); err != nil {
			rlm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (rlm *RateLimitManager) getProvider(name string) (RateLimitProvider, error) {
	if name == "" {
		return rlm.GetDefaultProvider()
	}
	return rlm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (rlm *RateLimitManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rlm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rlm.config.RetryDelay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			rlm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Rate limit operation failed, retrying")
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", rlm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (rlm *RateLimitManager) ListProviders() []string {
	providers := make([]string, 0, len(rlm.providers))
	for name := range rlm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (rlm *RateLimitManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range rlm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
