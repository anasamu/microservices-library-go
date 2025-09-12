package failover

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/sirupsen/logrus"
)

// FailoverManager manages multiple failover providers
type FailoverManager struct {
	providers map[string]FailoverProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds failover manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// FailoverProvider interface for failover backends
type FailoverProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.FailoverFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Endpoint management
	RegisterEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error
	DeregisterEndpoint(ctx context.Context, endpointID string) error
	UpdateEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error
	GetEndpoint(ctx context.Context, endpointID string) (*types.ServiceEndpoint, error)
	ListEndpoints(ctx context.Context) ([]*types.ServiceEndpoint, error)

	// Failover operations
	SelectEndpoint(ctx context.Context, config *types.FailoverConfig) (*types.FailoverResult, error)
	ExecuteWithFailover(ctx context.Context, config *types.FailoverConfig, fn func(*types.ServiceEndpoint) (interface{}, error)) (*types.FailoverResult, error)
	HealthCheck(ctx context.Context, endpointID string) (types.HealthStatus, error)
	HealthCheckAll(ctx context.Context) (map[string]types.HealthStatus, error)

	// Configuration management
	Configure(ctx context.Context, config *types.FailoverConfig) error
	GetConfig(ctx context.Context) (*types.FailoverConfig, error)

	// Statistics and monitoring
	GetStats(ctx context.Context) (*types.FailoverStats, error)
	GetEvents(ctx context.Context, limit int) ([]*types.FailoverEvent, error)
	Close(ctx context.Context) error
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(config *ManagerConfig, logger *logrus.Logger) *FailoverManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &FailoverManager{
		providers: make(map[string]FailoverProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a failover provider
func (fm *FailoverManager) RegisterProvider(provider FailoverProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := fm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	fm.providers[name] = provider
	fm.logger.WithField("provider", name).Info("Failover provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (fm *FailoverManager) GetProvider(name string) (FailoverProvider, error) {
	provider, exists := fm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (fm *FailoverManager) GetDefaultProvider() (FailoverProvider, error) {
	if fm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range fm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return fm.GetProvider(fm.config.DefaultProvider)
}

// RegisterEndpoint registers an endpoint using the default provider
func (fm *FailoverManager) RegisterEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	return fm.RegisterEndpointWithProvider(ctx, "", endpoint)
}

// RegisterEndpointWithProvider registers an endpoint using a specific provider
func (fm *FailoverManager) RegisterEndpointWithProvider(ctx context.Context, providerName string, endpoint *types.ServiceEndpoint) error {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return err
	}

	return fm.executeWithRetry(ctx, func() error {
		return provider.RegisterEndpoint(ctx, endpoint)
	})
}

// DeregisterEndpoint deregisters an endpoint using the default provider
func (fm *FailoverManager) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	return fm.DeregisterEndpointWithProvider(ctx, "", endpointID)
}

// DeregisterEndpointWithProvider deregisters an endpoint using a specific provider
func (fm *FailoverManager) DeregisterEndpointWithProvider(ctx context.Context, providerName string, endpointID string) error {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return err
	}

	return fm.executeWithRetry(ctx, func() error {
		return provider.DeregisterEndpoint(ctx, endpointID)
	})
}

// UpdateEndpoint updates an endpoint using the default provider
func (fm *FailoverManager) UpdateEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	return fm.UpdateEndpointWithProvider(ctx, "", endpoint)
}

// UpdateEndpointWithProvider updates an endpoint using a specific provider
func (fm *FailoverManager) UpdateEndpointWithProvider(ctx context.Context, providerName string, endpoint *types.ServiceEndpoint) error {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return err
	}

	return fm.executeWithRetry(ctx, func() error {
		return provider.UpdateEndpoint(ctx, endpoint)
	})
}

// GetEndpoint gets an endpoint using the default provider
func (fm *FailoverManager) GetEndpoint(ctx context.Context, endpointID string) (*types.ServiceEndpoint, error) {
	return fm.GetEndpointWithProvider(ctx, "", endpointID)
}

// GetEndpointWithProvider gets an endpoint using a specific provider
func (fm *FailoverManager) GetEndpointWithProvider(ctx context.Context, providerName string, endpointID string) (*types.ServiceEndpoint, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.ServiceEndpoint
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetEndpoint(ctx, endpointID)
		return err
	})

	return result, err
}

// ListEndpoints lists all endpoints using the default provider
func (fm *FailoverManager) ListEndpoints(ctx context.Context) ([]*types.ServiceEndpoint, error) {
	return fm.ListEndpointsWithProvider(ctx, "")
}

// ListEndpointsWithProvider lists all endpoints using a specific provider
func (fm *FailoverManager) ListEndpointsWithProvider(ctx context.Context, providerName string) ([]*types.ServiceEndpoint, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []*types.ServiceEndpoint
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.ListEndpoints(ctx)
		return err
	})

	return result, err
}

// SelectEndpoint selects an endpoint using the default provider
func (fm *FailoverManager) SelectEndpoint(ctx context.Context, config *types.FailoverConfig) (*types.FailoverResult, error) {
	return fm.SelectEndpointWithProvider(ctx, "", config)
}

// SelectEndpointWithProvider selects an endpoint using a specific provider
func (fm *FailoverManager) SelectEndpointWithProvider(ctx context.Context, providerName string, config *types.FailoverConfig) (*types.FailoverResult, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.FailoverResult
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.SelectEndpoint(ctx, config)
		return err
	})

	return result, err
}

// ExecuteWithFailover executes a function with failover using the default provider
func (fm *FailoverManager) ExecuteWithFailover(ctx context.Context, config *types.FailoverConfig, fn func(*types.ServiceEndpoint) (interface{}, error)) (*types.FailoverResult, error) {
	return fm.ExecuteWithFailoverAndProvider(ctx, "", config, fn)
}

// ExecuteWithFailoverAndProvider executes a function with failover using a specific provider
func (fm *FailoverManager) ExecuteWithFailoverAndProvider(ctx context.Context, providerName string, config *types.FailoverConfig, fn func(*types.ServiceEndpoint) (interface{}, error)) (*types.FailoverResult, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.FailoverResult
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.ExecuteWithFailover(ctx, config, fn)
		return err
	})

	return result, err
}

// HealthCheck performs health check on an endpoint using the default provider
func (fm *FailoverManager) HealthCheck(ctx context.Context, endpointID string) (types.HealthStatus, error) {
	return fm.HealthCheckWithProvider(ctx, "", endpointID)
}

// HealthCheckWithProvider performs health check on an endpoint using a specific provider
func (fm *FailoverManager) HealthCheckWithProvider(ctx context.Context, providerName string, endpointID string) (types.HealthStatus, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return types.HealthUnknown, err
	}

	var result types.HealthStatus
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.HealthCheck(ctx, endpointID)
		return err
	})

	return result, err
}

// HealthCheckAll performs health check on all endpoints using the default provider
func (fm *FailoverManager) HealthCheckAll(ctx context.Context) (map[string]types.HealthStatus, error) {
	return fm.HealthCheckAllWithProvider(ctx, "")
}

// HealthCheckAllWithProvider performs health check on all endpoints using a specific provider
func (fm *FailoverManager) HealthCheckAllWithProvider(ctx context.Context, providerName string) (map[string]types.HealthStatus, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result map[string]types.HealthStatus
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.HealthCheckAll(ctx)
		return err
	})

	return result, err
}

// Configure configures failover using the default provider
func (fm *FailoverManager) Configure(ctx context.Context, config *types.FailoverConfig) error {
	return fm.ConfigureWithProvider(ctx, "", config)
}

// ConfigureWithProvider configures failover using a specific provider
func (fm *FailoverManager) ConfigureWithProvider(ctx context.Context, providerName string, config *types.FailoverConfig) error {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return err
	}

	return fm.executeWithRetry(ctx, func() error {
		return provider.Configure(ctx, config)
	})
}

// GetConfig returns failover configuration using the default provider
func (fm *FailoverManager) GetConfig(ctx context.Context) (*types.FailoverConfig, error) {
	return fm.GetConfigWithProvider(ctx, "")
}

// GetConfigWithProvider returns failover configuration using a specific provider
func (fm *FailoverManager) GetConfigWithProvider(ctx context.Context, providerName string) (*types.FailoverConfig, error) {
	provider, err := fm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.FailoverConfig
	err = fm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetConfig(ctx)
		return err
	})

	return result, err
}

// GetStats returns failover statistics from all providers
func (fm *FailoverManager) GetStats(ctx context.Context) (map[string]*types.FailoverStats, error) {
	stats := make(map[string]*types.FailoverStats)

	for name, provider := range fm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			fm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// GetEvents returns failover events from all providers
func (fm *FailoverManager) GetEvents(ctx context.Context, limit int) (map[string][]*types.FailoverEvent, error) {
	events := make(map[string][]*types.FailoverEvent)

	for name, provider := range fm.providers {
		eventList, err := provider.GetEvents(ctx, limit)
		if err != nil {
			fm.logger.WithError(err).WithField("provider", name).Warn("Failed to get events from provider")
			continue
		}
		events[name] = eventList
	}

	return events, nil
}

// Close closes all providers
func (fm *FailoverManager) Close() error {
	var lastErr error

	for name, provider := range fm.providers {
		if err := provider.Close(context.Background()); err != nil {
			fm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (fm *FailoverManager) getProvider(name string) (FailoverProvider, error) {
	if name == "" {
		return fm.GetDefaultProvider()
	}
	return fm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (fm *FailoverManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= fm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(fm.config.RetryDelay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			fm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Failover operation failed, retrying")
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", fm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (fm *FailoverManager) ListProviders() []string {
	providers := make([]string, 0, len(fm.providers))
	for name := range fm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (fm *FailoverManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range fm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
