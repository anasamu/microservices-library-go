package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker"
	"github.com/anasamu/microservices-library-go/circuitbreaker/providers/custom"
	"github.com/anasamu/microservices-library-go/circuitbreaker/providers/gobreaker"
	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create manager with multiple providers
	config := &gateway.ManagerConfig{
		DefaultProvider: "custom",
		RetryAttempts:   2,
		RetryDelay:      10 * time.Millisecond,
		Timeout:         100 * time.Millisecond,
		FallbackEnabled: true,
	}
	manager := gateway.NewCircuitBreakerManager(config, logger)

	// Register custom provider
	customProvider := custom.NewCustomProvider(nil, logger)
	err := manager.RegisterProvider(customProvider)
	require.NoError(t, err)

	// Register gobreaker provider
	gobreakerProvider := gobreaker.NewGobreakerProvider(nil, logger)
	err = manager.RegisterProvider(gobreakerProvider)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Multi-Provider Execution", func(t *testing.T) {
		// Test execution with custom provider
		result, err := manager.ExecuteWithProvider(ctx, "custom", "test-service", func() (interface{}, error) {
			return "custom result", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "custom result", result.Result)

		// Test execution with gobreaker provider
		result, err = manager.ExecuteWithProvider(ctx, "gobreaker", "test-service", func() (interface{}, error) {
			return "gobreaker result", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "gobreaker result", result.Result)
	})

	t.Run("Provider Switching", func(t *testing.T) {
		// Configure circuit breaker in custom provider to fail
		customConfig := &types.CircuitBreakerConfig{
			Name:                "failing-service",
			MaxRequests:         5,
			Interval:            5 * time.Second,
			Timeout:             1 * time.Second,
			MaxConsecutiveFails: 1,
			FailureThreshold:    0.5,
			SuccessThreshold:    2,
			ReadyToTrip: func(counts types.Counts) bool {
				return counts.ConsecutiveFailures >= 1
			},
			IsSuccessful: func(err error) bool {
				return err == nil
			},
		}

		err := manager.ConfigureWithProvider(ctx, "custom", "failing-service", customConfig)
		require.NoError(t, err)

		// Make custom provider fail
		_, err = manager.ExecuteWithProvider(ctx, "custom", "failing-service", func() (interface{}, error) {
			return nil, errors.New("service error")
		})
		assert.NoError(t, err) // Manager doesn't return error

		// Check that custom provider circuit is open
		state, err := manager.GetStateWithProvider(ctx, "custom", "failing-service")
		assert.NoError(t, err)
		assert.Equal(t, types.StateOpen, state)

		// Gobreaker provider should still work
		result, err := manager.ExecuteWithProvider(ctx, "gobreaker", "failing-service", func() (interface{}, error) {
			return "gobreaker success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "gobreaker success", result.Result)
	})

	t.Run("Fallback Across Providers", func(t *testing.T) {
		// Test fallback with custom provider
		result, err := manager.ExecuteWithFallbackAndProvider(ctx, "custom", "test-service",
			func() (interface{}, error) {
				return nil, errors.New("primary service error")
			},
			func() (interface{}, error) {
				return "fallback from custom", nil
			},
		)
		assert.NoError(t, err)
		assert.Equal(t, "fallback from custom", result.Result)
		assert.True(t, result.Fallback)

		// Test fallback with gobreaker provider
		result, err = manager.ExecuteWithFallbackAndProvider(ctx, "gobreaker", "test-service",
			func() (interface{}, error) {
				return nil, errors.New("primary service error")
			},
			func() (interface{}, error) {
				return "fallback from gobreaker", nil
			},
		)
		assert.NoError(t, err)
		assert.Equal(t, "fallback from gobreaker", result.Result)
		assert.True(t, result.Fallback)
	})

	t.Run("Statistics Aggregation", func(t *testing.T) {
		// Get all states from all providers
		allStates, err := manager.GetAllStates(ctx)
		assert.NoError(t, err)
		assert.Contains(t, allStates, "custom")
		assert.Contains(t, allStates, "gobreaker")

		// Get all statistics from all providers
		allStats, err := manager.GetAllStats(ctx)
		assert.NoError(t, err)
		assert.Contains(t, allStats, "custom")
		assert.Contains(t, allStats, "gobreaker")

		// Verify statistics structure
		for provider, stats := range allStats {
			for name, stat := range stats {
				assert.NotNil(t, stat)
				assert.Equal(t, provider, stat.Provider)
				assert.NotEmpty(t, name)
			}
		}
	})

	t.Run("Provider Management", func(t *testing.T) {
		// List providers
		providers := manager.ListProviders()
		assert.Contains(t, providers, "custom")
		assert.Contains(t, providers, "gobreaker")

		// Get provider info
		info := manager.GetProviderInfo()
		assert.Contains(t, info, "custom")
		assert.Contains(t, info, "gobreaker")

		// Verify provider info
		for name, providerInfo := range info {
			assert.NotNil(t, providerInfo)
			assert.Equal(t, name, providerInfo.Name)
			assert.NotEmpty(t, providerInfo.SupportedFeatures)
			assert.NotNil(t, providerInfo.ConnectionInfo)
		}
	})

	t.Run("Reset Operations", func(t *testing.T) {
		// Reset specific circuit breaker
		err := manager.ResetWithProvider(ctx, "custom", "test-service")
		assert.NoError(t, err)

		err = manager.ResetWithProvider(ctx, "gobreaker", "test-service")
		assert.NoError(t, err)

		// Reset all circuit breakers
		err = manager.ResetAll(ctx)
		assert.NoError(t, err)
	})

	t.Run("Configuration Management", func(t *testing.T) {
		// Configure circuit breaker in both providers
		config := &types.CircuitBreakerConfig{
			Name:                "config-test",
			MaxRequests:         10,
			Interval:            10 * time.Second,
			Timeout:             30 * time.Second,
			MaxConsecutiveFails: 5,
			FailureThreshold:    0.6,
			SuccessThreshold:    3,
		}

		err := manager.ConfigureWithProvider(ctx, "custom", "config-test", config)
		assert.NoError(t, err)

		err = manager.ConfigureWithProvider(ctx, "gobreaker", "config-test", config)
		assert.NoError(t, err)

		// Retrieve configurations
		customConfig, err := manager.GetConfigWithProvider(ctx, "custom", "config-test")
		assert.NoError(t, err)
		assert.Equal(t, config.Name, customConfig.Name)

		gobreakerConfig, err := manager.GetConfigWithProvider(ctx, "gobreaker", "config-test")
		assert.NoError(t, err)
		assert.Equal(t, config.Name, gobreakerConfig.Name)
	})

	t.Run("Error Handling", func(t *testing.T) {
		// Test with non-existent provider
		_, err := manager.ExecuteWithProvider(ctx, "non-existent", "test-service", func() (interface{}, error) {
			return "should not execute", nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider non-existent not found")

		// Test with non-existent circuit breaker
		state, err := manager.GetStateWithProvider(ctx, "custom", "non-existent")
		assert.NoError(t, err)
		assert.Equal(t, types.StateClosed, state) // Default state
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Close manager
		err := manager.Close()
		assert.NoError(t, err)
	})
}

func TestCircuitBreakerStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := gateway.NewCircuitBreakerManager(nil, logger)
	customProvider := custom.NewCustomProvider(nil, logger)
	err := manager.RegisterProvider(customProvider)
	require.NoError(t, err)

	ctx := context.Background()

	// Configure circuit breaker for stress test
	config := &types.CircuitBreakerConfig{
		Name:                "stress-test",
		MaxRequests:         100,
		Interval:            1 * time.Second,
		Timeout:             100 * time.Millisecond,
		MaxConsecutiveFails: 10,
		FailureThreshold:    0.8,
		SuccessThreshold:    5,
		ReadyToTrip: func(counts types.Counts) bool {
			return counts.ConsecutiveFailures >= 10
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}

	err = manager.Configure(ctx, "stress-test", config)
	require.NoError(t, err)

	t.Run("Concurrent Execution", func(t *testing.T) {
		const numGoroutines = 10
		const numRequests = 100

		results := make(chan *types.ExecutionResult, numGoroutines*numRequests)
		errors := make(chan error, numGoroutines*numRequests)

		// Start multiple goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numRequests; j++ {
					result, err := manager.Execute(ctx, "stress-test", func() (interface{}, error) {
						// Simulate some work
						time.Sleep(1 * time.Millisecond)
						if j%10 == 0 { // 10% failure rate
							return nil, fmt.Errorf("simulated error")
						}
						return "success", nil
					})
					results <- result
					errors <- err
				}
			}(i)
		}

		// Collect results
		successCount := 0
		failureCount := 0
		fallbackCount := 0

		for i := 0; i < numGoroutines*numRequests; i++ {
			result := <-results
			err := <-errors

			assert.NoError(t, err) // Manager doesn't return errors

			if result.Error != nil {
				failureCount++
			} else {
				successCount++
			}

			if result.Fallback {
				fallbackCount++
			}
		}

		t.Logf("Success: %d, Failures: %d, Fallbacks: %d", successCount, failureCount, fallbackCount)

		// Verify we got some results
		assert.Greater(t, successCount+failureCount, 0)
	})

	t.Run("Circuit State Transitions", func(t *testing.T) {
		// Reset circuit breaker
		err := manager.Reset(ctx, "stress-test")
		require.NoError(t, err)

		// Force circuit to open by sending many failures
		for i := 0; i < 15; i++ {
			_, err := manager.Execute(ctx, "stress-test", func() (interface{}, error) {
				return nil, errors.New("forced failure")
			})
			assert.NoError(t, err) // Manager doesn't return errors
		}

		// Check that circuit is open
		state, err := manager.GetState(ctx, "stress-test")
		assert.NoError(t, err)
		assert.Equal(t, types.StateOpen, state)

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Send successful request to move to half-open
		result, err := manager.Execute(ctx, "stress-test", func() (interface{}, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.NoError(t, result.Error)
		assert.Equal(t, "success", result.Result)

		// Check that circuit moved to closed
		state, err = manager.GetState(ctx, "stress-test")
		assert.NoError(t, err)
		assert.Equal(t, types.StateClosed, state)
	})
}
