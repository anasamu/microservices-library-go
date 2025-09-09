package consul

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// ConsulProvider implements service discovery using HashiCorp Consul
type ConsulProvider struct {
	client    *api.Client
	config    *ConsulConfig
	logger    *logrus.Logger
	connected bool
}

// ConsulConfig holds Consul-specific configuration
type ConsulConfig struct {
	Address    string            `json:"address"`
	Token      string            `json:"token"`
	Datacenter string            `json:"datacenter"`
	Namespace  string            `json:"namespace"`
	Timeout    time.Duration     `json:"timeout"`
	Metadata   map[string]string `json:"metadata"`
}

// NewConsulProvider creates a new Consul provider
func NewConsulProvider(config *ConsulConfig, logger *logrus.Logger) (*ConsulProvider, error) {
	if config == nil {
		config = &ConsulConfig{
			Address: "localhost:8500",
			Timeout: 30 * time.Second,
		}
	}

	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address
	consulConfig.Token = config.Token
	consulConfig.Datacenter = config.Datacenter
	consulConfig.Namespace = config.Namespace

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &ConsulProvider{
		client:    client,
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (cp *ConsulProvider) GetName() string {
	return "consul"
}

// GetSupportedFeatures returns the features supported by this provider
func (cp *ConsulProvider) GetSupportedFeatures() []types.ServiceFeature {
	return []types.ServiceFeature{
		types.FeatureRegister,
		types.FeatureDeregister,
		types.FeatureDiscover,
		types.FeatureHealth,
		types.FeatureWatch,
		types.FeatureLoadBalancing,
		types.FeatureFailover,
		types.FeatureTags,
		types.FeatureMetadata,
		types.FeatureTTL,
		types.FeatureClustering,
		types.FeatureConsistency,
		types.FeatureSecurity,
	}
}

// GetConnectionInfo returns connection information
func (cp *ConsulProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if cp.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     cp.config.Address,
		Protocol: "http",
		Status:   status,
		Metadata: cp.config.Metadata,
	}
}

// Connect establishes connection to Consul
func (cp *ConsulProvider) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	// Test connection by getting the agent info
	_, err := cp.client.Agent().Self()
	if err != nil {
		cp.connected = false
		return fmt.Errorf("failed to connect to Consul: %w", err)
	}

	cp.connected = true
	cp.logger.Info("Connected to Consul")
	return nil
}

// Disconnect closes the connection to Consul
func (cp *ConsulProvider) Disconnect(ctx context.Context) error {
	cp.connected = false
	cp.logger.Info("Disconnected from Consul")
	return nil
}

// Ping checks if the connection to Consul is alive
func (cp *ConsulProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	_, err := cp.client.Agent().Self()
	if err != nil {
		cp.connected = false
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// IsConnected returns the connection status
func (cp *ConsulProvider) IsConnected() bool {
	return cp.connected
}

// RegisterService registers a service with Consul
func (cp *ConsulProvider) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      registration.ID,
		Name:    registration.Name,
		Address: registration.Address,
		Port:    registration.Port,
		Tags:    registration.Tags,
		Meta:    registration.Metadata,
	}

	// Add health check if TTL is specified
	if registration.TTL > 0 {
		serviceRegistration.Check = &api.AgentServiceCheck{
			TTL: registration.TTL.String(),
		}
	}

	err := cp.client.Agent().ServiceRegister(serviceRegistration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	cp.logger.WithFields(logrus.Fields{
		"service_id":   registration.ID,
		"service_name": registration.Name,
	}).Info("Service registered with Consul")

	return nil
}

// DeregisterService deregisters a service from Consul
func (cp *ConsulProvider) DeregisterService(ctx context.Context, serviceID string) error {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	err := cp.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	cp.logger.WithField("service_id", serviceID).Info("Service deregistered from Consul")
	return nil
}

// UpdateService updates a service registration in Consul
func (cp *ConsulProvider) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	// In Consul, updating is the same as registering
	return cp.RegisterService(ctx, registration)
}

// DiscoverServices discovers services from Consul
func (cp *ConsulProvider) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	services, _, err := cp.client.Health().Service(query.Name, "", false, &api.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	result := make([]*types.Service, 0, len(services))
	serviceMap := make(map[string]*types.Service)

	for _, service := range services {
		instance := &types.ServiceInstance{
			ID:          service.Service.ID,
			Name:        service.Service.Service,
			Address:     service.Service.Address,
			Port:        service.Service.Port,
			Tags:        service.Service.Tags,
			Metadata:    service.Service.Meta,
			Health:      convertConsulHealth(service.Checks.AggregatedStatus()),
			LastUpdated: time.Now(),
		}

		if serviceMap[service.Service.Service] == nil {
			serviceMap[service.Service.Service] = &types.Service{
				Name:      service.Service.Service,
				Instances: make([]*types.ServiceInstance, 0),
				Tags:      service.Service.Tags,
				Metadata:  service.Service.Meta,
			}
		}

		serviceMap[service.Service.Service].Instances = append(serviceMap[service.Service.Service].Instances, instance)
	}

	for _, service := range serviceMap {
		result = append(result, service)
	}

	return result, nil
}

// GetService gets a specific service from Consul
func (cp *ConsulProvider) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	query := &types.ServiceQuery{
		Name: serviceName,
	}

	services, err := cp.DiscoverServices(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return services[0], nil
}

// GetServiceInstance gets a specific service instance from Consul
func (cp *ConsulProvider) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	services, _, err := cp.client.Health().Service(serviceName, "", false, &api.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service instance: %w", err)
	}

	for _, service := range services {
		if service.Service.ID == instanceID {
			return &types.ServiceInstance{
				ID:          service.Service.ID,
				Name:        service.Service.Service,
				Address:     service.Service.Address,
				Port:        service.Service.Port,
				Tags:        service.Service.Tags,
				Metadata:    service.Service.Meta,
				Health:      convertConsulHealth(service.Checks.AggregatedStatus()),
				LastUpdated: time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("service instance %s not found for service %s", instanceID, serviceName)
}

// SetHealth sets the health status of a service in Consul
func (cp *ConsulProvider) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	consulHealth := convertToConsulHealth(health)
	err := cp.client.Agent().UpdateTTL("service:"+serviceID, "", consulHealth)
	if err != nil {
		return fmt.Errorf("failed to set health: %w", err)
	}

	return nil
}

// GetHealth gets the health status of a service from Consul
func (cp *ConsulProvider) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	checks, _, err := cp.client.Health().Checks(serviceID, &api.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return types.HealthUnknown, fmt.Errorf("failed to get health: %w", err)
	}

	if len(checks) == 0 {
		return types.HealthUnknown, nil
	}

	return convertConsulHealth(checks[0].Status), nil
}

// WatchServices watches for service changes in Consul
func (cp *ConsulProvider) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	eventChan := make(chan *types.ServiceEvent, 100)

	go func() {
		defer close(eventChan)

		queryOptions := &api.QueryOptions{
			Context: ctx,
		}

		if options.Timeout > 0 {
			queryOptions.WaitTime = options.Timeout
		}

		// Watch for service changes
		lastIndex := uint64(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				services, meta, err := cp.client.Health().Service(options.ServiceName, "", false, queryOptions)
				if err != nil {
					cp.logger.WithError(err).Error("Failed to watch services")
					return
				}

				if meta.LastIndex > lastIndex {
					lastIndex = meta.LastIndex

					// Convert to service events
					for _, service := range services {
						event := &types.ServiceEvent{
							Type: types.EventInstanceUpdated,
							Instance: &types.ServiceInstance{
								ID:          service.Service.ID,
								Name:        service.Service.Service,
								Address:     service.Service.Address,
								Port:        service.Service.Port,
								Tags:        service.Service.Tags,
								Metadata:    service.Service.Meta,
								Health:      convertConsulHealth(service.Checks.AggregatedStatus()),
								LastUpdated: time.Now(),
							},
							Timestamp: time.Now(),
							Provider:  cp.GetName(),
						}

						select {
						case eventChan <- event:
						case <-ctx.Done():
							return
						}
					}
				}

				queryOptions.WaitIndex = meta.LastIndex
			}
		}
	}()

	return eventChan, nil
}

// StopWatch stops watching for service changes (not applicable for Consul)
func (cp *ConsulProvider) StopWatch(ctx context.Context, watchID string) error {
	// Consul doesn't have a specific stop watch mechanism
	// The context cancellation handles this
	return nil
}

// GetStats returns discovery statistics from Consul
func (cp *ConsulProvider) GetStats(ctx context.Context) (*types.DiscoveryStats, error) {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	// Get agent info for basic stats
	agent, err := cp.client.Agent().Self()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}

	// Get services count
	services, _, err := cp.client.Catalog().Services(&api.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	servicesCount := int64(len(services))
	instancesCount := int64(0)

	// Count instances
	for serviceName := range services {
		serviceInstances, _, err := cp.client.Health().Service(serviceName, "", false, &api.QueryOptions{
			Context: ctx,
		})
		if err != nil {
			continue
		}
		instancesCount += int64(len(serviceInstances))
	}

	return &types.DiscoveryStats{
		ServicesRegistered: servicesCount,
		InstancesActive:    instancesCount,
		QueriesProcessed:   0, // Consul doesn't provide this metric
		Uptime:             0, // Would need to track this separately
		LastUpdate:         time.Now(),
		Provider:           cp.GetName(),
	}, nil
}

// ListServices returns a list of all service names from Consul
func (cp *ConsulProvider) ListServices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, cp.config.Timeout)
	defer cancel()

	services, _, err := cp.client.Catalog().Services(&api.QueryOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	serviceNames := make([]string, 0, len(services))
	for serviceName := range services {
		serviceNames = append(serviceNames, serviceName)
	}

	return serviceNames, nil
}

// convertConsulHealth converts Consul health status to our health status
func convertConsulHealth(status string) types.HealthStatus {
	switch status {
	case api.HealthPassing:
		return types.HealthPassing
	case api.HealthWarning:
		return types.HealthWarning
	case api.HealthCritical:
		return types.HealthCritical
	default:
		return types.HealthUnknown
	}
}

// convertToConsulHealth converts our health status to Consul health status
func convertToConsulHealth(health types.HealthStatus) string {
	switch health {
	case types.HealthPassing:
		return api.HealthPassing
	case types.HealthWarning:
		return api.HealthWarning
	case types.HealthCritical:
		return api.HealthCritical
	default:
		return api.HealthCritical
	}
}
