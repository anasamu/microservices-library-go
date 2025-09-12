# Cache Module

A comprehensive caching library for Go microservices with support for multiple cache backends and advanced features.

## Features

- **Multiple Providers**: Redis, Memcached, and in-memory caching
- **Unified Interface**: Single API for all cache operations
- **Advanced Features**: Tags, TTL management, batch operations, statistics
- **High Availability**: Retry logic, fallback support, connection management
- **Monitoring**: Built-in statistics and health checks
- **Thread-Safe**: Concurrent access support

## Supported Providers

### Redis
- Full feature support including tags, TTL, batch operations
- Connection pooling and retry logic
- Pub/Sub and Lua scripts support
- Persistence and clustering support

### Memcached
- Basic and advanced operations
- Connection pooling
- Batch operations
- Statistics support

### Memory
- In-memory caching with configurable limits
- Automatic cleanup of expired items
- Tag support with efficient invalidation
- Thread-safe operations

## Installation

```bash
go get github.com/anasamu/microservices-library-go/cache
```

## Quick Start

```go
package main

import (
    "context"
    "time"
    
    "github.com/anasamu/microservices-library-go/cache"
    "github.com/anasamu/microservices-library-go/cache/providers/memory"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create cache manager
    manager := gateway.NewCacheManager(nil, logger)
    
    // Register memory provider
    memoryProvider := memory.NewMemoryProvider(nil, logger)
    manager.RegisterProvider(memoryProvider)
    
    // Connect
    ctx := context.Background()
    memoryProvider.Connect(ctx)
    
    // Use cache
    manager.Set(ctx, "key", "value", 10*time.Minute)
    
    var value string
    manager.Get(ctx, "key", &value)
    
    // Cleanup
    manager.Close()
}
```

## Configuration

### Manager Configuration

```go
config := &gateway.ManagerConfig{
    DefaultProvider: "redis",     // Default provider name
    RetryAttempts:   3,           // Number of retry attempts
    RetryDelay:      time.Second, // Delay between retries
    Timeout:         30 * time.Second, // Operation timeout
    FallbackEnabled: true,        // Enable fallback to other providers
}
```

### Redis Configuration

```go
config := &redis.RedisConfig{
    Host:         "localhost",
    Port:         6379,
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    KeyPrefix:    "app",
    Namespace:    "production",
}
```

### Memory Configuration

```go
config := &memory.MemoryConfig{
    MaxSize:         10000,        // Maximum number of items
    DefaultTTL:      5 * time.Minute, // Default TTL
    CleanupInterval: 1 * time.Minute, // Cleanup interval
    KeyPrefix:       "app",
    Namespace:       "production",
}
```

### Memcached Configuration

```go
config := &memcache.MemcacheConfig{
    Servers:       []string{"localhost:11211"},
    MaxIdleConns:  2,
    Timeout:       100 * time.Millisecond,
    KeyPrefix:     "app",
    Namespace:     "production",
    MaxKeySize:    250,
    MaxValueSize:  1024 * 1024, // 1MB
}
```

## API Reference

### Basic Operations

```go
// Set a value
err := manager.Set(ctx, "key", value, 10*time.Minute)

// Get a value
var result MyType
err := manager.Get(ctx, "key", &result)

// Delete a value
err := manager.Delete(ctx, "key")

// Check if key exists
exists, err := manager.Exists(ctx, "key")
```

### Advanced Operations

```go
// Set with tags
err := manager.SetWithTags(ctx, "key", value, 10*time.Minute, []string{"tag1", "tag2"})

// Invalidate by tag
err := manager.InvalidateByTag(ctx, "tag1")

// Batch operations
items := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
}
err := manager.SetMultiple(ctx, items, 10*time.Minute)

keys := []string{"key1", "key2"}
results, err := manager.GetMultiple(ctx, keys)

err := manager.DeleteMultiple(ctx, keys)
```

### TTL Management

```go
// Get TTL
ttl, err := manager.GetTTL(ctx, "key")

// Set TTL
err := manager.SetTTL(ctx, "key", 5*time.Minute)
```

### Key Management

```go
// Get keys matching pattern
keys, err := manager.GetKeys(ctx, "user:*")
```

### Statistics

```go
// Get cache statistics
stats, err := manager.GetStats(ctx)
for provider, stat := range stats {
    fmt.Printf("Provider: %s, Hits: %d, Misses: %d\n", 
        provider, stat.Hits, stat.Misses)
}
```

### Provider Management

```go
// List registered providers
providers := manager.ListProviders()

// Get provider information
info := manager.GetProviderInfo()
for name, providerInfo := range info {
    fmt.Printf("Provider: %s, Connected: %v\n", 
        name, providerInfo.IsConnected)
}

// Use specific provider
err := manager.SetWithProvider(ctx, "redis", "key", value, ttl)
```

## Error Handling

The cache module uses structured error types:

```go
type CacheError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Key     string `json:"key,omitempty"`
}
```

Common error codes:
- `NOT_FOUND`: Key not found
- `INVALID_KEY`: Invalid key format
- `INVALID_VALUE`: Invalid value format
- `EXPIRED`: Key has expired
- `CONNECTION_ERROR`: Connection issue
- `TIMEOUT`: Operation timeout
- `QUOTA_EXCEEDED`: Cache quota exceeded
- `SERIALIZATION_ERROR`: Serialization failed
- `DESERIALIZATION_ERROR`: Deserialization failed

## Examples

See the `examples/` directory for complete usage examples including:
- Basic cache operations
- Tag-based invalidation
- Batch operations
- Provider management
- Statistics and monitoring

## Performance Considerations

- Use connection pooling for Redis and Memcached
- Set appropriate TTL values to avoid memory bloat
- Use batch operations for multiple items
- Monitor cache hit rates and adjust accordingly
- Consider cache warming strategies for critical data

## Thread Safety

All cache providers are thread-safe and can be used concurrently from multiple goroutines.

## License

This module is part of the microservices-library-go project.
