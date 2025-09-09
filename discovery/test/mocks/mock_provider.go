package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
)

// MockDiscoveryProvider is a mock implementation of DiscoveryProvider for testing
type MockDiscoveryProvider struct {
	Name              string
	SupportedFeatures []types.ServiceFeature
	ConnectionInfo    *types.ConnectionInfo
	IsConnected       bool
	Services          map[string]*types.Service
	Stats             *types.DiscoveryStats
	Events            chan *types.ServiceEvent
	ErrorOnConnect    error
	ErrorOnRegister   error
	ErrorOnDiscover   error
	ErrorOnHealth     error
	ErrorOnWatch      error
	mutex             sync.RWMutex
}

// NewMockDiscoveryProvider creates a new mock discovery provider
func NewMockDiscoveryProvider(name string) *MockDiscoveryProvider {
	return &MockDiscoveryProvider{
		Name: name,
		SupportedFeatures: []types.ServiceFeature{
			types.FeatureRegister,
			types.FeatureDeregister,
			types.FeatureDiscover,
			types.FeatureHealth,
			types.FeatureWatch,
			types.FeatureTags,
			types.FeatureMetadata,
		},
		ConnectionInfo: &types.ConnectionInfo{
			Host:     "mock-host",
			Port:     8080,
			Protocol: "mock",
			Status:   types.StatusDisconnected,
		},
		IsConnected: false,
		Services:    make(map[string]*types.Service),
		Stats: &types.DiscoveryStats{
			ServicesRegistered: 0,
			InstancesActive:    0,
			QueriesProcessed:   0,
			Uptime:             0,
			LastUpdate:         time.Now(),
			Provider:           name,
		},
		Events: make(chan *types.ServiceEvent, 100),
	}
}

// GetName returns the provider name
func (m *MockDiscoveryProvider) GetName() string {
	return m.Name
}

// GetSupportedFeatures returns the features supported by this provider
func (m *MockDiscoveryProvider) GetSupportedFeatures() []types.ServiceFeature {
	return m.SupportedFeatures
}

// GetConnectionInfo returns connection information
func (m *MockDiscoveryProvider) GetConnectionInfo() *types.ConnectionInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	info := *m.ConnectionInfo
	if m.IsConnected {
		info.Status = types.StatusConnected
	} else {
		info.Status = types.StatusDisconnected
	}
	return &info
}

// Connect establishes connection
func (m *MockDiscoveryProvider) Connect(ctx context.Context) error {
	if m.ErrorOnConnect != nil {
		return m.ErrorOnConnect
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.IsConnected = true
	m.ConnectionInfo.Status = types.StatusConnected
	return nil
}

// Disconnect closes the connection
func (m *MockDiscoveryProvider) Disconnect(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.IsConnected = false
	m.ConnectionInfo.Status = types.StatusDisconnected
	close(m.Events)
	return nil
}

// Ping checks if the connection is alive
func (m *MockDiscoveryProvider) Ping(ctx context.Context) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.IsConnected {
		return &types.DiscoveryError{
			Code:    types.ErrCodeConnection,
			Message: "not connected",
		}
	}
	return nil
}

// IsConnected returns the connection status
func (m *MockDiscoveryProvider) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.IsConnected
}

// RegisterService registers a service
func (m *MockDiscoveryProvider) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	if m.ErrorOnRegister != nil {
		return m.ErrorOnRegister
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

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

	if m.Services[registration.Name] == nil {
		m.Services[registration.Name] = &types.Service{
			Name:      registration.Name,
			Instances: make([]*types.ServiceInstance, 0),
			Tags:      registration.Tags,
			Metadata:  registration.Metadata,
		}
	}

	// Check if instance already exists
	for i, existingInstance := range m.Services[registration.Name].Instances {
		if existingInstance.ID == registration.ID {
			m.Services[registration.Name].Instances[i] = instance
			m.Stats.ServicesRegistered = int64(len(m.Services))
			return nil
		}
	}

	m.Services[registration.Name].Instances = append(m.Services[registration.Name].Instances, instance)
	m.Stats.ServicesRegistered = int64(len(m.Services))
	m.Stats.InstancesActive++

	// Send event
	select {
	case m.Events <- &types.ServiceEvent{
		Type:      types.EventInstanceAdded,
		Instance:  instance,
		Timestamp: time.Now(),
		Provider:  m.Name,
	}:
	default:
		// Channel might be closed
	}

	return nil
}

// DeregisterService deregisters a service
func (m *MockDiscoveryProvider) DeregisterService(ctx context.Context, serviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for serviceName, service := range m.Services {
		for i, instance := range service.Instances {
			if instance.ID == serviceID {
				// Remove instance
				service.Instances = append(service.Instances[:i], service.Instances[i+1:]...)

				// Remove service if no instances left
				if len(service.Instances) == 0 {
					delete(m.Services, serviceName)
				}

				m.Stats.ServicesRegistered = int64(len(m.Services))
				m.Stats.InstancesActive--

				// Send event
				select {
				case m.Events <- &types.ServiceEvent{
					Type:      types.EventInstanceRemoved,
					Instance:  instance,
					Timestamp: time.Now(),
					Provider:  m.Name,
				}:
				default:
					// Channel might be closed
				}

				return nil
			}
		}
	}

	return &types.DiscoveryError{
		Code:    types.ErrCodeServiceNotFound,
		Message: "service not found",
		Service: serviceID,
	}
}

// UpdateService updates a service registration
func (m *MockDiscoveryProvider) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	// In mock, updating is the same as registering
	return m.RegisterService(ctx, registration)
}

// DiscoverServices discovers services
func (m *MockDiscoveryProvider) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	if m.ErrorOnDiscover != nil {
		return nil, m.ErrorOnDiscover
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]*types.Service, 0)

	for _, service := range m.Services {
		// Filter by service name if specified
		if query.Name != "" && service.Name != query.Name {
			continue
		}

		// Filter instances by query criteria
		filteredInstances := make([]*types.ServiceInstance, 0)
		for _, instance := range service.Instances {
			if m.matchesQuery(instance, query) {
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

	m.Stats.QueriesProcessed++
	return result, nil
}

// GetService gets a specific service
func (m *MockDiscoveryProvider) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	query := &types.ServiceQuery{
		Name: serviceName,
	}

	services, err := m.DiscoverServices(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, &types.DiscoveryError{
			Code:    types.ErrCodeServiceNotFound,
			Message: "service not found",
			Service: serviceName,
		}
	}

	return services[0], nil
}

// GetServiceInstance gets a specific service instance
func (m *MockDiscoveryProvider) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	service, exists := m.Services[serviceName]
	if !exists {
		return nil, &types.DiscoveryError{
			Code:    types.ErrCodeServiceNotFound,
			Message: "service not found",
			Service: serviceName,
		}
	}

	for _, instance := range service.Instances {
		if instance.ID == instanceID {
			// Return a copy to avoid race conditions
			instanceCopy := *instance
			return &instanceCopy, nil
		}
	}

	return nil, &types.DiscoveryError{
		Code:    types.ErrCodeInstanceNotFound,
		Message: "service instance not found",
		Service: serviceName,
	}
}

// SetHealth sets the health status of a service
func (m *MockDiscoveryProvider) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	if m.ErrorOnHealth != nil {
		return m.ErrorOnHealth
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, service := range m.Services {
		for _, instance := range service.Instances {
			if instance.ID == serviceID {
				instance.Health = health
				instance.LastUpdated = time.Now()

				// Send event
				select {
				case m.Events <- &types.ServiceEvent{
					Type:      types.EventHealthChanged,
					Instance:  instance,
					Timestamp: time.Now(),
					Provider:  m.Name,
				}:
				default:
					// Channel might be closed
				}

				return nil
			}
		}
	}

	return &types.DiscoveryError{
		Code:    types.ErrCodeServiceNotFound,
		Message: "service not found",
		Service: serviceID,
	}
}

// GetHealth gets the health status of a service
func (m *MockDiscoveryProvider) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, service := range m.Services {
		for _, instance := range service.Instances {
			if instance.ID == serviceID {
				return instance.Health, nil
			}
		}
	}

	return types.HealthUnknown, &types.DiscoveryError{
		Code:    types.ErrCodeServiceNotFound,
		Message: "service not found",
		Service: serviceID,
	}
}

// WatchServices watches for service changes
func (m *MockDiscoveryProvider) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	if m.ErrorOnWatch != nil {
		return nil, m.ErrorOnWatch
	}

	// Return the events channel
	return m.Events, nil
}

// StopWatch stops watching for service changes
func (m *MockDiscoveryProvider) StopWatch(ctx context.Context, watchID string) error {
	// Mock doesn't need to do anything
	return nil
}

// GetStats returns discovery statistics
func (m *MockDiscoveryProvider) GetStats(ctx context.Context) (*types.DiscoveryStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := *m.Stats
	stats.LastUpdate = time.Now()
	return &stats, nil
}

// ListServices returns a list of all service names
func (m *MockDiscoveryProvider) ListServices(ctx context.Context) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	serviceNames := make([]string, 0, len(m.Services))
	for serviceName := range m.Services {
		serviceNames = append(serviceNames, serviceName)
	}

	return serviceNames, nil
}

// matchesQuery checks if a service instance matches the query criteria
func (m *MockDiscoveryProvider) matchesQuery(instance *types.ServiceInstance, query *types.ServiceQuery) bool {
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

// Helper methods for testing

// AddService adds a service to the mock provider
func (m *MockDiscoveryProvider) AddService(registration *types.ServiceRegistration) error {
	return m.RegisterService(context.Background(), registration)
}

// RemoveService removes a service from the mock provider
func (m *MockDiscoveryProvider) RemoveService(serviceID string) error {
	return m.DeregisterService(context.Background(), serviceID)
}

// GetServices returns all services (for testing)
func (m *MockDiscoveryProvider) GetServices() map[string]*types.Service {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a deep copy to avoid race conditions
	result := make(map[string]*types.Service)
	for name, service := range m.Services {
		serviceCopy := *service
		instancesCopy := make([]*types.ServiceInstance, len(service.Instances))
		copy(instancesCopy, service.Instances)
		serviceCopy.Instances = instancesCopy
		result[name] = &serviceCopy
	}

	return result
}

// SetErrorOnConnect sets an error to return on connect
func (m *MockDiscoveryProvider) SetErrorOnConnect(err error) {
	m.ErrorOnConnect = err
}

// SetErrorOnRegister sets an error to return on register
func (m *MockDiscoveryProvider) SetErrorOnRegister(err error) {
	m.ErrorOnRegister = err
}

// SetErrorOnDiscover sets an error to return on discover
func (m *MockDiscoveryProvider) SetErrorOnDiscover(err error) {
	m.ErrorOnDiscover = err
}

// SetErrorOnHealth sets an error to return on health operations
func (m *MockDiscoveryProvider) SetErrorOnHealth(err error) {
	m.ErrorOnHealth = err
}

// SetErrorOnWatch sets an error to return on watch
func (m *MockDiscoveryProvider) SetErrorOnWatch(err error) {
	m.ErrorOnWatch = err
}

// ClearErrors clears all error settings
func (m *MockDiscoveryProvider) ClearErrors() {
	m.ErrorOnConnect = nil
	m.ErrorOnRegister = nil
	m.ErrorOnDiscover = nil
	m.ErrorOnHealth = nil
	m.ErrorOnWatch = nil
}
