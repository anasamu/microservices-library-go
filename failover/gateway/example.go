package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/sirupsen/logrus"
)

// Example demonstrates how to use the FailoverManager
func Example() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager configuration
	config := &ManagerConfig{
		DefaultProvider: "consul",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
		Metadata: map[string]string{
			"environment": "production",
			"version":     "1.0.0",
		},
	}

	// Create failover manager
	manager := NewFailoverManager(config, logger)

	// Note: In a real application, you would register actual providers here
	// For example:
	// consulProvider := consul.NewProvider(consulConfig)
	// err := manager.RegisterProvider(consulProvider)
	// if err != nil {
	//     log.Fatal("Failed to register Consul provider:", err)
	// }

	ctx := context.Background()

	// Example: Create service endpoints
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

	// Example: Register endpoints
	for _, endpoint := range endpoints {
		err := manager.RegisterEndpoint(ctx, endpoint)
		if err != nil {
			log.Printf("Failed to register endpoint %s: %v", endpoint.ID, err)
		} else {
			log.Printf("Successfully registered endpoint: %s", endpoint.ID)
		}
	}

	// Example: Create failover configuration
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
		Metadata: map[string]string{
			"service": "web-service",
			"version": "1.0.0",
		},
	}

	// Example: Configure failover
	err := manager.Configure(ctx, failoverConfig)
	if err != nil {
		log.Printf("Failed to configure failover: %v", err)
	} else {
		log.Println("Successfully configured failover")
	}

	// Example: Select endpoint
	result, err := manager.SelectEndpoint(ctx, failoverConfig)
	if err != nil {
		log.Printf("Failed to select endpoint: %v", err)
	} else {
		log.Printf("Selected endpoint: %s (%s:%d)",
			result.Endpoint.ID, result.Endpoint.Address, result.Endpoint.Port)
		log.Printf("Strategy used: %s", result.Strategy)
		log.Printf("Duration: %v", result.Duration)
	}

	// Example: Execute with failover
	executionResult, err := manager.ExecuteWithFailover(ctx, failoverConfig, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
		// Simulate a service call
		log.Printf("Executing request on endpoint: %s", endpoint.ID)

		// Simulate different response times based on health
		switch endpoint.Health {
		case types.HealthHealthy:
			time.Sleep(100 * time.Millisecond)
			return "success", nil
		case types.HealthDegraded:
			time.Sleep(500 * time.Millisecond)
			return "degraded", nil
		case types.HealthUnhealthy:
			return nil, fmt.Errorf("endpoint is unhealthy")
		default:
			return nil, fmt.Errorf("unknown health status")
		}
	})

	if err != nil {
		log.Printf("Failed to execute with failover: %v", err)
	} else {
		log.Printf("Execution result: %v", executionResult)
		log.Printf("Endpoint used: %s", executionResult.Endpoint.ID)
		log.Printf("Attempts: %d", executionResult.Attempts)
		log.Printf("Duration: %v", executionResult.Duration)
	}

	// Example: Health check
	healthStatus, err := manager.HealthCheck(ctx, "web-1")
	if err != nil {
		log.Printf("Failed to perform health check: %v", err)
	} else {
		log.Printf("Health status of web-1: %s", healthStatus)
	}

	// Example: Health check all endpoints
	allHealth, err := manager.HealthCheckAll(ctx)
	if err != nil {
		log.Printf("Failed to perform health check on all endpoints: %v", err)
	} else {
		log.Println("Health status of all endpoints:")
		for endpointID, status := range allHealth {
			log.Printf("  %s: %s", endpointID, status)
		}
	}

	// Example: List endpoints
	endpointList, err := manager.ListEndpoints(ctx)
	if err != nil {
		log.Printf("Failed to list endpoints: %v", err)
	} else {
		log.Printf("Registered endpoints (%d):", len(endpointList))
		for _, endpoint := range endpointList {
			log.Printf("  %s: %s:%d (%s)",
				endpoint.ID, endpoint.Address, endpoint.Port, endpoint.Health)
		}
	}

	// Example: Get statistics
	stats, err := manager.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get statistics: %v", err)
	} else {
		log.Println("Failover statistics:")
		for providerName, stat := range stats {
			log.Printf("  Provider %s:", providerName)
			log.Printf("    Total requests: %d", stat.TotalRequests)
			log.Printf("    Successful requests: %d", stat.SuccessfulRequests)
			log.Printf("    Failed requests: %d", stat.FailedRequests)
			log.Printf("    Failovers: %d", stat.Failovers)
			log.Printf("    Recoveries: %d", stat.Recoveries)
		}
	}

	// Example: Get events
	events, err := manager.GetEvents(ctx, 10)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
	} else {
		log.Println("Recent failover events:")
		for providerName, eventList := range events {
			log.Printf("  Provider %s:", providerName)
			for _, event := range eventList {
				log.Printf("    %s: %s at %v",
					event.Type, event.Name, event.Timestamp)
			}
		}
	}

	// Example: List providers
	providers := manager.ListProviders()
	log.Printf("Registered providers: %v", providers)

	// Example: Get provider info
	providerInfo := manager.GetProviderInfo()
	log.Println("Provider information:")
	for name, info := range providerInfo {
		log.Printf("  %s:", name)
		log.Printf("    Name: %s", info.Name)
		log.Printf("    Connected: %t", info.IsConnected)
		log.Printf("    Features: %v", info.SupportedFeatures)
	}

	// Clean up
	err = manager.Close()
	if err != nil {
		log.Printf("Failed to close manager: %v", err)
	} else {
		log.Println("Manager closed successfully")
	}
}
