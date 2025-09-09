# Rate Limiting Library

A comprehensive rate limiting library for Go microservices with support for multiple algorithms and storage backends.

## Features

- **Multiple Algorithms**: Token Bucket, Sliding Window, Fixed Window, and Leaky Bucket
- **Multiple Providers**: Redis, In-Memory, and extensible provider system
- **High Performance**: Optimized for high-throughput applications
- **Distributed Support**: Redis provider for distributed rate limiting
- **Flexible Configuration**: Customizable limits, windows, and algorithms
- **Batch Operations**: Efficient batch processing for multiple requests
- **Statistics & Monitoring**: Built-in metrics and monitoring capabilities
- **Thread-Safe**: Concurrent access support
- **Extensible**: Easy to add new providers and algorithms

## Installation

```bash
go get github.com/anasamu/microservices-library-go/ratelimit
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/anasamu/microservices-library-go/ratelimit/gateway"
    "github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory"
    "github.com/anasamu/microservices-library-go/ratelimit/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()

    // Create rate limit manager
    config := &gateway.ManagerConfig{
        DefaultProvider: "inmemory",
        RetryAttempts:   3,
        RetryDelay:      time.Second,
        Timeout:         30 * time.Second,
        FallbackEnabled: true,
    }

    manager := gateway.NewRateLimitManager(config, logger)

    // Register in-memory provider
    inmemoryProvider, err := inmemory.NewProvider(&inmemory.Config{
        CleanupInterval: time.Minute,
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := manager.RegisterProvider(inmemoryProvider); err != nil {
        log.Fatal(err)
    }

    // Define rate limit
    limit := &types.RateLimit{
        Limit:     10,
        Window:    time.Minute,
        Algorithm: types.AlgorithmTokenBucket,
    }

    // Check if request is allowed
    ctx := context.Background()
    result, err := manager.Allow(ctx, "user:123", limit)
    if err != nil {
        log.Fatal(err)
    }

    if result.Allowed {
        log.Printf("Request allowed. Remaining: %d", result.Remaining)
    } else {
        log.Printf("Request blocked. Retry after: %v", result.RetryAfter)
    }
}
```

## Rate Limiting Algorithms

### Token Bucket
- **Use Case**: Burst traffic handling
- **Behavior**: Allows bursts up to bucket capacity, refills at constant rate
- **Best For**: APIs with occasional traffic spikes

```go
limit := &types.RateLimit{
    Limit:     100,           // Bucket capacity
    Window:    time.Minute,   // Refill window
    Algorithm: types.AlgorithmTokenBucket,
}
```

### Sliding Window
- **Use Case**: Precise rate limiting
- **Behavior**: Tracks individual requests within a sliding time window
- **Best For**: Strict rate limiting requirements

```go
limit := &types.RateLimit{
    Limit:     60,            // Requests per window
    Window:    time.Minute,   // Window size
    Algorithm: types.AlgorithmSlidingWindow,
}
```

### Fixed Window
- **Use Case**: Simple rate limiting
- **Behavior**: Resets counter at fixed intervals
- **Best For**: Simple use cases with acceptable burst at window boundaries

```go
limit := &types.RateLimit{
    Limit:     1000,          // Requests per window
    Window:    time.Hour,     // Window size
    Algorithm: types.AlgorithmFixedWindow,
}
```

## Providers

### In-Memory Provider
- **Use Case**: Single-instance applications
- **Pros**: Fast, no external dependencies
- **Cons**: Not distributed, lost on restart

```go
provider, err := inmemory.NewProvider(&inmemory.Config{
    CleanupInterval: time.Minute,
    MaxBuckets:      10000,
})
```

### Redis Provider
- **Use Case**: Distributed applications
- **Pros**: Distributed, persistent, scalable
- **Cons**: Requires Redis infrastructure

```go
provider, err := redis.NewProvider(&redis.Config{
    Host:     "localhost",
    Port:     6379,
    Database: 0,
    Password: "",
    KeyPrefix: "ratelimit:",
})
```

## Advanced Usage

### Batch Operations

```go
requests := []*types.RateLimitRequest{
    {Key: "user:1", Limit: limit},
    {Key: "user:2", Limit: limit},
    {Key: "user:3", Limit: limit},
}

results, err := manager.AllowMultiple(ctx, requests)
```

### Custom Provider Selection

```go
// Use specific provider
result, err := manager.AllowWithProvider(ctx, "redis", "user:123", limit)

// Reset with specific provider
err := manager.ResetWithProvider(ctx, "redis", "user:123")
```

### Statistics and Monitoring

```go
// Get statistics from all providers
stats, err := manager.GetStats(ctx)
for provider, stat := range stats {
    log.Printf("Provider %s: Total=%d, Allowed=%d, Blocked=%d",
        provider, stat.TotalRequests, stat.AllowedRequests, stat.BlockedRequests)
}

// Get active keys
keys, err := manager.GetKeys(ctx, "user:*")
```

### Provider Information

```go
// List all providers
providers := manager.ListProviders()

// Get provider information
info := manager.GetProviderInfo()
for name, providerInfo := range info {
    log.Printf("Provider: %s, Connected: %v, Features: %v",
        name, providerInfo.IsConnected, providerInfo.SupportedFeatures)
}
```

## Configuration

### Manager Configuration

```go
config := &gateway.ManagerConfig{
    DefaultProvider: "redis",        // Default provider name
    RetryAttempts:   3,              // Retry attempts for failed operations
    RetryDelay:      time.Second,    // Delay between retries
    Timeout:         30 * time.Second, // Operation timeout
    FallbackEnabled: true,           // Enable fallback to other providers
    Metadata: map[string]string{     // Custom metadata
        "environment": "production",
    },
}
```

### Redis Provider Configuration

```go
config := &redis.Config{
    Host:         "localhost",       // Redis host
    Port:         6379,              // Redis port
    Database:     0,                 // Redis database number
    Password:     "",                // Redis password
    KeyPrefix:    "ratelimit:",      // Key prefix for rate limit keys
    PoolSize:     10,                // Connection pool size
    MinIdleConns: 5,                 // Minimum idle connections
}
```

### In-Memory Provider Configuration

```go
config := &inmemory.Config{
    CleanupInterval: time.Minute,    // Cleanup interval for expired entries
    MaxBuckets:      10000,          // Maximum number of buckets to maintain
}
```

## Error Handling

The library provides structured error handling with specific error codes:

```go
result, err := manager.Allow(ctx, "user:123", limit)
if err != nil {
    switch err.(type) {
    case *types.RateLimitError:
        rateLimitErr := err.(*types.RateLimitError)
        switch rateLimitErr.Code {
        case types.ErrCodeLimitExceeded:
            // Handle rate limit exceeded
        case types.ErrCodeConnection:
            // Handle connection error
        case types.ErrCodeTimeout:
            // Handle timeout
        }
    default:
        // Handle other errors
    }
}
```

## Testing

The library includes comprehensive test utilities:

```go
import "github.com/anasamu/microservices-library-go/ratelimit/test/mocks"

// Create mock provider
mockProvider := mocks.NewMockProvider()

// Set custom behavior
mockProvider.SetAllowFunc(func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
    return &types.RateLimitResult{
        Allowed:   true,
        Limit:     limit.Limit,
        Remaining: limit.Limit - 1,
        ResetTime: time.Now().Add(limit.Window),
        Key:       key,
    }, nil
})
```

## Performance Considerations

### In-Memory Provider
- **Memory Usage**: O(n) where n is the number of active rate limit keys
- **CPU Usage**: O(1) for most operations
- **Cleanup**: Automatic cleanup of expired entries

### Redis Provider
- **Network Latency**: Consider Redis network latency
- **Lua Scripts**: Uses atomic Lua scripts for consistency
- **Connection Pooling**: Configured connection pooling for efficiency

### Best Practices
1. **Key Design**: Use meaningful, hierarchical keys (e.g., `user:123`, `api:endpoint`)
2. **Window Size**: Choose appropriate window sizes based on your use case
3. **Provider Selection**: Use Redis for distributed systems, in-memory for single instances
4. **Monitoring**: Monitor rate limit statistics and adjust limits accordingly
5. **Error Handling**: Implement proper error handling and fallback strategies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Examples

See the `examples/` directory for complete working examples:

- [Basic Usage](gateway/example.go)
- [Multiple Providers](examples/multiple_providers.go)
- [Custom Algorithms](examples/custom_algorithms.go)
- [HTTP Middleware](examples/http_middleware.go)

## Support

For questions, issues, or contributions, please visit the [GitHub repository](https://github.com/anasamu/microservices-library-go).
