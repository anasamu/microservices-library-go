# Failover Library

A comprehensive failover and load balancing library for Go microservices, providing high availability, fault tolerance, and intelligent traffic routing capabilities.

## Features

- **Multiple Failover Strategies**: Round-robin, weighted, least connections, random, health-based, and custom strategies
- **Provider Support**: Consul and Kubernetes providers with extensible architecture
- **Health Monitoring**: Built-in health checks and endpoint monitoring
- **Circuit Breaker**: Automatic circuit breaker pattern implementation
- **Retry Logic**: Configurable retry mechanisms with exponential backoff
- **Load Balancing**: Advanced load balancing algorithms
- **Service Discovery**: Integration with service discovery systems
- **Monitoring & Observability**: Comprehensive statistics and event tracking
- **High Performance**: Optimized for high-throughput scenarios

## Architecture

```
failover/
├── gateway/                    # Core failover gateway
│   ├── manager.go             # Failover manager
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Failover provider implementations
│   ├── consul/                # Consul-based failover
│   │   ├── provider.go        # Consul implementation
│   │   └── go.mod             # Consul dependencies
│   └── kubernetes/            # Kubernetes-based failover
│       ├── provider.go        # Kubernetes implementation
│       └── go.mod             # Kubernetes dependencies
├── test/                      # Test files
│   ├── integration/           # Integration tests
│   ├── unit/                  # Unit tests
│   └── mocks/                 # Mock providers
├── types/                     # Type definitions
│   ├── types.go               # Core types and interfaces
│   └── go.mod                 # Types dependencies
├── go.mod                     # Main module dependencies
└── README.md                  # This documentation
```

## Quick Start

### Installation

```bash
go get github.com/anasamu/microservices-library-go/failover
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/anasamu/microservices-library-go/failover"
    "github.com/anasamu/microservices-library-go/failover/providers/consul"
    "github.com/anasamu/microservices-library-go/failover/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()

    // Create manager configuration
    config := &gateway.ManagerConfig{
        DefaultProvider: "consul",
        RetryAttempts:   3,
        RetryDelay:      time.Second,
        Timeout:         30 * time.Second,
        FallbackEnabled: true,
    }

    // Create failover manager
    manager := gateway.NewFailoverManager(config, logger)

    // Create and register Consul provider
    consulConfig := &consul.Config{
        Address:    "localhost:8500",
        Datacenter: "dc1",
        Timeout:    10 * time.Second,
    }
    
    consulProvider, err := consul.NewProvider(consulConfig, logger)
    if err != nil {
        log.Fatal("Failed to create Consul provider:", err)
    }

    err = manager.RegisterProvider(consulProvider)
    if err != nil {
        log.Fatal("Failed to register Consul provider:", err)
    }

    // Connect to Consul
    ctx := context.Background()
    err = consulProvider.Connect(ctx)
    if err != nil {
        log.Fatal("Failed to connect to Consul:", err)
    }

    // Register service endpoints
    endpoints := []*types.ServiceEndpoint{
        {
            ID:       "web-1",
            Name:     "web-service",
            Address:  "192.168.1.10",
            Port:     8080,
            Protocol: "http",
            Health:   types.HealthHealthy,
            Weight:   100,
            Tags:     []string{"web", "frontend"},
        },
        {
            ID:       "web-2",
            Name:     "web-service",
            Address:  "192.168.1.11",
            Port:     8080,
            Protocol: "http",
            Health:   types.HealthHealthy,
            Weight:   100,
            Tags:     []string{"web", "frontend"},
        },
    }

    for _, endpoint := range endpoints {
        err := manager.RegisterEndpoint(ctx, endpoint)
        if err != nil {
            log.Printf("Failed to register endpoint %s: %v", endpoint.ID, err)
        }
    }

    // Configure failover
    failoverConfig := &types.FailoverConfig{
        Name:                "web-service-failover",
        Strategy:            types.StrategyRoundRobin,
        HealthCheckInterval: 30 * time.Second,
        HealthCheckTimeout:  5 * time.Second,
        MaxFailures:         3,
        RetryAttempts:       3,
        RetryDelay:          time.Second,
        Timeout:             30 * time.Second,
    }

    err = manager.Configure(ctx, failoverConfig)
    if err != nil {
        log.Fatal("Failed to configure failover:", err)
    }

    // Execute with failover
    result, err := manager.ExecuteWithFailover(ctx, failoverConfig, func(endpoint *types.ServiceEndpoint) (interface{}, error) {
        // Your service call logic here
        log.Printf("Executing request on endpoint: %s", endpoint.ID)
        return "success", nil
    })

    if err != nil {
        log.Printf("Failed to execute with failover: %v", err)
    } else {
        log.Printf("Execution result: %v", result)
    }

    // Clean up
    err = manager.Close()
    if err != nil {
        log.Printf("Failed to close manager: %v", err)
    }
}
```

## Configuration

### Manager Configuration

```go
type ManagerConfig struct {
    DefaultProvider string            `json:"default_provider"`
    RetryAttempts   int               `json:"retry_attempts"`
    RetryDelay      time.Duration     `json:"retry_delay"`
    Timeout         time.Duration     `json:"timeout"`
    FallbackEnabled bool              `json:"fallback_enabled"`
    Metadata        map[string]string `json:"metadata"`
}
```

### Failover Configuration

```go
type FailoverConfig struct {
    Name                string                    `json:"name"`
    Strategy            FailoverStrategy          `json:"strategy"`
    HealthCheckInterval time.Duration             `json:"health_check_interval"`
    HealthCheckTimeout  time.Duration             `json:"health_check_timeout"`
    MaxFailures         int                       `json:"max_failures"`
    RecoveryTime        time.Duration             `json:"recovery_time"`
    RetryAttempts       int                       `json:"retry_attempts"`
    RetryDelay          time.Duration             `json:"retry_delay"`
    Timeout             time.Duration             `json:"timeout"`
    CircuitBreaker      *CircuitBreakerConfig     `json:"circuit_breaker,omitempty"`
    LoadBalancer        *LoadBalancerConfig       `json:"load_balancer,omitempty"`
    Metadata            map[string]string         `json:"metadata"`
}
```

## Failover Strategies

### Round Robin
Distributes requests evenly across all healthy endpoints in a circular fashion.

```go
config := &types.FailoverConfig{
    Strategy: types.StrategyRoundRobin,
}
```

### Weighted
Distributes requests based on endpoint weights. Higher weight means more traffic.

```go
config := &types.FailoverConfig{
    Strategy: types.StrategyWeighted,
}
```

### Least Connections
Routes requests to the endpoint with the fewest active connections.

```go
config := &types.FailoverConfig{
    Strategy: types.StrategyLeastConn,
}
```

### Random
Randomly selects an endpoint from the healthy pool.

```go
config := &types.FailoverConfig{
    Strategy: types.StrategyRandom,
}
```

### Health-Based
Prioritizes healthy endpoints over degraded ones.

```go
config := &types.FailoverConfig{
    Strategy: types.StrategyHealthBased,
}
```

## Providers

### Consul Provider

The Consul provider integrates with HashiCorp Consul for service discovery and health monitoring.

```go
consulConfig := &consul.Config{
    Address:    "localhost:8500",
    Token:      "your-consul-token",
    Datacenter: "dc1",
    Namespace:  "production",
    Timeout:    10 * time.Second,
}

consulProvider, err := consul.NewProvider(consulConfig, logger)
```

**Features:**
- Service registration and discovery
- Health checks via Consul
- Multi-datacenter support
- Namespace isolation
- ACL token authentication

### Kubernetes Provider

The Kubernetes provider integrates with Kubernetes for service discovery and load balancing.

```go
k8sConfig := &kubernetes.Config{
    KubeConfigPath: "/path/to/kubeconfig",
    Namespace:      "default",
    Context:        "production",
    Timeout:        10 * time.Second,
}

k8sProvider, err := kubernetes.NewProvider(k8sConfig, logger)
```

**Features:**
- Kubernetes service discovery
- Pod health monitoring
- Namespace isolation
- In-cluster and external configurations
- Service mesh integration

## Health Monitoring

### Health Check Configuration

```go
config := &types.FailoverConfig{
    HealthCheckInterval: 30 * time.Second,
    HealthCheckTimeout:  5 * time.Second,
    MaxFailures:         3,
    RecoveryTime:        60 * time.Second,
}
```

### Health Status Types

- `HealthHealthy`: Endpoint is fully operational
- `HealthDegraded`: Endpoint is partially functional
- `HealthUnhealthy`: Endpoint is not responding
- `HealthUnknown`: Health status cannot be determined

### Manual Health Checks

```go
// Check single endpoint
status, err := manager.HealthCheck(ctx, "endpoint-id")

// Check all endpoints
statuses, err := manager.HealthCheckAll(ctx)
```

## Circuit Breaker

The circuit breaker pattern prevents cascading failures by temporarily stopping requests to failing services.

```go
config := &types.FailoverConfig{
    CircuitBreaker: &types.CircuitBreakerConfig{
        Enabled:          true,
        FailureThreshold: 5,
        SuccessThreshold: 3,
        Timeout:          60 * time.Second,
        MaxRequests:      10,
        Interval:         10 * time.Second,
    },
}
```

## Monitoring and Observability

### Statistics

```go
stats, err := manager.GetStats(ctx)
for providerName, stat := range stats {
    log.Printf("Provider %s:", providerName)
    log.Printf("  Total requests: %d", stat.TotalRequests)
    log.Printf("  Successful requests: %d", stat.SuccessfulRequests)
    log.Printf("  Failed requests: %d", stat.FailedRequests)
    log.Printf("  Failovers: %d", stat.Failovers)
    log.Printf("  Recoveries: %d", stat.Recoveries)
    log.Printf("  Average response time: %v", stat.AverageResponseTime)
}
```

### Events

```go
events, err := manager.GetEvents(ctx, 10)
for providerName, eventList := range events {
    log.Printf("Provider %s events:", providerName)
    for _, event := range eventList {
        log.Printf("  %s: %s at %v", event.Type, event.Name, event.Timestamp)
    }
}
```

## Error Handling

### Common Error Codes

- `ErrCodeNoEndpoints`: No healthy endpoints available
- `ErrCodeAllEndpointsDown`: All endpoints are unhealthy
- `ErrCodeHealthCheckFailed`: Health check failed
- `ErrCodeTimeout`: Operation timed out
- `ErrCodeCircuitOpen`: Circuit breaker is open
- `ErrCodeRetryExhausted`: All retry attempts exhausted

### Error Handling Example

```go
result, err := manager.ExecuteWithFailover(ctx, config, fn)
if err != nil {
    log.Printf("Failover execution failed: %v", err)
    return
}

if result.Error != nil {
    switch result.Error.(*types.FailoverError).Code {
    case types.ErrCodeNoEndpoints:
        log.Println("No healthy endpoints available")
    case types.ErrCodeCircuitOpen:
        log.Println("Circuit breaker is open")
    default:
        log.Printf("Failover error: %v", result.Error)
    }
    return
}

log.Printf("Success: %v", result.Metadata["result"])
```

## Testing

### Unit Tests

```bash
cd test/unit
go test -v
```

### Integration Tests

```bash
cd test/integration
go test -v
```

### Mock Provider

For testing, you can use the mock provider:

```go
import "github.com/anasamu/microservices-library-go/failover/test/mocks"

mockProvider := mocks.NewMockProvider("test-provider")
mockProvider.AddEndpoint(&types.ServiceEndpoint{
    ID:     "test-endpoint",
    Health: types.HealthHealthy,
})
```

## Performance Considerations

- **Connection Pooling**: Providers maintain connection pools for optimal performance
- **Caching**: Endpoint information is cached to reduce discovery overhead
- **Concurrent Operations**: All operations are designed to be thread-safe
- **Memory Efficiency**: Minimal memory footprint with efficient data structures

## Best Practices

1. **Health Checks**: Configure appropriate health check intervals based on your service requirements
2. **Retry Logic**: Use exponential backoff for retry delays
3. **Circuit Breakers**: Enable circuit breakers for critical services
4. **Monitoring**: Monitor failover statistics and events
5. **Testing**: Use mock providers for unit testing
6. **Configuration**: Use environment-specific configurations
7. **Logging**: Enable appropriate log levels for debugging

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support, please open an issue on GitHub or contact the maintainers.
