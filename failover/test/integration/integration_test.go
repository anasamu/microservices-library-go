package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/failover/gateway"
	"github.com/anasamu/microservices-library-go/failover/test/mocks"
	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailoverIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create manager configuration
	config := &gateway.ManagerConfig{
		DefaultProvider: "consul",
		RetryAttempts:   3,
		RetryDelay:      100 * time.Millisecond,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	// Create failover manager
	manager := gateway.NewFailoverManager(config, logger)

	// Create mock providers
	consulProvider := mocks.NewMockProvider("consul")
	kubernetesProvider := mocks.NewMockProvider("kubernetes")

	// Register providers
	err := manager.RegisterProvider(consulProvider)
	require.NoError(t, err)

	err = manager.RegisterProvider(kubernetesProvider)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("endpoint registration and discovery", func(t *testing.T) {
		// Create test endpoints
		endpoints := []*types.ServiceEndpoint{
			{
				ID:       "web-1",
				Name:     "web-service",
				Address:  "192.168.1.10",
				Port:     8080,
				Protocol: "http",
				Health:   types.HealthHealthy,
				Weight:   100,
				Priority: 1,
				Tags:     []string{"web", "frontend"},
				Metadata: map[string]string{
					"region": "us-east-1",
					"zone":   "us-east-1a",
				},
			},
			{
				ID:       "web-2",
				Name:     "web-service",
				Address:  "192.168.1.11",
				Port:     8080,
				Protocol: "http",
				Health:   types.HealthHealthy,
				Weight:   100,
				Priority: 1,
				Tags:     []string{"web", "frontend"},
				Metadata: map[string]string{
					"region": "us-east-1",
					"zone":   "us-east-1b",
				},
			},
			{
				ID:       "web-3",
				Name:     "web-service",
				Address:  "192.168.1.12",
				Port:     8080,
				Protocol: "http",
				Health:   types.HealthDegraded,
				Weight:   50,
				Priority: 2,
				Tags:     []string{"web", "frontend"},
				Metadata: map[string]string{
					"region": "us-west-1",
					"zone":   "us-west-1a",
				},
			},
		}

		// Register endpoints with Consul provider
		for _, endpoint := range endpoints {
			err := manager.RegisterEndpointWithProvider(ctx, "consul", endpoint)
			assert.NoError(t, err)
		}

		// List endpoints
		registeredEndpoints, err := manager.ListEndpointsWithProvider(ctx, "consul")
		assert.NoError(t, err)
		assert.Len(t, registeredEndpoints, 3)

		// Get specific endpoint
		endpoint, err := manager.GetEndpointWithProvider(ctx, "consul", "web-1")
		assert.NoError(t, err)
		assert.Equal(t, "web-1", endpoint.ID)
		assert.Equal(t, "192.168.1.10", endpoint.Address)
	})

	t.Run("failover configuration", func(t *testing.T) {
		// Create failover configuration
		failoverConfig := &types.FailoverConfig{
			Name:                "web-service-failover",
			Strategy:            types.StrategyRoundRobin,
			HealthCheckInterval: 30 * time.Second,
			HealthCheckTimeout:  5 * time.Second,
			MaxFailures:         3,
			RecoveryTime:        60 * time.Second,
			RetryAttempts:       3,
			RetryDelay:          time.Second,
			Timeout:             30 * time.Second,
			CircuitBreaker: &types.CircuitBreakerConfig{
				Enabled:          true,
				FailureThreshold: 5,
				SuccessThreshold: 3,
				Timeout:          60 * time.Second,
				MaxRequests:      10,
				Interval:         10 * time.Second,
			},
			LoadBalancer: &types.LoadBalancerConfig{
				Algorithm:         types.StrategyRoundRobin,
				StickySession:     false,
				SessionTimeout:    30 * time.Minute,
				MaxConnections:    1000,
				ConnectionTimeout: 10 * time.Second,
				KeepAliveTimeout:  30 * time.Second,
			},
		}

		// Configure failover
		err := manager.ConfigureWithProvider(ctx, "consul", failoverConfig)
		assert.NoError(t, err)

		// Get configuration
		retrievedConfig, err := manager.GetConfigWithProvider(ctx, "consul")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedConfig)
	})

	t.Run("endpoint selection strategies", func(t *testing.T) {
		strategies := []types.FailoverStrategy{
			types.StrategyRoundRobin,
			types.StrategyWeighted,
			types.StrategyRandom,
			types.StrategyHealthBased,
		}

		for _, strategy := range strategies {
			t.Run(string(strategy), func(t *testing.T) {
				config := &types.FailoverConfig{
					Name:     "test-config",
					Strategy: strategy,
				}

				result, err := manager.SelectEndpointWithProvider(ctx, "consul", config)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, strategy, result.Strategy)

				if result.Error == nil {
					assert.NotNil(t, result.Endpoint)
					assert.True(t, result.Duration > 0)
				}
			})
		}
	})

	t.Run("execute with failover", func(t *testing.T) {
		config := &types.FailoverConfig{
			Name:          "test-config",
			Strategy:      types.StrategyRoundRobin,
			RetryAttempts: 3,
			RetryDelay:    100 * time.Millisecond,
		}

		// Successful execution
		result, err := manager.ExecuteWithFailoverWithProvider(ctx, "consul", config, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "success", result.Metadata["result"])
		assert.Equal(t, 1, result.Attempts)

		// Failed execution
		result, err = manager.ExecuteWithFailoverWithProvider(ctx, "consul", config, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
			return nil, assert.AnError
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Equal(t, 3, result.Attempts) // Should retry 3 times
	})

	t.Run("health checks", func(t *testing.T) {
		// Health check single endpoint
		status, err := manager.HealthCheckWithProvider(ctx, "consul", "web-1")
		assert.NoError(t, err)
		assert.Equal(t, types.HealthHealthy, status)

		// Health check all endpoints
		statuses, err := manager.HealthCheckAllWithProvider(ctx, "consul")
		assert.NoError(t, err)
		assert.Len(t, statuses, 3)
		assert.Equal(t, types.HealthHealthy, statuses["web-1"])
		assert.Equal(t, types.HealthHealthy, statuses["web-2"])
		assert.Equal(t, types.HealthDegraded, statuses["web-3"])
	})

	t.Run("statistics and monitoring", func(t *testing.T) {
		// Get statistics
		stats, err := manager.GetStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "consul")
		assert.Contains(t, stats, "kubernetes")

		// Get events
		events, err := manager.GetEvents(ctx, 10)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Contains(t, events, "consul")
		assert.Contains(t, events, "kubernetes")
	})

	t.Run("provider management", func(t *testing.T) {
		// List providers
		providers := manager.ListProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "consul")
		assert.Contains(t, providers, "kubernetes")

		// Get provider info
		info := manager.GetProviderInfo()
		assert.Len(t, info, 2)
		assert.Contains(t, info, "consul")
		assert.Contains(t, info, "kubernetes")

		consulInfo := info["consul"]
		assert.Equal(t, "consul", consulInfo.Name)
		assert.NotEmpty(t, consulInfo.SupportedFeatures)
	})

	t.Run("multi-provider failover", func(t *testing.T) {
		// Add endpoints to Kubernetes provider
		k8sEndpoints := []*types.ServiceEndpoint{
			{
				ID:       "k8s-web-1",
				Name:     "k8s-web-service",
				Address:  "10.0.1.10",
				Port:     8080,
				Protocol: "http",
				Health:   types.HealthHealthy,
				Weight:   100,
			},
			{
				ID:       "k8s-web-2",
				Name:     "k8s-web-service",
				Address:  "10.0.1.11",
				Port:     8080,
				Protocol: "http",
				Health:   types.HealthHealthy,
				Weight:   100,
			},
		}

		for _, endpoint := range k8sEndpoints {
			err := manager.RegisterEndpointWithProvider(ctx, "kubernetes", endpoint)
			assert.NoError(t, err)
		}

		// Test failover between providers
		config := &types.FailoverConfig{
			Name:     "multi-provider-config",
			Strategy: types.StrategyRoundRobin,
		}

		// Test Consul provider
		result, err := manager.SelectEndpointWithProvider(ctx, "consul", config)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Test Kubernetes provider
		result, err = manager.SelectEndpointWithProvider(ctx, "kubernetes", config)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Test default provider (should use Consul as configured)
		result, err = manager.SelectEndpoint(ctx, config)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("cleanup", func(t *testing.T) {
		// Close manager
		err := manager.Close()
		assert.NoError(t, err)
	})
}

func TestFailoverErrorHandling(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewFailoverManager(nil, logger)

	ctx := context.Background()

	t.Run("no providers registered", func(t *testing.T) {
		config := &types.FailoverConfig{
			Name:     "test-config",
			Strategy: types.StrategyRoundRobin,
		}

		_, err := manager.SelectEndpoint(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no providers registered")
	})

	t.Run("non-existing provider", func(t *testing.T) {
		config := &types.FailoverConfig{
			Name:     "test-config",
			Strategy: types.StrategyRoundRobin,
		}

		_, err := manager.SelectEndpointWithProvider(ctx, "non-existing", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("provider connection errors", func(t *testing.T) {
		// Create a mock provider that fails to connect
		failingProvider := mocks.NewMockProvider("failing-provider")
		failingProvider.SetConnectError(assert.AnError)

		err := manager.RegisterProvider(failingProvider)
		assert.NoError(t, err)

		// Try to use the failing provider
		config := &types.FailoverConfig{
			Name:     "test-config",
			Strategy: types.StrategyRoundRobin,
		}

		_, err = manager.SelectEndpointWithProvider(ctx, "failing-provider", config)
		assert.Error(t, err)
	})
}

func TestFailoverPerformance(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewFailoverManager(nil, logger)

	// Create a provider with many endpoints
	provider := mocks.NewMockProvider("performance-provider")

	// Add 100 endpoints
	for i := 0; i < 100; i++ {
		endpoint := &types.ServiceEndpoint{
			ID:       fmt.Sprintf("endpoint-%d", i),
			Name:     "performance-service",
			Address:  fmt.Sprintf("192.168.1.%d", i+1),
			Port:     8080,
			Protocol: "http",
			Health:   types.HealthHealthy,
			Weight:   100,
		}
		provider.AddEndpoint(endpoint)
	}

	err := manager.RegisterProvider(provider)
	require.NoError(t, err)

	ctx := context.Background()
	config := &types.FailoverConfig{
		Name:     "performance-config",
		Strategy: types.StrategyRoundRobin,
	}

	t.Run("endpoint selection performance", func(t *testing.T) {
		iterations := 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			result, err := manager.SelectEndpointWithProvider(ctx, "performance-provider", config)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		t.Logf("Selected %d endpoints in %v (avg: %v per selection)", iterations, duration, avgDuration)
		assert.True(t, avgDuration < 10*time.Millisecond, "Average selection time should be less than 10ms")
	})

	t.Run("concurrent endpoint selection", func(t *testing.T) {
		concurrency := 10
		iterations := 100
		done := make(chan bool, concurrency)

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			go func() {
				for j := 0; j < iterations; j++ {
					result, err := manager.SelectEndpointWithProvider(ctx, "performance-provider", config)
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}

		duration := time.Since(start)
		totalOperations := concurrency * iterations

		t.Logf("Completed %d concurrent selections in %v", totalOperations, duration)
		assert.True(t, duration < 5*time.Second, "Concurrent operations should complete within 5 seconds")
	})
}
