package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// ServiceDiscoveryManager manages service discovery
type ServiceDiscoveryManager struct {
	consulClient *api.Client
	services     map[string]*Service
	watchers     map[string]*ServiceWatcher
	mutex        sync.RWMutex
	config       *ServiceDiscoveryConfig
	logger       *logrus.Logger
}

// ServiceDiscoveryConfig holds service discovery configuration
type ServiceDiscoveryConfig struct {
	ConsulAddress string
	ConsulToken   string
	Datacenter    string
	Namespace     string
	ServicePrefix string
	CheckInterval time.Duration
	CheckTimeout  time.Duration
	TTL           time.Duration
}

// Service represents a service in the discovery system
type Service struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Tags      []string          `json:"tags"`
	Meta      map[string]string `json:"meta"`
	Health    ServiceHealth     `json:"health"`
	LastCheck time.Time         `json:"last_check"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Output    string                 `json:"output"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ServiceWatcher watches for service changes
type ServiceWatcher struct {
	ServiceName string
	Callback    func(services []*Service)
	StopChan    chan struct{}
	Running     bool
}

// ServiceRegistration represents a service registration
type ServiceRegistration struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
	Check   *ServiceCheck
}

// ServiceCheck represents a health check for a service
type ServiceCheck struct {
	HTTP                           string
	TCP                            string
	Interval                       string
	Timeout                        string
	DeregisterCriticalServiceAfter string
	TLSSkipVerify                  bool
	Header                         map[string][]string
}

// NewServiceDiscoveryManager creates a new service discovery manager
func NewServiceDiscoveryManager(config *ServiceDiscoveryConfig, logger *logrus.Logger) (*ServiceDiscoveryManager, error) {
	// Set default values
	if config.CheckInterval == 0 {
		config.CheckInterval = 10 * time.Second
	}
	if config.CheckTimeout == 0 {
		config.CheckTimeout = 3 * time.Second
	}
	if config.TTL == 0 {
		config.TTL = 30 * time.Second
	}

	// Create Consul client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.ConsulAddress
	consulConfig.Token = config.ConsulToken
	consulConfig.Datacenter = config.Datacenter

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	manager := &ServiceDiscoveryManager{
		consulClient: client,
		services:     make(map[string]*Service),
		watchers:     make(map[string]*ServiceWatcher),
		config:       config,
		logger:       logger,
	}

	logger.WithFields(logrus.Fields{
		"consul_address": config.ConsulAddress,
		"datacenter":     config.Datacenter,
		"namespace":      config.Namespace,
		"check_interval": config.CheckInterval,
		"check_timeout":  config.CheckTimeout,
		"ttl":            config.TTL,
	}).Info("Service discovery manager initialized successfully")

	return manager, nil
}

// RegisterService registers a service with the discovery system
func (sdm *ServiceDiscoveryManager) RegisterService(ctx context.Context, registration *ServiceRegistration) error {
	// Create service ID if not provided
	if registration.ID == "" {
		registration.ID = uuid.New().String()
	}

	// Create Consul service registration
	serviceRegistration := &api.AgentServiceRegistration{
		ID:      registration.ID,
		Name:    registration.Name,
		Address: registration.Address,
		Port:    registration.Port,
		Tags:    registration.Tags,
		Meta:    registration.Meta,
	}

	// Add health check if provided
	if registration.Check != nil {
		serviceRegistration.Check = &api.AgentServiceCheck{
			HTTP:                           registration.Check.HTTP,
			TCP:                            registration.Check.TCP,
			Interval:                       registration.Check.Interval,
			Timeout:                        registration.Check.Timeout,
			DeregisterCriticalServiceAfter: registration.Check.DeregisterCriticalServiceAfter,
			TLSSkipVerify:                  registration.Check.TLSSkipVerify,
			Header:                         registration.Check.Header,
		}
	}

	// Register service with Consul
	err := sdm.consulClient.Agent().ServiceRegister(serviceRegistration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// Create service object
	service := &Service{
		ID:        registration.ID,
		Name:      registration.Name,
		Address:   registration.Address,
		Port:      registration.Port,
		Tags:      registration.Tags,
		Meta:      registration.Meta,
		Health:    ServiceHealth{Status: "passing"},
		LastCheck: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store service locally
	sdm.mutex.Lock()
	sdm.services[registration.ID] = service
	sdm.mutex.Unlock()

	sdm.logger.WithFields(logrus.Fields{
		"service_id":   registration.ID,
		"service_name": registration.Name,
		"address":      registration.Address,
		"port":         registration.Port,
		"tags":         registration.Tags,
	}).Info("Service registered successfully")

	return nil
}

// DeregisterService deregisters a service from the discovery system
func (sdm *ServiceDiscoveryManager) DeregisterService(ctx context.Context, serviceID string) error {
	// Deregister service from Consul
	err := sdm.consulClient.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// Remove service from local storage
	sdm.mutex.Lock()
	delete(sdm.services, serviceID)
	sdm.mutex.Unlock()

	sdm.logger.WithField("service_id", serviceID).Info("Service deregistered successfully")
	return nil
}

// DiscoverServices discovers services by name
func (sdm *ServiceDiscoveryManager) DiscoverServices(ctx context.Context, serviceName string) ([]*Service, error) {
	// Query Consul for services
	services, _, err := sdm.consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var discoveredServices []*Service
	for _, service := range services {
		discoveredService := &Service{
			ID:        service.Service.ID,
			Name:      service.Service.Service,
			Address:   service.Service.Address,
			Port:      service.Service.Port,
			Tags:      service.Service.Tags,
			Meta:      service.Service.Meta,
			Health:    ServiceHealth{Status: "passing"},
			LastCheck: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Set health status based on checks
		if len(service.Checks) > 0 {
			discoveredService.Health.Status = service.Checks[0].Status
			discoveredService.Health.Message = service.Checks[0].Output
			discoveredService.Health.Timestamp = time.Unix(0, service.Checks[0].ModifyIndex)
		}

		discoveredServices = append(discoveredServices, discoveredService)
	}

	sdm.logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"count":        len(discoveredServices),
	}).Debug("Services discovered")

	return discoveredServices, nil
}

// GetService returns a service by ID
func (sdm *ServiceDiscoveryManager) GetService(serviceID string) (*Service, error) {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	service, exists := sdm.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	return service, nil
}

// GetAllServices returns all registered services
func (sdm *ServiceDiscoveryManager) GetAllServices() map[string]*Service {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	services := make(map[string]*Service)
	for id, service := range sdm.services {
		services[id] = service
	}

	return services
}

// WatchService watches for changes to a service
func (sdm *ServiceDiscoveryManager) WatchService(ctx context.Context, serviceName string, callback func(services []*Service)) error {
	// Create service watcher
	watcher := &ServiceWatcher{
		ServiceName: serviceName,
		Callback:    callback,
		StopChan:    make(chan struct{}),
		Running:     true,
	}

	// Store watcher
	sdm.mutex.Lock()
	sdm.watchers[serviceName] = watcher
	sdm.mutex.Unlock()

	// Start watching
	go sdm.watchServiceLoop(ctx, watcher)

	sdm.logger.WithField("service_name", serviceName).Info("Service watcher started")
	return nil
}

// watchServiceLoop runs the service watching loop
func (sdm *ServiceDiscoveryManager) watchServiceLoop(ctx context.Context, watcher *ServiceWatcher) {
	ticker := time.NewTicker(sdm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			watcher.Running = false
			close(watcher.StopChan)
			return
		case <-watcher.StopChan:
			watcher.Running = false
			return
		case <-ticker.C:
			services, err := sdm.DiscoverServices(ctx, watcher.ServiceName)
			if err != nil {
				sdm.logger.WithError(err).WithField("service_name", watcher.ServiceName).Error("Failed to discover services")
				continue
			}

			// Call callback with discovered services
			if watcher.Callback != nil {
				watcher.Callback(services)
			}
		}
	}
}

// StopWatchingService stops watching a service
func (sdm *ServiceDiscoveryManager) StopWatchingService(serviceName string) error {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	watcher, exists := sdm.watchers[serviceName]
	if !exists {
		return fmt.Errorf("service watcher not found: %s", serviceName)
	}

	// Stop watcher
	close(watcher.StopChan)
	delete(sdm.watchers, serviceName)

	sdm.logger.WithField("service_name", serviceName).Info("Service watcher stopped")
	return nil
}

// UpdateServiceHealth updates the health status of a service
func (sdm *ServiceDiscoveryManager) UpdateServiceHealth(ctx context.Context, serviceID string, health *ServiceHealth) error {
	// Update health in Consul
	err := sdm.consulClient.Agent().UpdateTTL("service:"+serviceID, health.Message, health.Status)
	if err != nil {
		return fmt.Errorf("failed to update service health: %w", err)
	}

	// Update local service
	sdm.mutex.Lock()
	if service, exists := sdm.services[serviceID]; exists {
		service.Health = *health
		service.LastCheck = time.Now()
		service.UpdatedAt = time.Now()
	}
	sdm.mutex.Unlock()

	sdm.logger.WithFields(logrus.Fields{
		"service_id": serviceID,
		"status":     health.Status,
		"message":    health.Message,
	}).Debug("Service health updated")

	return nil
}

// GetHealthyServices returns only healthy services
func (sdm *ServiceDiscoveryManager) GetHealthyServices(serviceName string) ([]*Service, error) {
	services, err := sdm.DiscoverServices(context.Background(), serviceName)
	if err != nil {
		return nil, err
	}

	var healthyServices []*Service
	for _, service := range services {
		if service.Health.Status == "passing" {
			healthyServices = append(healthyServices, service)
		}
	}

	return healthyServices, nil
}

// GetServiceEndpoints returns service endpoints in a load balancer friendly format
func (sdm *ServiceDiscoveryManager) GetServiceEndpoints(serviceName string) ([]string, error) {
	services, err := sdm.GetHealthyServices(serviceName)
	if err != nil {
		return nil, err
	}

	var endpoints []string
	for _, service := range services {
		endpoint := fmt.Sprintf("%s:%d", service.Address, service.Port)
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// Close closes the service discovery manager
func (sdm *ServiceDiscoveryManager) Close() error {
	// Stop all watchers
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	for serviceName, watcher := range sdm.watchers {
		close(watcher.StopChan)
		delete(sdm.watchers, serviceName)
	}

	sdm.logger.Info("Service discovery manager closed")
	return nil
}

// CreateServiceRegistration creates a service registration with default values
func CreateServiceRegistration(name, address string, port int) *ServiceRegistration {
	return &ServiceRegistration{
		ID:      uuid.New().String(),
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    []string{},
		Meta:    make(map[string]string),
		Check: &ServiceCheck{
			Interval: "10s",
			Timeout:  "3s",
		},
	}
}

// CreateHTTPHealthCheck creates an HTTP health check
func CreateHTTPHealthCheck(path string, interval, timeout time.Duration) *ServiceCheck {
	return &ServiceCheck{
		HTTP:                           path,
		Interval:                       interval.String(),
		Timeout:                        timeout.String(),
		DeregisterCriticalServiceAfter: "30s",
		TLSSkipVerify:                  false,
		Header:                         make(map[string][]string),
	}
}

// CreateTCPHealthCheck creates a TCP health check
func CreateTCPHealthCheck(interval, timeout time.Duration) *ServiceCheck {
	return &ServiceCheck{
		TCP:                            "",
		Interval:                       interval.String(),
		Timeout:                        timeout.String(),
		DeregisterCriticalServiceAfter: "30s",
	}
}
