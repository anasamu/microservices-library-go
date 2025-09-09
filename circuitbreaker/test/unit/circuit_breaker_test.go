package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/gateway"
	"github.com/anasamu/microservices-library-go/circuitbreaker/providers/custom"
	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Create manager
	config := &gateway.ManagerConfig{
		DefaultProvider: "custom",
		RetryAttempts:   1,
		RetryDelay:      10 * time.Millisecond,
		Timeout:         100 * time.Millisecond,
		FallbackEnabled: true,
	}
	manager := gateway.NewCircuitBreakerManager(config, logger)

	// Create and register custom provider
	customProvider := custom.NewCustomProvider(nil, logger)
	err := manager.RegisterProvider(customProvider)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Basic Execution", func(t *testing.T) {
		result, err := manager.Execute(ctx, "test-service", func() (interface{}, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result.Result)
		assert.Equal(t, types.StateClosed, result.State)
		assert.False(t, result.Fallback)
	})

	t.Run("Execution with Error", func(t *testing.T) {
		result, err := manager.Execute(ctx, "test-service", func() (interface{}, error) {
			return nil, errors.New("service error")
		})

		assert.NoError(t, err) // Manager doesn't return error, it's in the result
		assert.Error(t, result.Error)
		assert.Equal(t, "service error", result.Error.Error())
	})

	t.Run("Execution with Fallback", func(t *testing.T) {
		result, err := manager.ExecuteWithFallback(ctx, "test-service",
			func() (interface{}, error) {
				return nil, errors.New("service error")
			},
			func() (interface{}, error) {
				return "fallback result", nil
			},
		)

		assert.NoError(t, err)
		assert.Equal(t, "fallback result", result.Result)
		assert.True(t, result.Fallback)
	})

	t.Run("Get State", func(t *testing.T) {
		state, err := manager.GetState(ctx, "test-service")
		assert.NoError(t, err)
		assert.Equal(t, types.StateClosed, state)
	})

	t.Run("Get Stats", func(t *testing.T) {
		stats, err := manager.GetStats(ctx, "test-service")
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, "custom", stats.Provider)
	})

	t.Run("Configure Circuit Breaker", func(t *testing.T) {
		config := &types.CircuitBreakerConfig{
			Name:                "test-service",
			MaxRequests:         5,
			Interval:            5 * time.Second,
			Timeout:             10 * time.Second,
			MaxConsecutiveFails: 2,
			FailureThreshold:    0.5,
			SuccessThreshold:    2,
			ReadyToTrip: func(counts types.Counts) bool {
				return counts.ConsecutiveFailures >= 2
			},
			IsSuccessful: func(err error) bool {
				return err == nil
			},
		}

		err := manager.Configure(ctx, "test-service", config)
		assert.NoError(t, err)
	})

	t.Run("Reset Circuit Breaker", func(t *testing.T) {
		err := manager.Reset(ctx, "test-service")
		assert.NoError(t, err)
	})

	t.Run("List Providers", func(t *testing.T) {
		providers := manager.ListProviders()
		assert.Contains(t, providers, "custom")
	})

	t.Run("Get Provider Info", func(t *testing.T) {
		info := manager.GetProviderInfo()
		assert.Contains(t, info, "custom")
		assert.Equal(t, "custom", info["custom"].Name)
		assert.True(t, info["custom"].IsConnected)
	})
}

func TestCircuitBreakerStates(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create custom provider directly for more control
	provider := custom.NewCustomProvider(nil, logger)
	ctx := context.Background()
	err := provider.Connect(ctx)
	require.NoError(t, err)

	// Configure circuit breaker to trip easily
	config := &types.CircuitBreakerConfig{
		Name:                "test-service",
		MaxRequests:         10,
		Interval:            10 * time.Second,
		Timeout:             1 * time.Second,
		MaxConsecutiveFails: 2,
		FailureThreshold:    0.5,
		SuccessThreshold:    2,
		ReadyToTrip: func(counts types.Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}

	err = provider.Configure(ctx, "test-service", config)
	require.NoError(t, err)

	t.Run("Circuit Opens After Failures", func(t *testing.T) {
		// Execute failing functions to trigger circuit opening
		for i := 0; i < 3; i++ {
			_, _ = provider.Execute(ctx, "test-service", func() (interface{}, error) {
				return nil, errors.New("service error")
			})
			// Don't check error here as we expect failures
		}

		// Check that circuit is now open
		state, err := provider.GetState(ctx, "test-service")
		assert.NoError(t, err)
		assert.Equal(t, types.StateOpen, state)
	})

	t.Run("Circuit Rejects Requests When Open", func(t *testing.T) {
		result, err := provider.Execute(ctx, "test-service", func() (interface{}, error) {
			return "should not execute", nil
		})

		assert.NoError(t, err) // Manager doesn't return error
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "Circuit breaker is open")
	})

	t.Run("Circuit Moves to Half-Open After Timeout", func(t *testing.T) {
		// Wait for timeout
		time.Sleep(1100 * time.Millisecond)

		// Execute a successful function
		result, err := provider.Execute(ctx, "test-service", func() (interface{}, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.NoError(t, result.Error)
		assert.Equal(t, "success", result.Result)

		// Check that circuit moved to half-open and then closed
		state, err := provider.GetState(ctx, "test-service")
		assert.NoError(t, err)
		assert.Equal(t, types.StateClosed, state)
	})
}

func TestCustomProvider(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	provider := custom.NewCustomProvider(nil, logger)
	ctx := context.Background()

	t.Run("Connection Management", func(t *testing.T) {
		assert.False(t, provider.IsConnected())

		err := provider.Connect(ctx)
		assert.NoError(t, err)
		assert.True(t, provider.IsConnected())

		err = provider.Ping(ctx)
		assert.NoError(t, err)

		err = provider.Disconnect(ctx)
		assert.NoError(t, err)
		assert.False(t, provider.IsConnected())
	})

	t.Run("Provider Information", func(t *testing.T) {
		assert.Equal(t, "custom", provider.GetName())

		features := provider.GetSupportedFeatures()
		assert.Contains(t, features, types.FeatureExecute)
		assert.Contains(t, features, types.FeatureState)
		assert.Contains(t, features, types.FeatureMetrics)

		connInfo := provider.GetConnectionInfo()
		assert.Equal(t, "custom", connInfo.Database)
		assert.Equal(t, types.StatusDisconnected, connInfo.Status)
	})

	t.Run("Configuration Management", func(t *testing.T) {
		err := provider.Connect(ctx)
		require.NoError(t, err)

		config := &types.CircuitBreakerConfig{
			Name:                "test-service",
			MaxRequests:         5,
			Interval:            5 * time.Second,
			Timeout:             10 * time.Second,
			MaxConsecutiveFails: 3,
			FailureThreshold:    0.6,
			SuccessThreshold:    2,
		}

		err = provider.Configure(ctx, "test-service", config)
		assert.NoError(t, err)

		retrievedConfig, err := provider.GetConfig(ctx, "test-service")
		assert.NoError(t, err)
		assert.Equal(t, config.Name, retrievedConfig.Name)
		assert.Equal(t, config.MaxRequests, retrievedConfig.MaxRequests)
	})

	t.Run("Bulk Operations", func(t *testing.T) {
		err := provider.Connect(ctx)
		require.NoError(t, err)

		// Configure multiple circuit breakers
		for i := 0; i < 3; i++ {
			config := &types.CircuitBreakerConfig{
				Name:                fmt.Sprintf("service-%d", i),
				MaxRequests:         5,
				Interval:            5 * time.Second,
				Timeout:             10 * time.Second,
				MaxConsecutiveFails: 2,
			}
			err = provider.Configure(ctx, fmt.Sprintf("service-%d", i), config)
			require.NoError(t, err)
		}

		// Get all states
		states, err := provider.GetAllStates(ctx)
		assert.NoError(t, err)
		assert.Len(t, states, 3)

		// Get all stats
		stats, err := provider.GetAllStats(ctx)
		assert.NoError(t, err)
		assert.Len(t, stats, 3)

		// Reset all
		err = provider.ResetAll(ctx)
		assert.NoError(t, err)
	})
}
