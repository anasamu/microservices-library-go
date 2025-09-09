package integration

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/gateway"
	"github.com/anasamu/microservices-library-go/discovery/providers/static"
	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for the discovery library
// These tests require actual provider implementations and may need external services

func TestStaticProviderIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create static provider with predefined services
	staticConfig := &static.StaticConfig{
		Services: []*types.ServiceRegistration{
			{
				ID:       "web-service-1",
				Name:     "web-service",
				Address:  "192.168.1.100",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api", "v1"},
				Metadata: map[string]string{
					"version": "1.0.0",
					"env":     "production",
				},
				Health: types.HealthPassing,
			},
			{
				ID:       "web-service-2",
				Name:     "web-service",
				Address:  "192.168.1.101",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api", "v1"},
				Metadata: map[string]string{
					"version": "1.0.0",
					"env":     "production",
				},
				Health: types.HealthPassing,
			},
			{
				ID:       "db-service-1",
				Name:     "database",
				Address:  "192.168.1.200",
				Port:     5432,
				Protocol: "tcp",
				Tags:     []string{"database", "postgresql"},
				Metadata: map[string]string{
					"version": "13.0",
					"env":     "production",
				},
				Health: types.HealthPassing,
			},
		},
		Timeout: 30 * time.Second,
	}

	// Create static provider
	staticProvider, err := static.NewStaticProvider(staticConfig, logger)
	require.NoError(t, err)

	// Create discovery manager
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register static provider
	err = manager.RegisterProvider(staticProvider)
	require.NoError(t, err)

	// Connect to provider
	ctx := context.Background()
	err = staticProvider.Connect(ctx)
	require.NoError(t, err)

	// Test service discovery
	t.Run("DiscoverAllServices", func(t *testing.T) {
		query := &types.ServiceQuery{}
		services, err := manager.DiscoverServices(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, services, 2) // web-service and database

		// Verify web-service
		webService := findServiceByName(services, "web-service")
		require.NotNil(t, webService)
		assert.Len(t, webService.Instances, 2)
		assert.Equal(t, "web-service", webService.Name)

		// Verify database service
		dbService := findServiceByName(services, "database")
		require.NotNil(t, dbService)
		assert.Len(t, dbService.Instances, 1)
		assert.Equal(t, "database", dbService.Name)
	})

	// Test service discovery with filtering
	t.Run("DiscoverWithFilters", func(t *testing.T) {
		// Filter by service name
		query := &types.ServiceQuery{
			Name: "web-service",
		}
		services, err := manager.DiscoverServices(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "web-service", services[0].Name)
		assert.Len(t, services[0].Instances, 2)

		// Filter by tags
		query = &types.ServiceQuery{
			Tags: []string{"api"},
		}
		services, err = manager.DiscoverServices(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "web-service", services[0].Name)

		// Filter by health
		query = &types.ServiceQuery{
			Health: types.HealthPassing,
		}
		services, err = manager.DiscoverServices(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, services, 2)

		// Filter by metadata
		query = &types.ServiceQuery{
			Metadata: map[string]string{
				"env": "production",
			},
		}
		services, err = manager.DiscoverServices(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, services, 2)
	})

	// Test getting specific service
	t.Run("GetSpecificService", func(t *testing.T) {
		service, err := manager.GetService(ctx, "web-service")
		assert.NoError(t, err)
		assert.Equal(t, "web-service", service.Name)
		assert.Len(t, service.Instances, 2)

		// Test getting non-existent service
		_, err = manager.GetService(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test getting specific service instance
	t.Run("GetSpecificServiceInstance", func(t *testing.T) {
		instance, err := manager.GetServiceInstance(ctx, "web-service", "web-service-1")
		assert.NoError(t, err)
		assert.Equal(t, "web-service-1", instance.ID)
		assert.Equal(t, "web-service", instance.Name)
		assert.Equal(t, "192.168.1.100", instance.Address)
		assert.Equal(t, 8080, instance.Port)
		assert.Equal(t, types.HealthPassing, instance.Health)

		// Test getting non-existent instance
		_, err = manager.GetServiceInstance(ctx, "web-service", "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test health management
	t.Run("HealthManagement", func(t *testing.T) {
		// Get current health
		health, err := manager.GetHealth(ctx, "web-service-1")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthPassing, health)

		// Set health to critical
		err = manager.SetHealth(ctx, "web-service-1", types.HealthCritical)
		assert.NoError(t, err)

		// Verify health was updated
		health, err = manager.GetHealth(ctx, "web-service-1")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthCritical, health)

		// Set health back to passing
		err = manager.SetHealth(ctx, "web-service-1", types.HealthPassing)
		assert.NoError(t, err)

		// Verify health was updated
		health, err = manager.GetHealth(ctx, "web-service-1")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthPassing, health)
	})

	// Test service registration and deregistration
	t.Run("ServiceRegistration", func(t *testing.T) {
		// Register a new service
		registration := &types.ServiceRegistration{
			ID:       "cache-service-1",
			Name:     "cache",
			Address:  "192.168.1.300",
			Port:     6379,
			Protocol: "tcp",
			Tags:     []string{"cache", "redis"},
			Metadata: map[string]string{
				"version": "6.0",
				"env":     "production",
			},
			Health: types.HealthPassing,
		}

		err = manager.RegisterService(ctx, registration)
		assert.NoError(t, err)

		// Verify service was registered
		service, err := manager.GetService(ctx, "cache")
		assert.NoError(t, err)
		assert.Equal(t, "cache", service.Name)
		assert.Len(t, service.Instances, 1)

		// Deregister the service
		err = manager.DeregisterService(ctx, "cache-service-1")
		assert.NoError(t, err)

		// Verify service was deregistered
		_, err = manager.GetService(ctx, "cache")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test statistics
	t.Run("Statistics", func(t *testing.T) {
		stats, err := manager.GetStats(ctx)
		assert.NoError(t, err)
		assert.Len(t, stats, 1)
		assert.Contains(t, stats, "static")

		providerStats := stats["static"]
		assert.Equal(t, "static", providerStats.Provider)
		assert.Equal(t, int64(2), providerStats.ServicesRegistered) // web-service and database
		assert.Equal(t, int64(3), providerStats.InstancesActive)    // 2 web + 1 db
	})

	// Test listing services
	t.Run("ListServices", func(t *testing.T) {
		allServices, err := manager.ListServices(ctx)
		assert.NoError(t, err)
		assert.Len(t, allServices, 1)
		assert.Contains(t, allServices, "static")

		services := allServices["static"]
		assert.Len(t, services, 2)
		assert.Contains(t, services, "web-service")
		assert.Contains(t, services, "database")
	})

	// Test provider information
	t.Run("ProviderInfo", func(t *testing.T) {
		providerInfo := manager.GetProviderInfo()
		assert.Len(t, providerInfo, 1)
		assert.Contains(t, providerInfo, "static")

		info := providerInfo["static"]
		assert.Equal(t, "static", info.Name)
		assert.True(t, info.IsConnected)
		assert.NotNil(t, info.ConnectionInfo)
		assert.NotEmpty(t, info.SupportedFeatures)
	})

	// Clean up
	err = manager.Close()
	assert.NoError(t, err)
}

func TestMultipleProvidersIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create discovery manager
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Create multiple static providers with different services
	staticConfig1 := &static.StaticConfig{
		Services: []*types.ServiceRegistration{
			{
				ID:       "web-service-1",
				Name:     "web-service",
				Address:  "192.168.1.100",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api"},
				Health:   types.HealthPassing,
			},
		},
		Timeout: 30 * time.Second,
	}

	staticConfig2 := &static.StaticConfig{
		Services: []*types.ServiceRegistration{
			{
				ID:       "db-service-1",
				Name:     "database",
				Address:  "192.168.1.200",
				Port:     5432,
				Protocol: "tcp",
				Tags:     []string{"database", "postgresql"},
				Health:   types.HealthPassing,
			},
		},
		Timeout: 30 * time.Second,
	}

	// Create providers
	staticProvider1, err := static.NewStaticProvider(staticConfig1, logger)
	require.NoError(t, err)

	staticProvider2, err := static.NewStaticProvider(staticConfig2, logger)
	require.NoError(t, err)

	// Register providers
	err = manager.RegisterProvider(staticProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider(staticProvider2)
	require.NoError(t, err)

	// Connect providers
	ctx := context.Background()
	err = staticProvider1.Connect(ctx)
	require.NoError(t, err)

	err = staticProvider2.Connect(ctx)
	require.NoError(t, err)

	// Test discovering services from specific provider
	t.Run("DiscoverFromSpecificProvider", func(t *testing.T) {
		// Discover from provider 1
		query := &types.ServiceQuery{}
		services, err := manager.DiscoverServicesWithProvider(ctx, "static", query)
		assert.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "web-service", services[0].Name)

		// Discover from provider 2
		services, err = manager.DiscoverServicesWithProvider(ctx, "static", query)
		assert.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "database", services[0].Name)
	})

	// Test provider-specific operations
	t.Run("ProviderSpecificOperations", func(t *testing.T) {
		// Register service with specific provider
		registration := &types.ServiceRegistration{
			ID:       "cache-service-1",
			Name:     "cache",
			Address:  "192.168.1.300",
			Port:     6379,
			Protocol: "tcp",
			Tags:     []string{"cache", "redis"},
			Health:   types.HealthPassing,
		}

		err = manager.RegisterServiceWithProvider(ctx, "static", registration)
		assert.NoError(t, err)

		// Verify service was registered
		service, err := manager.GetServiceWithProvider(ctx, "static", "cache")
		assert.NoError(t, err)
		assert.Equal(t, "cache", service.Name)

		// Set health with specific provider
		err = manager.SetHealthWithProvider(ctx, "static", "cache-service-1", types.HealthCritical)
		assert.NoError(t, err)

		// Verify health was set
		health, err := manager.GetHealthWithProvider(ctx, "static", "cache-service-1")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthCritical, health)
	})

	// Clean up
	err = manager.Close()
	assert.NoError(t, err)
}

func TestLoadBalancingIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create static provider with multiple instances
	staticConfig := &static.StaticConfig{
		Services: []*types.ServiceRegistration{
			{
				ID:       "web-service-1",
				Name:     "web-service",
				Address:  "192.168.1.100",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api"},
				Health:   types.HealthPassing,
				Weight:   1,
			},
			{
				ID:       "web-service-2",
				Name:     "web-service",
				Address:  "192.168.1.101",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api"},
				Health:   types.HealthPassing,
				Weight:   2,
			},
			{
				ID:       "web-service-3",
				Name:     "web-service",
				Address:  "192.168.1.102",
				Port:     8080,
				Protocol: "http",
				Tags:     []string{"web", "api"},
				Health:   types.HealthPassing,
				Weight:   3,
			},
		},
		Timeout: 30 * time.Second,
	}

	// Create static provider
	staticProvider, err := static.NewStaticProvider(staticConfig, logger)
	require.NoError(t, err)

	// Create discovery manager
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register static provider
	err = manager.RegisterProvider(staticProvider)
	require.NoError(t, err)

	// Connect to provider
	ctx := context.Background()
	err = staticProvider.Connect(ctx)
	require.NoError(t, err)

	// Test load balancing
	t.Run("LoadBalancing", func(t *testing.T) {
		// Get service with multiple instances
		service, err := manager.GetService(ctx, "web-service")
		assert.NoError(t, err)
		assert.Len(t, service.Instances, 3)

		// Test different load balancing strategies
		instances := service.Instances

		// Round Robin
		lb := &types.LoadBalancerRoundRobin{}
		selected, err := lb.SelectInstance(instances)
		assert.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Contains(t, []string{"web-service-1", "web-service-2", "web-service-3"}, selected.ID)

		// Random
		lb = &types.LoadBalancerRandom{}
		selected, err = lb.SelectInstance(instances)
		assert.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Contains(t, []string{"web-service-1", "web-service-2", "web-service-3"}, selected.ID)

		// Weighted
		lb = &types.LoadBalancerWeighted{}
		selected, err = lb.SelectInstance(instances)
		assert.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Contains(t, []string{"web-service-1", "web-service-2", "web-service-3"}, selected.ID)
	})

	// Clean up
	err = manager.Close()
	assert.NoError(t, err)
}

// Helper function to find service by name
func findServiceByName(services []*types.Service, name string) *types.Service {
	for _, service := range services {
		if service.Name == name {
			return service
		}
	}
	return nil
}
