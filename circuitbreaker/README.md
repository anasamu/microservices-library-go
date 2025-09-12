# Circuit Breaker Library

A comprehensive circuit breaker implementation for Go microservices, providing fault tolerance and resilience patterns.

## Features

- **Multiple Providers**: Support for different circuit breaker implementations
- **Fallback Support**: Automatic fallback execution when primary service fails
- **Retry Logic**: Configurable retry mechanisms with exponential backoff
- **State Management**: Real-time circuit breaker state monitoring
- **Statistics**: Detailed metrics and monitoring capabilities
- **Configuration**: Flexible configuration options for different use cases
- **Thread-Safe**: Concurrent execution support
- **Context Support**: Full context.Context integration

## Architecture

```
circuitbreaker/
├── gateway/                    # Core circuit breaker gateway
│   ├── manager.go             # Circuit breaker manager
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Circuit breaker provider implementations
│   ├── gobreaker/             # Gobreaker-based circuit breaker
│   │   ├── provider.go        # Gobreaker implementation
│   │   └── go.mod             # Gobreaker dependencies
│   └── custom/                # Custom circuit breaker
│       ├── provider.go        # Custom implementation
│       └── go.mod             # Custom dependencies
├── test/                      # Test files
│   ├── integration/           # Integration tests
│   ├── unit/                  # Unit tests
│   └── mocks/                 # Mock providers
├── types/                     # Common types and interfaces
│   ├── types.go               # Type definitions
│   └── go.mod                 # Types dependencies
├── go.mod                     # Main module dependencies
└── README.md                  # This documentation
```

## Quick Start

### Installation

```bash
go get github.com/anasamu/microservices-library-go/circuitbreaker
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/anasamu/microservices-library-go/circuitbreaker"
    "github.com/anasamu/microservices-library-go/circuitbreaker/providers/custom"
    "github.com/anasamu/microservices-library-go/circuitbreaker/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()

    // Create manager configuration
    config := &gateway.ManagerConfig{
        DefaultProvider: "custom",
        RetryAttempts:   3,
        RetryDelay:      time.Second,
        Timeout:         30 * time.Second,
        FallbackEnabled: true,
    }

    // Create circuit breaker manager
    manager := gateway.NewCircuitBreakerManager(config, logger)

    // Register custom provider
    customProvider := custom.NewCustomProvider(nil, logger)
    err := manager.RegisterProvider(customProvider)
    if err != nil {
        logger.Fatal("Failed to register provider:", err)
    }

    ctx := context.Background()

    // Execute a function through the circuit breaker
    result, err := manager.Execute(ctx, "my-service", func() (interface{}, error) {
        // Your service call here
        return "success", nil
    })

    if err != nil {
        logger.WithError(err).Error("Execution failed")
    } else {
        logger.WithField("result", result.Result).Info("Execution successful")
    }

    // Clean up
    manager.Close()
}
```

### With Fallback

```go
result, err := manager.ExecuteWithFallback(ctx, "my-service",
    func() (interface{}, error) {
        // Primary service call
        return callPrimaryService()
    },
    func() (interface{}, error) {
        // Fallback service call
        return callFallbackService()
    },
)
```

### Configuration

```go
config := &types.CircuitBreakerConfig{
    Name:                "my-service",
    MaxRequests:         10,
    Interval:            10 * time.Second,
    Timeout:             60 * time.Second,
    MaxConsecutiveFails: 5,
    FailureThreshold:    0.5,
    SuccessThreshold:    3,
    FallbackEnabled:     true,
    RetryEnabled:        true,
    RetryAttempts:       3,
    RetryDelay:          time.Second,
    ReadyToTrip: func(counts types.Counts) bool {
        return counts.ConsecutiveFailures >= 5
    },
    IsSuccessful: func(err error) bool {
        return err == nil
    },
    OnStateChange: func(name string, from types.CircuitState, to types.CircuitState) {
        logger.WithFields(logrus.Fields{
            "name": name,
            "from": from,
            "to":   to,
        }).Warn("Circuit breaker state changed")
    },
}

err := manager.Configure(ctx, "my-service", config)
```

## Providers

### Custom Provider

A custom implementation with full control over circuit breaker logic:

```go
customProvider := custom.NewCustomProvider(&custom.CustomConfig{
    DefaultMaxRequests: 10,
    DefaultInterval:    10 * time.Second,
    DefaultTimeout:     60 * time.Second,
    EnableMetrics:      true,
    EnableLogging:      true,
}, logger)
```

**Features:**
- Custom state transition logic
- Configurable failure thresholds
- Built-in metrics collection
- Health check support
- Rate limiting capabilities

### Gobreaker Provider

Based on the popular [sony/gobreaker](https://github.com/sony/gobreaker) library:

```go
gobreakerProvider := gobreaker.NewGobreakerProvider(&gobreaker.GobreakerConfig{
    DefaultMaxRequests: 10,
    DefaultInterval:    10 * time.Second,
    DefaultTimeout:     60 * time.Second,
}, logger)
```

**Features:**
- Battle-tested implementation
- High performance
- Minimal overhead
- Standard circuit breaker patterns

## Circuit Breaker States

- **Closed**: Normal operation, requests pass through
- **Open**: Circuit is open, requests fail fast
- **Half-Open**: Testing if service is back, limited requests allowed

## Monitoring and Statistics

```go
// Get circuit breaker state
state, err := manager.GetState(ctx, "my-service")

// Get detailed statistics
stats, err := manager.GetStats(ctx, "my-service")
fmt.Printf("Requests: %d, Successes: %d, Failures: %d, State: %s\n",
    stats.Requests, stats.Successes, stats.Failures, stats.State)

// Get all states from all providers
allStates, err := manager.GetAllStates(ctx)

// Get all statistics from all providers
allStats, err := manager.GetAllStats(ctx)
```

## Error Handling

The library provides specific error types for different scenarios:

```go
if cbErr, ok := err.(*types.CircuitBreakerError); ok {
    switch cbErr.Code {
    case types.ErrCodeCircuitOpen:
        // Circuit breaker is open
    case types.ErrCodeTimeout:
        // Operation timed out
    case types.ErrCodeMaxRequests:
        // Maximum requests exceeded
    case types.ErrCodeFallbackFailed:
        // Fallback function failed
    }
}
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

For testing, use the mock provider:

```go
import "github.com/anasamu/microservices-library-go/circuitbreaker/test/mocks"

mockProvider := &mocks.MockCircuitBreakerProvider{}
mockProvider.On("GetName").Return("mock")
mockProvider.On("Execute", mock.Anything, "test", mock.Anything).Return(
    mocks.NewMockExecutionResult("success", nil, types.StateClosed), nil)
```

## Best Practices

1. **Configure Appropriately**: Set reasonable thresholds based on your service characteristics
2. **Monitor Metrics**: Regularly check circuit breaker statistics and adjust configuration
3. **Use Fallbacks**: Always provide fallback mechanisms for critical services
4. **Test Failure Scenarios**: Test your application behavior when circuit breakers are open
5. **Log State Changes**: Monitor circuit breaker state changes for debugging
6. **Context Timeouts**: Use context timeouts to prevent hanging operations

## Configuration Examples

### High Availability Service

```go
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
}
```

### Resilient Service

```go
config := &types.CircuitBreakerConfig{
    Name:                "resilient-service",
    MaxRequests:         50,
    Interval:            30 * time.Second,
    Timeout:             120 * time.Second,
    MaxConsecutiveFails: 10,
    FailureThreshold:    0.7, // 70% failure rate
    SuccessThreshold:    3,
    FallbackEnabled:     true,
    RetryEnabled:        true,
    RetryAttempts:       5,
    RetryDelay:          2 * time.Second,
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Dependencies

- [logrus](https://github.com/sirupsen/logrus) - Structured logging
- [sony/gobreaker](https://github.com/sony/gobreaker) - Circuit breaker implementation
- [testify](https://github.com/stretchr/testify) - Testing framework

## Support

For questions, issues, or contributions, please visit the [GitHub repository](https://github.com/anasamu/microservices-library-go/circuitbreaker).
