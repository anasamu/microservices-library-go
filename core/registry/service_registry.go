package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/core/interfaces"
)

// ServiceRegistry manages service instances and their lifecycle
type ServiceRegistry struct {
	services     map[string]interfaces.Service
	dependencies map[string][]string
	healthChecks map[string]interfaces.HealthChecker
	mu           sync.RWMutex
	started      bool
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]interfaces.Service),
		dependencies: make(map[string][]string),
		healthChecks: make(map[string]interfaces.HealthChecker),
	}
}

// RegisterService registers a service with the registry
func (r *ServiceRegistry) RegisterService(name string, service interfaces.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return fmt.Errorf("cannot register service after registry has been started")
	}

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s is already registered", name)
	}

	r.services[name] = service
	return nil
}

// RegisterServiceWithDependencies registers a service with its dependencies
func (r *ServiceRegistry) RegisterServiceWithDependencies(name string, service interfaces.Service, dependencies []string) error {
	if err := r.RegisterService(name, service); err != nil {
		return err
	}

	r.mu.Lock()
	r.dependencies[name] = dependencies
	r.mu.Unlock()

	return nil
}

// RegisterHealthCheck registers a health check for a service
func (r *ServiceRegistry) RegisterHealthCheck(serviceName string, healthChecker interfaces.HealthChecker) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[serviceName]; !exists {
		return fmt.Errorf("service %s is not registered", serviceName)
	}

	r.healthChecks[serviceName] = healthChecker
	return nil
}

// GetService retrieves a service by name
func (r *ServiceRegistry) GetService(name string) (interfaces.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// GetServices returns all registered services
func (r *ServiceRegistry) GetServices() map[string]interfaces.Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make(map[string]interfaces.Service)
	for name, service := range r.services {
		services[name] = service
	}

	return services
}

// ListServices returns a list of all service names
func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}

	return names
}

// StartAll starts all registered services in dependency order
func (r *ServiceRegistry) StartAll(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return fmt.Errorf("registry has already been started")
	}
	r.started = true
	r.mu.Unlock()

	// Resolve dependency order
	startOrder, err := r.resolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Start services in dependency order
	for _, serviceName := range startOrder {
		service := r.services[serviceName]

		// Set dependencies
		if deps, exists := r.dependencies[serviceName]; exists {
			for _, depName := range deps {
				depService, exists := r.services[depName]
				if !exists {
					return fmt.Errorf("dependency %s not found for service %s", depName, serviceName)
				}
				if err := service.SetDependency(depName, depService); err != nil {
					return fmt.Errorf("failed to set dependency %s for service %s: %w", depName, serviceName, err)
				}
			}
		}

		// Start service
		if err := service.Start(ctx); err != nil {
			// Stop already started services
			r.stopServicesInReverseOrder(ctx, startOrder[:len(startOrder)-1])
			return fmt.Errorf("failed to start service %s: %w", serviceName, err)
		}
	}

	return nil
}

// StopAll stops all registered services in reverse dependency order
func (r *ServiceRegistry) StopAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Resolve dependency order
	startOrder, err := r.resolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Stop services in reverse order
	return r.stopServicesInReverseOrder(ctx, startOrder)
}

// HealthCheck performs health check on all services
func (r *ServiceRegistry) HealthCheck(ctx context.Context) map[string]interfaces.HealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]interfaces.HealthStatus)

	for serviceName, service := range r.services {
		// Check if service has dedicated health checker
		if healthChecker, exists := r.healthChecks[serviceName]; exists {
			results[serviceName] = healthChecker.CheckHealth(ctx)
		} else {
			// Use service's built-in health check
			if err := service.HealthCheck(ctx); err != nil {
				results[serviceName] = interfaces.HealthStatusUnhealthy
			} else {
				results[serviceName] = interfaces.HealthStatusHealthy
			}
		}
	}

	return results
}

// StartService starts a specific service
func (r *ServiceRegistry) StartService(ctx context.Context, name string) error {
	r.mu.RLock()
	service, exists := r.services[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	// Set dependencies
	if deps, exists := r.dependencies[name]; exists {
		for _, depName := range deps {
			depService, exists := r.services[depName]
			if !exists {
				return fmt.Errorf("dependency %s not found for service %s", depName, name)
			}
			if err := service.SetDependency(depName, depService); err != nil {
				return fmt.Errorf("failed to set dependency %s for service %s: %w", depName, name, err)
			}
		}
	}

	return service.Start(ctx)
}

// StopService stops a specific service
func (r *ServiceRegistry) StopService(ctx context.Context, name string) error {
	r.mu.RLock()
	service, exists := r.services[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	return service.Stop(ctx)
}

// IsStarted returns whether the registry has been started
func (r *ServiceRegistry) IsStarted() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.started
}

// resolveDependencies resolves the dependency order using topological sort
func (r *ServiceRegistry) resolveDependencies() ([]string, error) {
	// Create adjacency list
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize graph and in-degree
	for serviceName := range r.services {
		graph[serviceName] = []string{}
		inDegree[serviceName] = 0
	}

	// Build graph from dependencies
	for serviceName, deps := range r.dependencies {
		for _, dep := range deps {
			if _, exists := r.services[dep]; !exists {
				return nil, fmt.Errorf("dependency %s not found for service %s", dep, serviceName)
			}
			graph[dep] = append(graph[dep], serviceName)
			inDegree[serviceName]++
		}
	}

	// Topological sort using Kahn's algorithm
	var queue []string
	var result []string

	// Add services with no dependencies to queue
	for serviceName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, serviceName)
		}
	}

	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Process dependencies
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(r.services) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// stopServicesInReverseOrder stops services in reverse order
func (r *ServiceRegistry) stopServicesInReverseOrder(ctx context.Context, startOrder []string) error {
	// Stop in reverse order
	for i := len(startOrder) - 1; i >= 0; i-- {
		serviceName := startOrder[i]
		service := r.services[serviceName]

		if err := service.Stop(ctx); err != nil {
			// Log error but continue stopping other services
			fmt.Printf("Error stopping service %s: %v\n", serviceName, err)
		}
	}

	return nil
}

// ServiceManager provides high-level service management
type ServiceManager struct {
	registry *ServiceRegistry
	config   *ServiceManagerConfig
}

// ServiceManagerConfig represents service manager configuration
type ServiceManagerConfig struct {
	HealthCheckInterval time.Duration
	StartTimeout        time.Duration
	StopTimeout         time.Duration
	AutoRestart         bool
	MaxRestartAttempts  int
}

// NewServiceManager creates a new service manager
func NewServiceManager(config *ServiceManagerConfig) *ServiceManager {
	if config == nil {
		config = &ServiceManagerConfig{
			HealthCheckInterval: 30 * time.Second,
			StartTimeout:        30 * time.Second,
			StopTimeout:         30 * time.Second,
			AutoRestart:         false,
			MaxRestartAttempts:  3,
		}
	}

	return &ServiceManager{
		registry: NewServiceRegistry(),
		config:   config,
	}
}

// GetRegistry returns the underlying service registry
func (sm *ServiceManager) GetRegistry() *ServiceRegistry {
	return sm.registry
}

// StartWithTimeout starts all services with timeout
func (sm *ServiceManager) StartWithTimeout(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, sm.config.StartTimeout)
	defer cancel()

	return sm.registry.StartAll(timeoutCtx)
}

// StopWithTimeout stops all services with timeout
func (sm *ServiceManager) StopWithTimeout(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, sm.config.StopTimeout)
	defer cancel()

	return sm.registry.StopAll(timeoutCtx)
}

// StartHealthMonitoring starts health monitoring for all services
func (sm *ServiceManager) StartHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(sm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			healthStatus := sm.registry.HealthCheck(ctx)
			sm.handleHealthStatus(healthStatus)
		}
	}
}

// handleHealthStatus handles health status results
func (sm *ServiceManager) handleHealthStatus(status map[string]interfaces.HealthStatus) {
	for serviceName, health := range status {
		if health == interfaces.HealthStatusUnhealthy {
			fmt.Printf("Service %s is unhealthy\n", serviceName)
			// Implement restart logic if AutoRestart is enabled
			if sm.config.AutoRestart {
				// TODO: Implement restart logic
			}
		}
	}
}
