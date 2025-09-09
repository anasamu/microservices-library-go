package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/failover/gateway"
	"github.com/anasamu/microservices-library-go/failover/test/mocks"
	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFailoverManager(t *testing.T) {
	logger := logrus.New()

	t.Run("with config", func(t *testing.T) {
		config := &gateway.ManagerConfig{
			DefaultProvider: "test",
			RetryAttempts:   5,
			RetryDelay:      2 * time.Second,
			Timeout:         60 * time.Second,
		}

		manager := gateway.NewFailoverManager(config, logger)
		assert.NotNil(t, manager)
		assert.Equal(t, "test", manager.GetDefaultProvider())
	})

	t.Run("without config", func(t *testing.T) {
		manager := gateway.NewFailoverManager(nil, logger)
		assert.NotNil(t, manager)
	})
}

func TestRegisterProvider(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	t.Run("successful registration", func(t *testing.T) {
		err := manager.RegisterProvider(mockProvider)
		assert.NoError(t, err)

		providers := manager.ListProviders()
		assert.Contains(t, providers, "test-provider")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		err := manager.RegisterProvider(mockProvider)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("nil provider", func(t *testing.T) {
		err := manager.RegisterProvider(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty provider name", func(t *testing.T) {
		emptyProvider := mocks.NewMockProvider("")
		err := manager.RegisterProvider(emptyProvider)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestGetProvider(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("existing provider", func(t *testing.T) {
		provider, err := manager.GetProvider("test-provider")
		assert.NoError(t, err)
		assert.Equal(t, "test-provider", provider.GetName())
	})

	t.Run("non-existing provider", func(t *testing.T) {
		provider, err := manager.GetProvider("non-existing")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetDefaultProvider(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())

	t.Run("no providers", func(t *testing.T) {
		provider, err := manager.GetDefaultProvider()
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "no providers registered")
	})

	t.Run("with default provider", func(t *testing.T) {
		config := &gateway.ManagerConfig{
			DefaultProvider: "test-provider",
		}
		manager := gateway.NewFailoverManager(config, logrus.New())
		mockProvider := mocks.NewMockProvider("test-provider")

		err := manager.RegisterProvider(mockProvider)
		require.NoError(t, err)

		provider, err := manager.GetDefaultProvider()
		assert.NoError(t, err)
		assert.Equal(t, "test-provider", provider.GetName())
	})

	t.Run("first available provider", func(t *testing.T) {
		manager := gateway.NewFailoverManager(nil, logrus.New())
		mockProvider := mocks.NewMockProvider("first-provider")

		err := manager.RegisterProvider(mockProvider)
		require.NoError(t, err)

		provider, err := manager.GetDefaultProvider()
		assert.NoError(t, err)
		assert.Equal(t, "first-provider", provider.GetName())
	})
}

func TestRegisterEndpoint(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	endpoint := &types.ServiceEndpoint{
		ID:       "test-endpoint",
		Name:     "test-service",
		Address:  "127.0.0.1",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthHealthy,
	}

	t.Run("successful registration", func(t *testing.T) {
		err := manager.RegisterEndpoint(context.Background(), endpoint)
		assert.NoError(t, err)

		// Verify endpoint was registered
		registeredEndpoint, err := manager.GetEndpoint(context.Background(), "test-endpoint")
		assert.NoError(t, err)
		assert.Equal(t, endpoint.ID, registeredEndpoint.ID)
	})

	t.Run("with specific provider", func(t *testing.T) {
		err := manager.RegisterEndpointWithProvider(context.Background(), "test-provider", endpoint)
		assert.NoError(t, err)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		err := manager.RegisterEndpointWithProvider(context.Background(), "non-existing", endpoint)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDeregisterEndpoint(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	endpoint := &types.ServiceEndpoint{
		ID:       "test-endpoint",
		Name:     "test-service",
		Address:  "127.0.0.1",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthHealthy,
	}

	// Register endpoint first
	err = manager.RegisterEndpoint(context.Background(), endpoint)
	require.NoError(t, err)

	t.Run("successful deregistration", func(t *testing.T) {
		err := manager.DeregisterEndpoint(context.Background(), "test-endpoint")
		assert.NoError(t, err)

		// Verify endpoint was deregistered
		_, err = manager.GetEndpoint(context.Background(), "test-endpoint")
		assert.Error(t, err)
	})
}

func TestSelectEndpoint(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	// Add a healthy endpoint to the mock provider
	healthyEndpoint := &types.ServiceEndpoint{
		ID:       "healthy-endpoint",
		Name:     "test-service",
		Address:  "127.0.0.1",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthHealthy,
	}
	mockProvider.AddEndpoint(healthyEndpoint)

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	config := &types.FailoverConfig{
		Name:     "test-config",
		Strategy: types.StrategyRoundRobin,
	}

	t.Run("successful selection", func(t *testing.T) {
		result, err := manager.SelectEndpoint(context.Background(), config)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "healthy-endpoint", result.Endpoint.ID)
		assert.Equal(t, types.StrategyRoundRobin, result.Strategy)
	})

	t.Run("no healthy endpoints", func(t *testing.T) {
		// Remove the healthy endpoint
		mockProvider.RemoveEndpoint("healthy-endpoint")

		result, err := manager.SelectEndpoint(context.Background(), config)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, result.Error.Error(), "no healthy endpoints")
	})
}

func TestExecuteWithFailover(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	// Add a healthy endpoint to the mock provider
	healthyEndpoint := &types.ServiceEndpoint{
		ID:       "healthy-endpoint",
		Name:     "test-service",
		Address:  "127.0.0.1",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthHealthy,
	}
	mockProvider.AddEndpoint(healthyEndpoint)

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	config := &types.FailoverConfig{
		Name:          "test-config",
		Strategy:      types.StrategyRoundRobin,
		RetryAttempts: 3,
		RetryDelay:    100 * time.Millisecond,
	}

	t.Run("successful execution", func(t *testing.T) {
		result, err := manager.ExecuteWithFailover(context.Background(), config, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "healthy-endpoint", result.Endpoint.ID)
		assert.Equal(t, 1, result.Attempts)
		assert.Equal(t, "success", result.Metadata["result"])
	})

	t.Run("failed execution", func(t *testing.T) {
		result, err := manager.ExecuteWithFailover(context.Background(), config, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
			return nil, errors.New("execution failed")
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, result.Error.Error(), "execution failed")
	})
}

func TestHealthCheck(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	// Add an endpoint to the mock provider
	endpoint := &types.ServiceEndpoint{
		ID:       "test-endpoint",
		Name:     "test-service",
		Address:  "127.0.0.1",
		Port:     8080,
		Protocol: "http",
		Health:   types.HealthHealthy,
	}
	mockProvider.AddEndpoint(endpoint)

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("successful health check", func(t *testing.T) {
		status, err := manager.HealthCheck(context.Background(), "test-endpoint")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthHealthy, status)
	})

	t.Run("non-existing endpoint", func(t *testing.T) {
		status, err := manager.HealthCheck(context.Background(), "non-existing")
		assert.Error(t, err)
		assert.Equal(t, types.HealthUnknown, status)
	})
}

func TestHealthCheckAll(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	// Add multiple endpoints
	endpoints := []*types.ServiceEndpoint{
		{
			ID:     "endpoint-1",
			Name:   "test-service",
			Health: types.HealthHealthy,
		},
		{
			ID:     "endpoint-2",
			Name:   "test-service",
			Health: types.HealthUnhealthy,
		},
	}

	for _, endpoint := range endpoints {
		mockProvider.AddEndpoint(endpoint)
	}

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("health check all endpoints", func(t *testing.T) {
		statuses, err := manager.HealthCheckAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, statuses, 2)
		assert.Equal(t, types.HealthHealthy, statuses["endpoint-1"])
		assert.Equal(t, types.HealthUnhealthy, statuses["endpoint-2"])
	})
}

func TestConfigure(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	config := &types.FailoverConfig{
		Name:     "test-config",
		Strategy: types.StrategyWeighted,
	}

	t.Run("successful configuration", func(t *testing.T) {
		err := manager.Configure(context.Background(), config)
		assert.NoError(t, err)
	})
}

func TestGetStats(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("get statistics", func(t *testing.T) {
		stats, err := manager.GetStats(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "test-provider")
	})
}

func TestGetEvents(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("get events", func(t *testing.T) {
		events, err := manager.GetEvents(context.Background(), 10)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Contains(t, events, "test-provider")
	})
}

func TestListProviders(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())

	t.Run("no providers", func(t *testing.T) {
		providers := manager.ListProviders()
		assert.Empty(t, providers)
	})

	t.Run("multiple providers", func(t *testing.T) {
		provider1 := mocks.NewMockProvider("provider-1")
		provider2 := mocks.NewMockProvider("provider-2")

		err := manager.RegisterProvider(provider1)
		require.NoError(t, err)

		err = manager.RegisterProvider(provider2)
		require.NoError(t, err)

		providers := manager.ListProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "provider-1")
		assert.Contains(t, providers, "provider-2")
	})
}

func TestGetProviderInfo(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("get provider info", func(t *testing.T) {
		info := manager.GetProviderInfo()
		assert.NotNil(t, info)
		assert.Contains(t, info, "test-provider")

		providerInfo := info["test-provider"]
		assert.Equal(t, "test-provider", providerInfo.Name)
		assert.False(t, providerInfo.IsConnected)
		assert.NotEmpty(t, providerInfo.SupportedFeatures)
	})
}

func TestClose(t *testing.T) {
	manager := gateway.NewFailoverManager(nil, logrus.New())
	mockProvider := mocks.NewMockProvider("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	t.Run("close manager", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)
	})
}
