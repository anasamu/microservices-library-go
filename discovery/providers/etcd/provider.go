package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

// EtcdProvider implements service discovery using etcd
type EtcdProvider struct {
	client    *clientv3.Client
	config    *EtcdConfig
	logger    *logrus.Logger
	connected bool
}

// EtcdConfig holds etcd-specific configuration
type EtcdConfig struct {
	Endpoints []string          `json:"endpoints"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
	Namespace string            `json:"namespace"`
	Timeout   time.Duration     `json:"timeout"`
	Metadata  map[string]string `json:"metadata"`
}

// NewEtcdProvider creates a new etcd provider
func NewEtcdProvider(config *EtcdConfig, logger *logrus.Logger) (*EtcdProvider, error) {
	if config == nil {
		config = &EtcdConfig{
			Endpoints: []string{"localhost:2379"},
			Timeout:   30 * time.Second,
			Namespace: "/services",
		}
	}

	etcdConfig := clientv3.Config{
		Endpoints:   config.Endpoints,
		Username:    config.Username,
		Password:    config.Password,
		DialTimeout: config.Timeout,
	}

	client, err := clientv3.New(etcdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdProvider{
		client:    client,
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (ep *EtcdProvider) GetName() string {
	return "etcd"
}

// GetSupportedFeatures returns the features supported by this provider
func (ep *EtcdProvider) GetSupportedFeatures() []types.ServiceFeature {
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
	}
}

// GetConnectionInfo returns connection information
func (ep *EtcdProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if ep.connected {
		status = types.StatusConnected
	}

	host := strings.Join(ep.config.Endpoints, ",")
	return &types.ConnectionInfo{
		Host:     host,
		Protocol: "etcd",
		Status:   status,
		Metadata: ep.config.Metadata,
	}
}

// Connect establishes connection to etcd
func (ep *EtcdProvider) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Test connection by getting cluster info
	_, err := ep.client.Status(ctx, ep.config.Endpoints[0])
	if err != nil {
		ep.connected = false
		return fmt.Errorf("failed to connect to etcd: %w", err)
	}

	ep.connected = true
	ep.logger.Info("Connected to etcd")
	return nil
}

// Disconnect closes the connection to etcd
func (ep *EtcdProvider) Disconnect(ctx context.Context) error {
	ep.connected = false
	err := ep.client.Close()
	if err != nil {
		ep.logger.WithError(err).Error("Failed to close etcd client")
	}
	ep.logger.Info("Disconnected from etcd")
	return err
}

// Ping checks if the connection to etcd is alive
func (ep *EtcdProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	_, err := ep.client.Status(ctx, ep.config.Endpoints[0])
	if err != nil {
		ep.connected = false
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// IsConnected returns the connection status
func (ep *EtcdProvider) IsConnected() bool {
	return ep.connected
}

// RegisterService registers a service with etcd
func (ep *EtcdProvider) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Create service key
	serviceKey := ep.getServiceKey(registration.Name, registration.ID)

	// Create service instance data
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

	// Serialize instance data
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal service instance: %w", err)
	}

	// Set TTL if specified
	var lease clientv3.LeaseID
	if registration.TTL > 0 {
		leaseResp, err := ep.client.Grant(ctx, int64(registration.TTL.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to create lease: %w", err)
		}
		lease = leaseResp.ID

		// Keep lease alive
		_, err = ep.client.KeepAlive(ctx, lease)
		if err != nil {
			return fmt.Errorf("failed to keep lease alive: %w", err)
		}
	}

	// Put service instance
	_, err = ep.client.Put(ctx, serviceKey, string(data), clientv3.WithLease(lease))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	ep.logger.WithFields(logrus.Fields{
		"service_id":   registration.ID,
		"service_name": registration.Name,
		"service_key":  serviceKey,
	}).Info("Service registered with etcd")

	return nil
}

// DeregisterService deregisters a service from etcd
func (ep *EtcdProvider) DeregisterService(ctx context.Context, serviceID string) error {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Find and delete the service key
	serviceKey := ep.getServiceKeyByID(serviceID)
	if serviceKey == "" {
		return fmt.Errorf("service %s not found", serviceID)
	}

	_, err := ep.client.Delete(ctx, serviceKey)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	ep.logger.WithField("service_id", serviceID).Info("Service deregistered from etcd")
	return nil
}

// UpdateService updates a service registration in etcd
func (ep *EtcdProvider) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	// In etcd, updating is the same as registering
	return ep.RegisterService(ctx, registration)
}

// DiscoverServices discovers services from etcd
func (ep *EtcdProvider) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Build search prefix
	searchPrefix := ep.config.Namespace
	if query.Name != "" {
		searchPrefix = path.Join(ep.config.Namespace, query.Name)
	}

	// Get all services
	resp, err := ep.client.Get(ctx, searchPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	serviceMap := make(map[string]*types.Service)

	for _, kv := range resp.Kvs {
		var instance types.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			ep.logger.WithError(err).WithField("key", string(kv.Key)).Warn("Failed to unmarshal service instance")
			continue
		}

		// Filter by query criteria
		if !ep.matchesQuery(&instance, query) {
			continue
		}

		if serviceMap[instance.Name] == nil {
			serviceMap[instance.Name] = &types.Service{
				Name:      instance.Name,
				Instances: make([]*types.ServiceInstance, 0),
				Tags:      instance.Tags,
				Metadata:  instance.Metadata,
			}
		}

		serviceMap[instance.Name].Instances = append(serviceMap[instance.Name].Instances, &instance)
	}

	result := make([]*types.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		result = append(result, service)
	}

	return result, nil
}

// GetService gets a specific service from etcd
func (ep *EtcdProvider) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	query := &types.ServiceQuery{
		Name: serviceName,
	}

	services, err := ep.DiscoverServices(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return services[0], nil
}

// GetServiceInstance gets a specific service instance from etcd
func (ep *EtcdProvider) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	serviceKey := ep.getServiceKey(serviceName, instanceID)
	resp, err := ep.client.Get(ctx, serviceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get service instance: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("service instance %s not found for service %s", instanceID, serviceName)
	}

	var instance types.ServiceInstance
	if err := json.Unmarshal(resp.Kvs[0].Value, &instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service instance: %w", err)
	}

	return &instance, nil
}

// SetHealth sets the health status of a service in etcd
func (ep *EtcdProvider) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Find the service key
	serviceKey := ep.getServiceKeyByID(serviceID)
	if serviceKey == "" {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Get current service instance
	resp, err := ep.client.Get(ctx, serviceKey)
	if err != nil {
		return fmt.Errorf("failed to get service instance: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("service instance not found")
	}

	var instance types.ServiceInstance
	if err := json.Unmarshal(resp.Kvs[0].Value, &instance); err != nil {
		return fmt.Errorf("failed to unmarshal service instance: %w", err)
	}

	// Update health status
	instance.Health = health
	instance.LastUpdated = time.Now()

	// Serialize and update
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal service instance: %w", err)
	}

	_, err = ep.client.Put(ctx, serviceKey, string(data))
	if err != nil {
		return fmt.Errorf("failed to update health: %w", err)
	}

	return nil
}

// GetHealth gets the health status of a service from etcd
func (ep *EtcdProvider) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	instance, err := ep.GetServiceInstance(ctx, "", serviceID)
	if err != nil {
		return types.HealthUnknown, err
	}

	return instance.Health, nil
}

// WatchServices watches for service changes in etcd
func (ep *EtcdProvider) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	eventChan := make(chan *types.ServiceEvent, 100)

	go func() {
		defer close(eventChan)

		// Build watch prefix
		watchPrefix := ep.config.Namespace
		if options.ServiceName != "" {
			watchPrefix = path.Join(ep.config.Namespace, options.ServiceName)
		}

		watchChan := ep.client.Watch(ctx, watchPrefix, clientv3.WithPrefix())

		for {
			select {
			case <-ctx.Done():
				return
			case watchResp := <-watchChan:
				if watchResp.Err() != nil {
					ep.logger.WithError(watchResp.Err()).Error("Watch error")
					return
				}

				for _, event := range watchResp.Events {
					serviceEvent := &types.ServiceEvent{
						Timestamp: time.Now(),
						Provider:  ep.GetName(),
					}

					switch event.Type {
					case clientv3.EventTypePut:
						var instance types.ServiceInstance
						if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
							ep.logger.WithError(err).Warn("Failed to unmarshal service instance")
							continue
						}

						serviceEvent.Type = types.EventInstanceUpdated
						serviceEvent.Instance = &instance

					case clientv3.EventTypeDelete:
						// Extract service info from key
						keyParts := strings.Split(string(event.Kv.Key), "/")
						if len(keyParts) >= 3 {
							serviceName := keyParts[len(keyParts)-2]
							instanceID := keyParts[len(keyParts)-1]

							serviceEvent.Type = types.EventInstanceRemoved
							serviceEvent.Instance = &types.ServiceInstance{
								ID:   instanceID,
								Name: serviceName,
							}
						}
					}

					select {
					case eventChan <- serviceEvent:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// StopWatch stops watching for service changes (not applicable for etcd)
func (ep *EtcdProvider) StopWatch(ctx context.Context, watchID string) error {
	// etcd doesn't have a specific stop watch mechanism
	// The context cancellation handles this
	return nil
}

// GetStats returns discovery statistics from etcd
func (ep *EtcdProvider) GetStats(ctx context.Context) (*types.DiscoveryStats, error) {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Get all services
	resp, err := ep.client.Get(ctx, ep.config.Namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	servicesCount := int64(0)
	instancesCount := int64(len(resp.Kvs))
	serviceNames := make(map[string]bool)

	for _, kv := range resp.Kvs {
		var instance types.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		serviceNames[instance.Name] = true
	}

	servicesCount = int64(len(serviceNames))

	return &types.DiscoveryStats{
		ServicesRegistered: servicesCount,
		InstancesActive:    instancesCount,
		QueriesProcessed:   0, // etcd doesn't provide this metric
		Uptime:             0, // Would need to track this separately
		LastUpdate:         time.Now(),
		Provider:           ep.GetName(),
	}, nil
}

// ListServices returns a list of all service names from etcd
func (ep *EtcdProvider) ListServices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, ep.config.Timeout)
	defer cancel()

	// Get all services
	resp, err := ep.client.Get(ctx, ep.config.Namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	serviceNames := make(map[string]bool)
	for _, kv := range resp.Kvs {
		var instance types.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		serviceNames[instance.Name] = true
	}

	result := make([]string, 0, len(serviceNames))
	for serviceName := range serviceNames {
		result = append(result, serviceName)
	}

	return result, nil
}

// getServiceKey creates a service key for etcd
func (ep *EtcdProvider) getServiceKey(serviceName, instanceID string) string {
	return path.Join(ep.config.Namespace, serviceName, instanceID)
}

// getServiceKeyByID finds a service key by instance ID
func (ep *EtcdProvider) getServiceKeyByID(instanceID string) string {
	// This is a simplified implementation
	// In practice, you might want to maintain an index or use a different approach
	ctx, cancel := context.WithTimeout(context.Background(), ep.config.Timeout)
	defer cancel()

	resp, err := ep.client.Get(ctx, ep.config.Namespace, clientv3.WithPrefix())
	if err != nil {
		return ""
	}

	for _, kv := range resp.Kvs {
		var instance types.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		if instance.ID == instanceID {
			return string(kv.Key)
		}
	}

	return ""
}

// matchesQuery checks if a service instance matches the query criteria
func (ep *EtcdProvider) matchesQuery(instance *types.ServiceInstance, query *types.ServiceQuery) bool {
	// Filter by service name
	if query.Name != "" && instance.Name != query.Name {
		return false
	}

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
