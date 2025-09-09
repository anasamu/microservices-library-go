package static

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
)

// StaticProvider implements service discovery using static configuration
type StaticProvider struct {
	services  map[string]*types.Service
	config    *StaticConfig
	logger    *logrus.Logger
	connected bool
	mutex     sync.RWMutex
}

// StaticConfig holds static configuration
type StaticConfig struct {
	Services []*types.ServiceRegistration `json:"services"`
	Timeout  time.Duration                `json:"timeout"`
	Metadata map[string]string            `json:"metadata"`
}

// NewStaticProvider creates a new static provider
func NewStaticProvider(config *StaticConfig, logger *logrus.Logger) (*StaticProvider, error) {
	if config == nil {
		config = &StaticConfig{
			Services: make([]*types.ServiceRegistration, 0),
			Timeout:  30 * time.Second,
		}
	}

	provider := &StaticProvider{
		services:  make(map[string]*types.Service),
		config:    config,
		logger:    logger,
		connected: false,
	}

	// Initialize services from configuration
	for _, registration := range config.Services {
		instance := &types.ServiceInstance{
			ID:          registration.ID,
			Name:        registration.Name,
			Address:     registration.Address,
			Port:        registration.Port,
			Protocol:    registration.Protocol,
			Tags:        registration.Tags,
			Metadata:    registration.Metadata,
			Health:      registration.Health,
			LastUpdated: time.Now(),
		}

		provider.mutex.Lock()
		if provider.services[registration.Name] == nil {
			provider.services[registration.Name] = &types.Service{
				Name:      registration.Name,
				Instances: make([]*types.ServiceInstance, 0),
				Tags:      registration.Tags,
				Metadata:  registration.Metadata,
			}
		}
		provider.services[registration.Name].Instances = append(provider.services[registration.Name].Instances, instance)
		provider.mutex.Unlock()
	}

	return provider, nil
}

// GetName returns the provider name
func (sp *StaticProvider) GetName() string {
	return "static"
}

// GetSupportedFeatures returns the features supported by this provider
func (sp *StaticProvider) GetSupportedFeatures() []types.ServiceFeature {
	return []types.ServiceFeature{
		types.FeatureRegister,
		types.FeatureDeregister,
		types.FeatureDiscover,
		types.FeatureHealth,
		types.FeatureTags,
		types.FeatureMetadata,
	}
}

// GetConnectionInfo returns connection information
func (sp *StaticProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusConnected // Static provider is always "connected"
	return &types.ConnectionInfo{
		Host:     "static",
		Protocol: "static",
		Status:   status,
		Metadata: sp.config.Metadata,
	}
}

// Connect establishes connection (no-op for static provider)
func (sp *StaticProvider) Connect(ctx context.Context) error {
	sp.connected = true
	sp.logger.Info("Static provider connected")
	return nil
}

// Disconnect closes the connection (no-op for static provider)
func (sp *StaticProvider) Disconnect(ctx context.Context) error {
	sp.connected = false
	sp.logger.Info("Static provider disconnected")
	return nil
}

// Ping checks if the provider is available (always true for static)
func (sp *StaticProvider) Ping(ctx context.Context) error {
	return nil
}

// IsConnected returns the connection status
func (sp *StaticProvider) IsConnected() bool {
	return sp.connected
}

// RegisterService registers a service with the static provider
func (sp *StaticProvider) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	ctx, cancel := context.WithTimeout(ctx, sp.config.Timeout)
	defer cancel()

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	instance := &types.ServiceInstance{
		ID:          registration.ID,
		Name:        registration.Name,
		Address:     registration.Address,
		Port:        registration.Port,
		Protocol:    registration.Protocol,
		Tags:        registration.Tags,
		Metadata:    registration.Metadata,
		Health:      registration.Health,
		LastUpdated: time.Now(),
	}

	if sp.services[registration.Name] == nil {
		sp.services[registration.Name] = &types.Service{
			Name:      registration.Name,
			Instances: make([]*types.ServiceInstance, 0),
			Tags:      registration.Tags,
			Metadata:  registration.Metadata,
		}
	}

	// Check if instance already exists
	for i, existingInstance := range sp.services[registration.Name].Instances {
		if existingInstance.ID == registration.ID {
			// Update existing instance
			sp.services[registration.Name].Instances[i] = instance
			sp.logger.WithFields(logrus.Fields{
				"service_id":   registration.ID,
				"service_name": registration.Name,
			}).Info("Service instance updated in static provider")
			return nil
		}
	}

	// Add new instance
	sp.services[registration.Name].Instances = append(sp.services[registration.Name].Instances, instance)
	sp.logger.WithFields(logrus.Fields{
		"service_id":   registration.ID,
		"service_name": registration.Name,
	}).Info("Service instance registered with static provider")

	return nil
}

// DeregisterService deregisters a service from the static provider
func (sp *StaticProvider) DeregisterService(ctx context.Context, serviceID string) error {
	ctx, cancel := context.WithTimeout(ctx, sp.config.Timeout)
	defer cancel()

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for serviceName, service := range sp.services {
		for i, instance := range service.Instances {
			if instance.ID == serviceID {
				// Remove instance
				service.Instances = append(service.Instances[:i], service.Instances[i+1:]...)

				// Remove service if no instances left
				if len(service.Instances) == 0 {
					delete(sp.services, serviceName)
				}

				sp.logger.WithField("service_id", serviceID).Info("Service instance deregistered from static provider")
				return nil
			}
		}
	}

	return fmt.Errorf("service %s not found", serviceID)
}

// UpdateService updates a service registration in the static provider
func (sp *StaticProvider) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	// In static provider, updating is the same as registering
	return sp.RegisterService(ctx, registration)
}

// DiscoverServices discovers services from the static provider
func (sp *StaticProvider) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, sp.config.Timeout)
	defer cancel()

	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	result := make([]*types.Service, 0)

	for _, service := range sp.services {
		// Filter by service name if specified
		if query.Name != "" && service.Name != query.Name {
			continue
		}

		// Filter instances by query criteria
		filteredInstances := make([]*types.ServiceInstance, 0)
		for _, instance := range service.Instances {
			if sp.matchesQuery(instance, query) {
				filteredInstances = append(filteredInstances, instance)
			}
		}

		// Apply limit if specified
		if query.Limit > 0 && len(filteredInstances) > query.Limit {
			filteredInstances = filteredInstances[:query.Limit]
		}

		if len(filteredInstances) > 0 {
			serviceCopy := *service
			serviceCopy.Instances = filteredInstances
			result = append(result, &serviceCopy)
		}
	}

	return result, nil
}

// GetService gets a specific service from the static provider
func (sp *StaticProvider) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	service, exists := sp.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Return a copy to avoid race conditions
	serviceCopy := *service
	instancesCopy := make([]*types.ServiceInstance, len(service.Instances))
	copy(instancesCopy, service.Instances)
	serviceCopy.Instances = instancesCopy

	return &serviceCopy, nil
}

// GetServiceInstance gets a specific service instance from the static provider
func (sp *StaticProvider) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	service, exists := sp.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	for _, instance := range service.Instances {
		if instance.ID == instanceID {
			// Return a copy to avoid race conditions
			instanceCopy := *instance
			return &instanceCopy, nil
		}
	}

	return nil, fmt.Errorf("service instance %s not found for service %s", instanceID, serviceName)
}

// SetHealth sets the health status of a service in the static provider
func (sp *StaticProvider) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	ctx, cancel := context.WithTimeout(ctx, sp.config.Timeout)
	defer cancel()

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, service := range sp.services {
		for _, instance := range service.Instances {
			if instance.ID == serviceID {
				instance.Health = health
				instance.LastUpdated = time.Now()
				sp.logger.WithFields(logrus.Fields{
					"service_id": serviceID,
					"health":     health,
				}).Info("Service health updated in static provider")
				return nil
			}
		}
	}

	return fmt.Errorf("service %s not found", serviceID)
}

// GetHealth gets the health status of a service from the static provider
func (sp *StaticProvider) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	for _, service := range sp.services {
		for _, instance := range service.Instances {
			if instance.ID == serviceID {
				return instance.Health, nil
			}
		}
	}

	return types.HealthUnknown, fmt.Errorf("service %s not found", serviceID)
}

// WatchServices watches for service changes in the static provider
func (sp *StaticProvider) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	// Static provider doesn't support watching since services are static
	// Return a channel that will be closed immediately
	eventChan := make(chan *types.ServiceEvent)
	close(eventChan)

	sp.logger.Warn("WatchServices is not supported by static provider")
	return eventChan, fmt.Errorf("watching is not supported by static provider")
}

// StopWatch stops watching for service changes (not applicable for static provider)
func (sp *StaticProvider) StopWatch(ctx context.Context, watchID string) error {
	return fmt.Errorf("watching is not supported by static provider")
}

// GetStats returns discovery statistics from the static provider
func (sp *StaticProvider) GetStats(ctx context.Context) (*types.DiscoveryStats, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	servicesCount := int64(len(sp.services))
	instancesCount := int64(0)

	for _, service := range sp.services {
		instancesCount += int64(len(service.Instances))
	}

	return &types.DiscoveryStats{
		ServicesRegistered: servicesCount,
		InstancesActive:    instancesCount,
		QueriesProcessed:   0, // Static provider doesn't track this
		Uptime:             0, // Would need to track this separately
		LastUpdate:         time.Now(),
		Provider:           sp.GetName(),
	}, nil
}

// ListServices returns a list of all service names from the static provider
func (sp *StaticProvider) ListServices(ctx context.Context) ([]string, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	serviceNames := make([]string, 0, len(sp.services))
	for serviceName := range sp.services {
		serviceNames = append(serviceNames, serviceName)
	}

	return serviceNames, nil
}

// matchesQuery checks if a service instance matches the query criteria
func (sp *StaticProvider) matchesQuery(instance *types.ServiceInstance, query *types.ServiceQuery) bool {
	// Filter by health status
	if query.Health != "" && instance.Health != query.Health {
		return false
	}

	// Filter by tags
	if len(query.Tags) > 0 {
		instanceTags := make(map[string]bool)
		for _, tag := range instance.Tags {
			instanceTags[tag] = true
		}

		for _, queryTag := range query.Tags {
			if !instanceTags[queryTag] {
				return false
			}
		}
	}

	// Filter by metadata
	if len(query.Metadata) > 0 {
		for key, value := range query.Metadata {
			if instance.Metadata[key] != value {
				return false
			}
		}
	}

	return true
}

// AddService adds a service to the static configuration
func (sp *StaticProvider) AddService(registration *types.ServiceRegistration) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	instance := &types.ServiceInstance{
		ID:          registration.ID,
		Name:        registration.Name,
		Address:     registration.Address,
		Port:        registration.Port,
		Protocol:    registration.Protocol,
		Tags:        registration.Tags,
		Metadata:    registration.Metadata,
		Health:      registration.Health,
		LastUpdated: time.Now(),
	}

	if sp.services[registration.Name] == nil {
		sp.services[registration.Name] = &types.Service{
			Name:      registration.Name,
			Instances: make([]*types.ServiceInstance, 0),
			Tags:      registration.Tags,
			Metadata:  registration.Metadata,
		}
	}

	sp.services[registration.Name].Instances = append(sp.services[registration.Name].Instances, instance)
	return nil
}

// RemoveService removes a service from the static configuration
func (sp *StaticProvider) RemoveService(serviceName string) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	delete(sp.services, serviceName)
	return nil
}

// GetServices returns all services (for testing/debugging)
func (sp *StaticProvider) GetServices() map[string]*types.Service {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	// Return a deep copy to avoid race conditions
	result := make(map[string]*types.Service)
	for name, service := range sp.services {
		serviceCopy := *service
		instancesCopy := make([]*types.ServiceInstance, len(service.Instances))
		copy(instancesCopy, service.Instances)
		serviceCopy.Instances = instancesCopy
		result[name] = &serviceCopy
	}

	return result
}
