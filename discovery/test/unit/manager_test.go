package unit

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/gateway"
	"github.com/anasamu/microservices-library-go/discovery/test/mocks"
	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscoveryManager(t *testing.T) {
	logger := logrus.New()

	// Test with nil config
	manager := gateway.NewDiscoveryManager(nil, logger)
	assert.NotNil(t, manager)

	// Test with custom config
	config := &gateway.ManagerConfig{
		DefaultProvider: "test",
		RetryAttempts:   5,
		RetryDelay:      2 * time.Second,
		Timeout:         60 * time.Second,
		FallbackEnabled: false,
	}

	manager = gateway.NewDiscoveryManager(config, logger)
	assert.NotNil(t, manager)
}

func TestRegisterProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Test registering a valid provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	// Test registering the same provider again
	err = manager.RegisterProvider(mockProvider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test registering nil provider
	err = manager.RegisterProvider(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test registering provider with empty name
	emptyNameProvider := mocks.NewMockDiscoveryProvider("")
	err = manager.RegisterProvider(emptyNameProvider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestGetProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Test getting non-existent provider
	_, err := manager.GetProvider("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Register a provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err = manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Test getting existing provider
	provider, err := manager.GetProvider("test-provider")
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)
}

func TestGetDefaultProvider(t *testing.T) {
	logger := logrus.New()

	// Test with no providers
	manager := gateway.NewDiscoveryManager(nil, logger)
	_, err := manager.GetDefaultProvider()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no providers registered")

	// Test with providers but no default specified
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err = manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	provider, err := manager.GetDefaultProvider()
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)

	// Test with default provider specified
	config := &gateway.ManagerConfig{
		DefaultProvider: "test-provider",
	}
	manager = gateway.NewDiscoveryManager(config, logger)
	err = manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	provider, err = manager.GetDefaultProvider()
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)
}

func TestRegisterService(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Test registering a service
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Tags:     []string{"test", "api"},
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	assert.NoError(t, err)

	// Verify service was registered
	services := mockProvider.GetServices()
	assert.Contains(t, services, "test-service")
	assert.Len(t, services["test-service"].Instances, 1)
	assert.Equal(t, "test-service-1", services["test-service"].Instances[0].ID)
}

func TestDeregisterService(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register a service first
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	require.NoError(t, err)

	// Deregister the service
	err = manager.DeregisterService(ctx, "test-service-1")
	assert.NoError(t, err)

	// Verify service was deregistered
	services := mockProvider.GetServices()
	assert.NotContains(t, services, "test-service")
}

func TestDiscoverServices(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register some services
	registrations := []*types.ServiceRegistration{
		{
			ID:       "test-service-1",
			Name:     "test-service",
			Address:  "192.168.1.100",
			Port:     8080,
			Protocol: "http",
			Tags:     []string{"test", "api"},
			Health:   types.HealthPassing,
		},
		{
			ID:       "test-service-2",
			Name:     "test-service",
			Address:  "192.168.1.101",
			Port:     8080,
			Protocol: "http",
			Tags:     []string{"test", "api"},
			Health:   types.HealthPassing,
		},
		{
			ID:       "other-service-1",
			Name:     "other-service",
			Address:  "192.168.1.102",
			Port:     9090,
			Protocol: "grpc",
			Tags:     []string{"other", "grpc"},
			Health:   types.HealthPassing,
		},
	}

	for _, reg := range registrations {
		err = manager.RegisterService(ctx, reg)
		require.NoError(t, err)
	}

	// Test discovering all services
	query := &types.ServiceQuery{}
	services, err := manager.DiscoverServices(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, services, 2) // test-service and other-service

	// Test discovering specific service
	query = &types.ServiceQuery{
		Name: "test-service",
	}
	services, err = manager.DiscoverServices(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "test-service", services[0].Name)
	assert.Len(t, services[0].Instances, 2)

	// Test discovering with tags filter
	query = &types.ServiceQuery{
		Tags: []string{"api"},
	}
	services, err = manager.DiscoverServices(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "test-service", services[0].Name)

	// Test discovering with health filter
	query = &types.ServiceQuery{
		Health: types.HealthPassing,
	}
	services, err = manager.DiscoverServices(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestGetService(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	require.NoError(t, err)

	// Test getting existing service
	service, err := manager.GetService(ctx, "test-service")
	assert.NoError(t, err)
	assert.Equal(t, "test-service", service.Name)
	assert.Len(t, service.Instances, 1)

	// Test getting non-existent service
	_, err = manager.GetService(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetServiceInstance(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	require.NoError(t, err)

	// Test getting existing service instance
	instance, err := manager.GetServiceInstance(ctx, "test-service", "test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, "test-service-1", instance.ID)
	assert.Equal(t, "test-service", instance.Name)

	// Test getting non-existent service instance
	_, err = manager.GetServiceInstance(ctx, "test-service", "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetHealth(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	require.NoError(t, err)

	// Test setting health
	err = manager.SetHealth(ctx, "test-service-1", types.HealthCritical)
	assert.NoError(t, err)

	// Verify health was set
	health, err := manager.GetHealth(ctx, "test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, types.HealthCritical, health)
}

func TestGetHealth(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	err = manager.RegisterService(ctx, registration)
	require.NoError(t, err)

	// Test getting health
	health, err := manager.GetHealth(ctx, "test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, types.HealthPassing, health)

	// Test getting health for non-existent service
	_, err = manager.GetHealth(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListProviders(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Test with no providers
	providers := manager.ListProviders()
	assert.Len(t, providers, 0)

	// Register some providers
	mockProvider1 := mocks.NewMockDiscoveryProvider("provider-1")
	mockProvider2 := mocks.NewMockDiscoveryProvider("provider-2")

	err := manager.RegisterProvider(mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider(mockProvider2)
	require.NoError(t, err)

	// Test listing providers
	providers = manager.ListProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "provider-1")
	assert.Contains(t, providers, "provider-2")
}

func TestGetProviderInfo(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Test with no providers
	info := manager.GetProviderInfo()
	assert.Len(t, info, 0)

	// Register a provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Test getting provider info
	info = manager.GetProviderInfo()
	assert.Len(t, info, 1)
	assert.Contains(t, info, "test-provider")

	providerInfo := info["test-provider"]
	assert.Equal(t, "test-provider", providerInfo.Name)
	assert.True(t, providerInfo.IsConnected)
	assert.NotNil(t, providerInfo.ConnectionInfo)
	assert.NotEmpty(t, providerInfo.SupportedFeatures)
}

func TestGetStats(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register some services
	registrations := []*types.ServiceRegistration{
		{
			ID:       "test-service-1",
			Name:     "test-service",
			Address:  "192.168.1.100",
			Port:     8080,
			Protocol: "http",
			Health:   types.HealthPassing,
		},
		{
			ID:       "test-service-2",
			Name:     "test-service",
			Address:  "192.168.1.101",
			Port:     8080,
			Protocol: "http",
			Health:   types.HealthPassing,
		},
	}

	for _, reg := range registrations {
		err = manager.RegisterService(ctx, reg)
		require.NoError(t, err)
	}

	// Test getting stats
	stats, err := manager.GetStats(ctx)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Contains(t, stats, "test-provider")

	providerStats := stats["test-provider"]
	assert.Equal(t, "test-provider", providerStats.Provider)
	assert.Equal(t, int64(1), providerStats.ServicesRegistered) // 1 service
	assert.Equal(t, int64(2), providerStats.InstancesActive)    // 2 instances
}

func TestListServices(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Register some services
	registrations := []*types.ServiceRegistration{
		{
			ID:       "test-service-1",
			Name:     "test-service",
			Address:  "192.168.1.100",
			Port:     8080,
			Protocol: "http",
			Health:   types.HealthPassing,
		},
		{
			ID:       "other-service-1",
			Name:     "other-service",
			Address:  "192.168.1.101",
			Port:     9090,
			Protocol: "grpc",
			Health:   types.HealthPassing,
		},
	}

	for _, reg := range registrations {
		err = manager.RegisterService(ctx, reg)
		require.NoError(t, err)
	}

	// Test listing services
	allServices, err := manager.ListServices(ctx)
	assert.NoError(t, err)
	assert.Len(t, allServices, 1)
	assert.Contains(t, allServices, "test-provider")

	services := allServices["test-provider"]
	assert.Len(t, services, 2)
	assert.Contains(t, services, "test-service")
	assert.Contains(t, services, "other-service")
}

func TestClose(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewDiscoveryManager(nil, logger)

	// Register a mock provider
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Test closing
	err = manager.Close()
	assert.NoError(t, err)

	// Verify provider is disconnected
	assert.False(t, mockProvider.IsConnected())
}

func TestRetryLogic(t *testing.T) {
	logger := logrus.New()
	config := &gateway.ManagerConfig{
		RetryAttempts: 2,
		RetryDelay:    100 * time.Millisecond,
		Timeout:       1 * time.Second,
	}
	manager := gateway.NewDiscoveryManager(config, logger)

	// Register a mock provider that fails on first attempt
	mockProvider := mocks.NewMockDiscoveryProvider("test-provider")
	mockProvider.SetErrorOnRegister(&types.DiscoveryError{
		Code:    types.ErrCodeConnection,
		Message: "temporary error",
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Connect the provider
	ctx := context.Background()
	err = mockProvider.Connect(ctx)
	require.NoError(t, err)

	// Test that retry logic works
	registration := &types.ServiceRegistration{
		ID:       "test-service-1",
		Name:     "test-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthPassing,
	}

	// This should fail due to the error we set
	err = manager.RegisterService(ctx, registration)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temporary error")

	// Clear the error and try again
	mockProvider.ClearErrors()
	err = manager.RegisterService(ctx, registration)
	assert.NoError(t, err)
}
