package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
)

// DiscoveryManager manages multiple service discovery providers
type DiscoveryManager struct {
	providers map[string]DiscoveryProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds discovery manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// DiscoveryProvider interface for service discovery backends
type DiscoveryProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.ServiceFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Service registration and deregistration
	RegisterService(ctx context.Context, registration *types.ServiceRegistration) error
	DeregisterService(ctx context.Context, serviceID string) error
	UpdateService(ctx context.Context, registration *types.ServiceRegistration) error

	// Service discovery
	DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error)
	GetService(ctx context.Context, serviceName string) (*types.Service, error)
	GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error)

	// Health management
	SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error
	GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error)

	// Service watching
	WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error)
	StopWatch(ctx context.Context, watchID string) error

	// Statistics and monitoring
	GetStats(ctx context.Context) (*types.DiscoveryStats, error)
	ListServices(ctx context.Context) ([]string, error)
}

// NewDiscoveryManager creates a new discovery manager
func NewDiscoveryManager(config *ManagerConfig, logger *logrus.Logger) *DiscoveryManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &DiscoveryManager{
		providers: make(map[string]DiscoveryProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a discovery provider
func (dm *DiscoveryManager) RegisterProvider(provider DiscoveryProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := dm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	dm.providers[name] = provider
	dm.logger.WithField("provider", name).Info("Discovery provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (dm *DiscoveryManager) GetProvider(name string) (DiscoveryProvider, error) {
	provider, exists := dm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (dm *DiscoveryManager) GetDefaultProvider() (DiscoveryProvider, error) {
	if dm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range dm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return dm.GetProvider(dm.config.DefaultProvider)
}

// RegisterService registers a service using the default provider
func (dm *DiscoveryManager) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	return dm.RegisterServiceWithProvider(ctx, "", registration)
}

// RegisterServiceWithProvider registers a service using a specific provider
func (dm *DiscoveryManager) RegisterServiceWithProvider(ctx context.Context, providerName string, registration *types.ServiceRegistration) error {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return err
	}

	return dm.executeWithRetry(ctx, func() error {
		return provider.RegisterService(ctx, registration)
	})
}

// DeregisterService deregisters a service using the default provider
func (dm *DiscoveryManager) DeregisterService(ctx context.Context, serviceID string) error {
	return dm.DeregisterServiceWithProvider(ctx, "", serviceID)
}

// DeregisterServiceWithProvider deregisters a service using a specific provider
func (dm *DiscoveryManager) DeregisterServiceWithProvider(ctx context.Context, providerName string, serviceID string) error {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return err
	}

	return dm.executeWithRetry(ctx, func() error {
		return provider.DeregisterService(ctx, serviceID)
	})
}

// UpdateService updates a service using the default provider
func (dm *DiscoveryManager) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	return dm.UpdateServiceWithProvider(ctx, "", registration)
}

// UpdateServiceWithProvider updates a service using a specific provider
func (dm *DiscoveryManager) UpdateServiceWithProvider(ctx context.Context, providerName string, registration *types.ServiceRegistration) error {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return err
	}

	return dm.executeWithRetry(ctx, func() error {
		return provider.UpdateService(ctx, registration)
	})
}

// DiscoverServices discovers services using the default provider
func (dm *DiscoveryManager) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	return dm.DiscoverServicesWithProvider(ctx, "", query)
}

// DiscoverServicesWithProvider discovers services using a specific provider
func (dm *DiscoveryManager) DiscoverServicesWithProvider(ctx context.Context, providerName string, query *types.ServiceQuery) ([]*types.Service, error) {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []*types.Service
	err = dm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.DiscoverServices(ctx, query)
		return err
	})

	return result, err
}

// GetService gets a specific service using the default provider
func (dm *DiscoveryManager) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	return dm.GetServiceWithProvider(ctx, "", serviceName)
}

// GetServiceWithProvider gets a specific service using a specific provider
func (dm *DiscoveryManager) GetServiceWithProvider(ctx context.Context, providerName string, serviceName string) (*types.Service, error) {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.Service
	err = dm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetService(ctx, serviceName)
		return err
	})

	return result, err
}

// GetServiceInstance gets a specific service instance using the default provider
func (dm *DiscoveryManager) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	return dm.GetServiceInstanceWithProvider(ctx, "", serviceName, instanceID)
}

// GetServiceInstanceWithProvider gets a specific service instance using a specific provider
func (dm *DiscoveryManager) GetServiceInstanceWithProvider(ctx context.Context, providerName string, serviceName, instanceID string) (*types.ServiceInstance, error) {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result *types.ServiceInstance
	err = dm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetServiceInstance(ctx, serviceName, instanceID)
		return err
	})

	return result, err
}

// SetHealth sets the health status of a service using the default provider
func (dm *DiscoveryManager) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	return dm.SetHealthWithProvider(ctx, "", serviceID, health)
}

// SetHealthWithProvider sets the health status of a service using a specific provider
func (dm *DiscoveryManager) SetHealthWithProvider(ctx context.Context, providerName string, serviceID string, health types.HealthStatus) error {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return err
	}

	return dm.executeWithRetry(ctx, func() error {
		return provider.SetHealth(ctx, serviceID, health)
	})
}

// GetHealth gets the health status of a service using the default provider
func (dm *DiscoveryManager) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	return dm.GetHealthWithProvider(ctx, "", serviceID)
}

// GetHealthWithProvider gets the health status of a service using a specific provider
func (dm *DiscoveryManager) GetHealthWithProvider(ctx context.Context, providerName string, serviceID string) (types.HealthStatus, error) {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return types.HealthUnknown, err
	}

	var result types.HealthStatus
	err = dm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.GetHealth(ctx, serviceID)
		return err
	})

	return result, err
}

// WatchServices watches for service changes using the default provider
func (dm *DiscoveryManager) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	return dm.WatchServicesWithProvider(ctx, "", options)
}

// WatchServicesWithProvider watches for service changes using a specific provider
func (dm *DiscoveryManager) WatchServicesWithProvider(ctx context.Context, providerName string, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.WatchServices(ctx, options)
}

// StopWatch stops watching for service changes using the default provider
func (dm *DiscoveryManager) StopWatch(ctx context.Context, watchID string) error {
	return dm.StopWatchWithProvider(ctx, "", watchID)
}

// StopWatchWithProvider stops watching for service changes using a specific provider
func (dm *DiscoveryManager) StopWatchWithProvider(ctx context.Context, providerName string, watchID string) error {
	provider, err := dm.getProvider(providerName)
	if err != nil {
		return err
	}

	return provider.StopWatch(ctx, watchID)
}

// GetStats returns discovery statistics from all providers
func (dm *DiscoveryManager) GetStats(ctx context.Context) (map[string]*types.DiscoveryStats, error) {
	stats := make(map[string]*types.DiscoveryStats)

	for name, provider := range dm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			dm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// ListServices returns a list of all services from all providers
func (dm *DiscoveryManager) ListServices(ctx context.Context) (map[string][]string, error) {
	services := make(map[string][]string)

	for name, provider := range dm.providers {
		serviceList, err := provider.ListServices(ctx)
		if err != nil {
			dm.logger.WithError(err).WithField("provider", name).Warn("Failed to list services from provider")
			continue
		}
		services[name] = serviceList
	}

	return services, nil
}

// Close closes all providers
func (dm *DiscoveryManager) Close() error {
	var lastErr error

	for name, provider := range dm.providers {
		if err := provider.Disconnect(context.Background()); err != nil {
			dm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (dm *DiscoveryManager) getProvider(name string) (DiscoveryProvider, error) {
	if name == "" {
		return dm.GetDefaultProvider()
	}
	return dm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (dm *DiscoveryManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= dm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dm.config.RetryDelay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			dm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Discovery operation failed, retrying")
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", dm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (dm *DiscoveryManager) ListProviders() []string {
	providers := make([]string, 0, len(dm.providers))
	for name := range dm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (dm *DiscoveryManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range dm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
