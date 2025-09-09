package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
)

// CircuitBreakerManager manages multiple circuit breaker providers
type CircuitBreakerManager struct {
	providers map[string]CircuitBreakerProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds circuit breaker manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// CircuitBreakerProvider interface for circuit breaker backends
type CircuitBreakerProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.CircuitBreakerFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Circuit breaker operations
	Execute(ctx context.Context, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error)
	ExecuteWithFallback(ctx context.Context, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error)
	GetState(ctx context.Context, name string) (types.CircuitState, error)
	GetStats(ctx context.Context, name string) (*types.CircuitBreakerStats, error)
	Reset(ctx context.Context, name string) error

	// Configuration management
	Configure(ctx context.Context, name string, config *types.CircuitBreakerConfig) error
	GetConfig(ctx context.Context, name string) (*types.CircuitBreakerConfig, error)

	// Advanced operations
	GetAllStates(ctx context.Context) (map[string]types.CircuitState, error)
	GetAllStats(ctx context.Context) (map[string]*types.CircuitBreakerStats, error)
	ResetAll(ctx context.Context) error
	Close(ctx context.Context) error
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config *ManagerConfig, logger *logrus.Logger) *CircuitBreakerManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &CircuitBreakerManager{
		providers: make(map[string]CircuitBreakerProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a circuit breaker provider
func (cbm *CircuitBreakerManager) RegisterProvider(provider CircuitBreakerProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := cbm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	cbm.providers[name] = provider
	cbm.logger.WithField("provider", name).Info("Circuit breaker provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (cbm *CircuitBreakerManager) GetProvider(name string) (CircuitBreakerProvider, error) {
	provider, exists := cbm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (cbm *CircuitBreakerManager) GetDefaultProvider() (CircuitBreakerProvider, error) {
	if cbm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range cbm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return cbm.GetProvider(cbm.config.DefaultProvider)
}

// Execute runs a function through the circuit breaker using the default provider
func (cbm *CircuitBreakerManager) Execute(ctx context.Context, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error) {
	return cbm.ExecuteWithProvider(ctx, "", name, fn)
}

// ExecuteWithProvider runs a function through the circuit breaker using a specific provider
func (cbm *CircuitBreakerManager) ExecuteWithProvider(ctx context.Context, providerName string, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error) {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		return provider.Execute(ctx, name, fn)
	})
	return result, err
}

// ExecuteWithFallback runs a function with fallback through the circuit breaker using the default provider
func (cbm *CircuitBreakerManager) ExecuteWithFallback(ctx context.Context, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error) {
	return cbm.ExecuteWithFallbackAndProvider(ctx, "", name, fn, fallback)
}

// ExecuteWithFallbackAndProvider runs a function with fallback through the circuit breaker using a specific provider
func (cbm *CircuitBreakerManager) ExecuteWithFallbackAndProvider(ctx context.Context, providerName string, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error) {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		return provider.ExecuteWithFallback(ctx, name, fn, fallback)
	})
	return result, err
}

// GetState returns the state of a circuit breaker using the default provider
func (cbm *CircuitBreakerManager) GetState(ctx context.Context, name string) (types.CircuitState, error) {
	return cbm.GetStateWithProvider(ctx, "", name)
}

// GetStateWithProvider returns the state of a circuit breaker using a specific provider
func (cbm *CircuitBreakerManager) GetStateWithProvider(ctx context.Context, providerName string, name string) (types.CircuitState, error) {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return "", err
	}

	var result types.CircuitState
	_, err = cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		var err error
		result, err = provider.GetState(ctx, name)
		return nil, err
	})

	return result, err
}

// GetStats returns circuit breaker statistics using the default provider
func (cbm *CircuitBreakerManager) GetStats(ctx context.Context, name string) (*types.CircuitBreakerStats, error) {
	return cbm.GetStatsWithProvider(ctx, "", name)
}

// GetStatsWithProvider returns circuit breaker statistics using a specific provider
func (cbm *CircuitBreakerManager) GetStatsWithProvider(ctx context.Context, providerName string, name string) (*types.CircuitBreakerStats, error) {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.CircuitBreakerStats
	_, err = cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		var err error
		result, err = provider.GetStats(ctx, name)
		return nil, err
	})

	return result, err
}

// Reset resets a circuit breaker using the default provider
func (cbm *CircuitBreakerManager) Reset(ctx context.Context, name string) error {
	return cbm.ResetWithProvider(ctx, "", name)
}

// ResetWithProvider resets a circuit breaker using a specific provider
func (cbm *CircuitBreakerManager) ResetWithProvider(ctx context.Context, providerName string, name string) error {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return err
	}

	_, err = cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		return nil, provider.Reset(ctx, name)
	})
	return err
}

// Configure configures a circuit breaker using the default provider
func (cbm *CircuitBreakerManager) Configure(ctx context.Context, name string, config *types.CircuitBreakerConfig) error {
	return cbm.ConfigureWithProvider(ctx, "", name, config)
}

// ConfigureWithProvider configures a circuit breaker using a specific provider
func (cbm *CircuitBreakerManager) ConfigureWithProvider(ctx context.Context, providerName string, name string, config *types.CircuitBreakerConfig) error {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return err
	}

	_, err = cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		return nil, provider.Configure(ctx, name, config)
	})
	return err
}

// GetConfig returns circuit breaker configuration using the default provider
func (cbm *CircuitBreakerManager) GetConfig(ctx context.Context, name string) (*types.CircuitBreakerConfig, error) {
	return cbm.GetConfigWithProvider(ctx, "", name)
}

// GetConfigWithProvider returns circuit breaker configuration using a specific provider
func (cbm *CircuitBreakerManager) GetConfigWithProvider(ctx context.Context, providerName string, name string) (*types.CircuitBreakerConfig, error) {
	provider, err := cbm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.CircuitBreakerConfig
	_, err = cbm.executeWithRetry(ctx, func() (*types.ExecutionResult, error) {
		var err error
		result, err = provider.GetConfig(ctx, name)
		return nil, err
	})

	return result, err
}

// GetAllStates returns states of all circuit breakers from all providers
func (cbm *CircuitBreakerManager) GetAllStates(ctx context.Context) (map[string]map[string]types.CircuitState, error) {
	allStates := make(map[string]map[string]types.CircuitState)

	for name, provider := range cbm.providers {
		states, err := provider.GetAllStates(ctx)
		if err != nil {
			cbm.logger.WithError(err).WithField("provider", name).Warn("Failed to get states from provider")
			continue
		}
		allStates[name] = states
	}

	return allStates, nil
}

// GetAllStats returns statistics of all circuit breakers from all providers
func (cbm *CircuitBreakerManager) GetAllStats(ctx context.Context) (map[string]map[string]*types.CircuitBreakerStats, error) {
	allStats := make(map[string]map[string]*types.CircuitBreakerStats)

	for name, provider := range cbm.providers {
		stats, err := provider.GetAllStats(ctx)
		if err != nil {
			cbm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		allStats[name] = stats
	}

	return allStats, nil
}

// ResetAll resets all circuit breakers from all providers
func (cbm *CircuitBreakerManager) ResetAll(ctx context.Context) error {
	var lastErr error

	for name, provider := range cbm.providers {
		if err := provider.ResetAll(ctx); err != nil {
			cbm.logger.WithError(err).WithField("provider", name).Error("Failed to reset all circuit breakers in provider")
			lastErr = err
		}
	}

	return lastErr
}

// Close closes all providers
func (cbm *CircuitBreakerManager) Close() error {
	var lastErr error

	for name, provider := range cbm.providers {
		if err := provider.Close(context.Background()); err != nil {
			cbm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (cbm *CircuitBreakerManager) getProvider(name string) (CircuitBreakerProvider, error) {
	if name == "" {
		return cbm.GetDefaultProvider()
	}
	return cbm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (cbm *CircuitBreakerManager) executeWithRetry(ctx context.Context, fn func() (*types.ExecutionResult, error)) (*types.ExecutionResult, error) {
	var lastErr error
	var lastResult *types.ExecutionResult

	for attempt := 0; attempt <= cbm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(cbm.config.RetryDelay):
			}
		}

		result, err := fn()
		if err != nil {
			lastErr = err
			lastResult = result
			cbm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Circuit breaker operation failed, retrying")
			continue
		}

		return result, nil
	}

	return lastResult, fmt.Errorf("operation failed after %d attempts: %w", cbm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (cbm *CircuitBreakerManager) ListProviders() []string {
	providers := make([]string, 0, len(cbm.providers))
	for name := range cbm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (cbm *CircuitBreakerManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range cbm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
