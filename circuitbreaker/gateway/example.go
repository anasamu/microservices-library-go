package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
)

// Example demonstrates how to use the CircuitBreakerManager
func Example() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager configuration
	config := &ManagerConfig{
		DefaultProvider: "gobreaker",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
		Metadata: map[string]string{
			"environment": "development",
			"version":     "1.0.0",
		},
	}

	// Create circuit breaker manager
	manager := NewCircuitBreakerManager(config, logger)

	// Note: In a real application, you would register actual providers here
	// For example:
	// gobreakerProvider := gobreaker.NewProvider(gobreakerConfig, logger)
	// err := manager.RegisterProvider(gobreakerProvider)
	// if err != nil {
	//     logger.Fatal("Failed to register gobreaker provider:", err)
	// }

	ctx := context.Background()

	// Example 1: Basic execution
	fmt.Println("=== Example 1: Basic Execution ===")
	result, err := manager.Execute(ctx, "example-service", func() (interface{}, error) {
		// Simulate some work
		time.Sleep(100 * time.Millisecond)
		return "success", nil
	})
	if err != nil {
		logger.WithError(err).Error("Execution failed")
	} else {
		logger.WithField("result", result.Result).Info("Execution successful")
	}

	// Example 2: Execution with fallback
	fmt.Println("\n=== Example 2: Execution with Fallback ===")
	result, err = manager.ExecuteWithFallback(ctx, "example-service",
		func() (interface{}, error) {
			// Simulate a failure
			return nil, fmt.Errorf("service unavailable")
		},
		func() (interface{}, error) {
			// Fallback function
			return "fallback result", nil
		},
	)
	if err != nil {
		logger.WithError(err).Error("Execution with fallback failed")
	} else {
		logger.WithFields(logrus.Fields{
			"result":   result.Result,
			"fallback": result.Fallback,
		}).Info("Execution with fallback successful")
	}

	// Example 3: Get circuit breaker state
	fmt.Println("\n=== Example 3: Get Circuit Breaker State ===")
	state, err := manager.GetState(ctx, "example-service")
	if err != nil {
		logger.WithError(err).Error("Failed to get state")
	} else {
		logger.WithField("state", state).Info("Circuit breaker state")
	}

	// Example 4: Get circuit breaker statistics
	fmt.Println("\n=== Example 4: Get Circuit Breaker Statistics ===")
	stats, err := manager.GetStats(ctx, "example-service")
	if err != nil {
		logger.WithError(err).Error("Failed to get stats")
	} else {
		logger.WithFields(logrus.Fields{
			"requests":  stats.Requests,
			"successes": stats.Successes,
			"failures":  stats.Failures,
			"state":     stats.State,
			"uptime":    stats.Uptime,
		}).Info("Circuit breaker statistics")
	}

	// Example 5: Configure circuit breaker
	fmt.Println("\n=== Example 5: Configure Circuit Breaker ===")
	cbConfig := &types.CircuitBreakerConfig{
		Name:                "example-service",
		MaxRequests:         5,
		Interval:            10 * time.Second,
		Timeout:             30 * time.Second,
		MaxConsecutiveFails: 3,
		FailureThreshold:    0.6,
		SuccessThreshold:    2,
		FallbackEnabled:     true,
		RetryEnabled:        true,
		RetryAttempts:       2,
		RetryDelay:          500 * time.Millisecond,
		ReadyToTrip: func(counts types.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
		Metadata: map[string]string{
			"service": "example-service",
			"version": "1.0.0",
		},
	}

	err = manager.Configure(ctx, "example-service", cbConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to configure circuit breaker")
	} else {
		logger.Info("Circuit breaker configured successfully")
	}

	// Example 6: Get all states from all providers
	fmt.Println("\n=== Example 6: Get All States ===")
	allStates, err := manager.GetAllStates(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get all states")
	} else {
		for provider, states := range allStates {
			for name, state := range states {
				logger.WithFields(logrus.Fields{
					"provider": provider,
					"name":     name,
					"state":    state,
				}).Info("Circuit breaker state")
			}
		}
	}

	// Example 7: Get all statistics from all providers
	fmt.Println("\n=== Example 7: Get All Statistics ===")
	allStats, err := manager.GetAllStats(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get all statistics")
	} else {
		for provider, stats := range allStats {
			for name, stat := range stats {
				logger.WithFields(logrus.Fields{
					"provider":  provider,
					"name":      name,
					"requests":  stat.Requests,
					"successes": stat.Successes,
					"failures":  stat.Failures,
					"state":     stat.State,
				}).Info("Circuit breaker statistics")
			}
		}
	}

	// Example 8: List providers
	fmt.Println("\n=== Example 8: List Providers ===")
	providers := manager.ListProviders()
	logger.WithField("providers", providers).Info("Registered providers")

	// Example 9: Get provider information
	fmt.Println("\n=== Example 9: Get Provider Information ===")
	providerInfo := manager.GetProviderInfo()
	for name, info := range providerInfo {
		logger.WithFields(logrus.Fields{
			"name":              name,
			"features":          info.SupportedFeatures,
			"connected":         info.IsConnected,
			"connection_status": info.ConnectionInfo.Status,
		}).Info("Provider information")
	}

	// Example 10: Reset circuit breaker
	fmt.Println("\n=== Example 10: Reset Circuit Breaker ===")
	err = manager.Reset(ctx, "example-service")
	if err != nil {
		logger.WithError(err).Error("Failed to reset circuit breaker")
	} else {
		logger.Info("Circuit breaker reset successfully")
	}

	// Example 11: Reset all circuit breakers
	fmt.Println("\n=== Example 11: Reset All Circuit Breakers ===")
	err = manager.ResetAll(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to reset all circuit breakers")
	} else {
		logger.Info("All circuit breakers reset successfully")
	}

	// Clean up
	fmt.Println("\n=== Cleanup ===")
	err = manager.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close manager")
	} else {
		logger.Info("Manager closed successfully")
	}
}

// ExampleWithErrorHandling demonstrates error handling patterns
func ExampleWithErrorHandling() {
	logger := logrus.New()
	manager := NewCircuitBreakerManager(nil, logger)
	ctx := context.Background()

	// Example of handling different types of errors
	result, err := manager.Execute(ctx, "test-service", func() (interface{}, error) {
		// Simulate different error scenarios
		return nil, &types.CircuitBreakerError{
			Code:    types.ErrCodeCircuitOpen,
			Message: "Circuit breaker is open",
			State:   types.StateOpen,
		}
	})

	if err != nil {
		if cbErr, ok := err.(*types.CircuitBreakerError); ok {
			switch cbErr.Code {
			case types.ErrCodeCircuitOpen:
				logger.WithField("state", cbErr.State).Warn("Circuit breaker is open, using fallback")
			case types.ErrCodeTimeout:
				logger.Warn("Operation timed out")
			case types.ErrCodeMaxRequests:
				logger.Warn("Maximum requests exceeded")
			default:
				logger.WithError(err).Error("Circuit breaker error")
			}
		} else {
			logger.WithError(err).Error("Unexpected error")
		}
	} else {
		logger.WithField("result", result.Result).Info("Operation successful")
	}
}

// ExampleWithCustomConfiguration demonstrates custom configuration
func ExampleWithCustomConfiguration() {
	logger := logrus.New()
	manager := NewCircuitBreakerManager(nil, logger)
	ctx := context.Background()

	// Custom configuration for a specific service
	config := &types.CircuitBreakerConfig{
		Name:                "critical-service",
		MaxRequests:         1, // Very restrictive
		Interval:            5 * time.Second,
		Timeout:             10 * time.Second,
		MaxConsecutiveFails: 2,
		FailureThreshold:    0.3, // 30% failure rate
		SuccessThreshold:    5,
		FallbackEnabled:     true,
		RetryEnabled:        false, // No retries for critical service
		ReadyToTrip: func(counts types.Counts) bool {
			// Custom trip logic
			return counts.ConsecutiveFailures >= 2 ||
				(counts.TotalFailures > 0 && float64(counts.TotalFailures)/float64(counts.Requests) > 0.3)
		},
		IsSuccessful: func(err error) bool {
			// Custom success logic - only nil errors are considered successful
			return err == nil
		},
		OnStateChange: func(name string, from types.CircuitState, to types.CircuitState) {
			logger.WithFields(logrus.Fields{
				"name": name,
				"from": from,
				"to":   to,
			}).Warn("Circuit breaker state changed")
		},
		Metadata: map[string]string{
			"service":     "critical-service",
			"environment": "production",
			"team":        "platform",
		},
	}

	err := manager.Configure(ctx, "critical-service", config)
	if err != nil {
		logger.WithError(err).Fatal("Failed to configure critical service circuit breaker")
	}

	logger.Info("Critical service circuit breaker configured with custom settings")
}
