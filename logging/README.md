# Logging Library

A comprehensive logging library for Go microservices with support for multiple backends, structured logging, and advanced features.

## Features

- **Multiple Providers**: Console, File, Elasticsearch, and more
- **Structured Logging**: JSON and text formats with custom fields
- **Context Support**: Automatic context propagation (trace IDs, user IDs, etc.)
- **Batch Operations**: Efficient batch logging for high-throughput scenarios
- **Search & Query**: Advanced log search capabilities (Elasticsearch)
- **Retry Logic**: Built-in retry mechanisms for reliability
- **Provider Management**: Easy switching between different logging backends
- **Performance**: Optimized for high-performance logging

## Architecture

The logging library follows a provider-based architecture:

```
┌─────────────────┐
│   Application   │
└─────────┬───────┘
          │
┌─────────▼───────┐
│ Logging Manager │
└─────────┬───────┘
          │
    ┌─────┴─────┐
    │           │
┌───▼───┐   ┌───▼───┐   ┌─────────────┐
│Console│   │ File  │   │Elasticsearch│
└───────┘   └───────┘   └─────────────┘
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/anasamu/microservices-library-go/logging"
    "github.com/anasamu/microservices-library-go/logging/providers/console"
    "github.com/anasamu/microservices-library-go/logging/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create manager
    logger := logrus.New()
    manager := gateway.NewLoggingManager(nil, logger)

    // Create and register console provider
    config := &types.LoggingConfig{
        Level:   types.LevelInfo,
        Format:  types.FormatJSON,
        Output:  types.OutputStdout,
        Service: "my-service",
        Version: "1.0.0",
    }
    
    provider := console.NewConsoleProvider(config)
    manager.RegisterProvider(provider)

    // Log messages
    ctx := context.Background()
    manager.Info(ctx, "Application started")
    manager.Error(ctx, "Something went wrong", map[string]interface{}{
        "error_code": "ERR_001",
        "details":    "Database connection failed",
    })
}
```

### Multiple Providers

```go
// Register multiple providers
consoleProvider := console.NewConsoleProvider(consoleConfig)
fileProvider := file.NewFileProvider(fileConfig)
elasticProvider := elasticsearch.NewElasticsearchProvider(elasticConfig)

manager.RegisterProvider(consoleProvider)
manager.RegisterProvider(fileProvider)
manager.RegisterProvider(elasticProvider)

// Log to specific provider
manager.LogWithProvider(ctx, "file", types.LevelInfo, "This goes to file")
manager.LogWithProvider(ctx, "elasticsearch", types.LevelError, "This goes to Elasticsearch")
```

## Providers

### Console Provider

Logs to stdout/stderr with configurable formatting.

```go
config := &types.LoggingConfig{
    Level:   types.LevelInfo,
    Format:  types.FormatJSON, // or FormatText
    Output:  types.OutputStdout, // or OutputStderr
    Service: "my-service",
    Version: "1.0.0",
}

provider := console.NewConsoleProvider(config)
```

**Supported Features:**
- Basic logging
- Structured logging
- Formatted logging
- Context logging

### File Provider

Logs to files with rotation and retention support.

```go
config := &types.LoggingConfig{
    Level:   types.LevelDebug,
    Format:  types.FormatJSON,
    Output:  types.OutputFile,
    Service: "my-service",
    Version: "1.0.0",
    Metadata: map[string]interface{}{
        "file_path": "/var/log/myapp.log",
    },
}

provider := file.NewFileProvider(config)
provider.Connect(ctx) // Must connect before logging
```

**Supported Features:**
- Basic logging
- Structured logging
- Formatted logging
- Context logging
- Batch logging
- Retention

### Elasticsearch Provider

Logs to Elasticsearch with search capabilities.

```go
config := &types.LoggingConfig{
    Level:   types.LevelInfo,
    Format:  types.FormatJSON,
    Service: "my-service",
    Version: "1.0.0",
    Index:   "application-logs",
    Metadata: map[string]interface{}{
        "elastic_url": "http://localhost:9200",
    },
}

provider, err := elasticsearch.NewElasticsearchProvider(config)
provider.Connect(ctx) // Must connect before logging
```

**Supported Features:**
- Basic logging
- Structured logging
- Formatted logging
- Context logging
- Batch logging
- Search & filtering
- Aggregation
- Retention
- Compression
- Real-time

## Advanced Features

### Context Logging

Automatic context propagation for distributed tracing:

```go
ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
ctx = context.WithValue(ctx, "user_id", "user-456")
ctx = context.WithValue(ctx, "request_id", "req-789")

manager.Info(ctx, "Request processed") // Automatically includes context fields
```

### Structured Logging

Rich structured logging with custom fields:

```go
manager.Info(ctx, "User action", map[string]interface{}{
    "user_id":    "123",
    "action":     "login",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "duration":   150,
})
```

### Batch Logging

Efficient batch operations for high-throughput scenarios:

```go
entries := []types.LogEntry{
    {
        Timestamp: time.Now(),
        Level:     types.LevelInfo,
        Message:   "Batch entry 1",
        Service:   "my-service",
        Fields:    map[string]interface{}{"batch_id": "batch-001"},
    },
    // ... more entries
}

manager.LogBatch(ctx, entries)
```

### Search & Query (Elasticsearch)

Advanced log search capabilities:

```go
query := types.LogQuery{
    Levels:   []types.LogLevel{types.LevelWarn, types.LevelError},
    Services: []string{"my-service"},
    StartTime: &startTime,
    EndTime:   &endTime,
    Message:   "error",
    Limit:     100,
}

entries, err := manager.Search(ctx, query)
```

### Provider Management

Easy provider management and switching:

```go
// List all providers
providers := manager.ListProviders()

// Get provider information
info := manager.GetProviderInfo()

// Get specific provider
provider, err := manager.GetProvider("elasticsearch")

// Get default provider
defaultProvider, err := manager.GetDefaultProvider()
```

## Configuration

### LoggingConfig

```go
type LoggingConfig struct {
    Level       LogLevel               `json:"level"`        // trace, debug, info, warn, error, fatal, panic
    Format      LogFormat              `json:"format"`       // json, text, xml
    Output      LogOutput              `json:"output"`       // stdout, stderr, file, syslog
    Service     string                 `json:"service"`      // Service name
    Version     string                 `json:"version"`      // Service version
    Index       string                 `json:"index"`        // Index name (for Elasticsearch)
    Retention   time.Duration          `json:"retention"`    // Log retention period
    Compression bool                   `json:"compression"`  // Enable compression
    Encryption  bool                   `json:"encryption"`   // Enable encryption
    Metadata    map[string]interface{} `json:"metadata"`     // Provider-specific config
}
```

### ManagerConfig

```go
type ManagerConfig struct {
    DefaultProvider string            `json:"default_provider"` // Default provider name
    RetryAttempts   int               `json:"retry_attempts"`   // Number of retry attempts
    RetryDelay      time.Duration     `json:"retry_delay"`      // Delay between retries
    Timeout         time.Duration     `json:"timeout"`          // Operation timeout
    FallbackEnabled bool              `json:"fallback_enabled"` // Enable fallback
    Metadata        map[string]string `json:"metadata"`         // Manager metadata
}
```

## Error Handling

The library provides comprehensive error handling:

```go
// Check connection status
if !provider.IsConnected() {
    log.Printf("Provider not connected")
}

// Handle connection errors
if err := provider.Connect(ctx); err != nil {
    log.Printf("Failed to connect: %v", err)
}

// Handle logging errors
if err := manager.Error(ctx, "Error message"); err != nil {
    log.Printf("Failed to log: %v", err)
}
```

## Performance Considerations

- **Batch Operations**: Use batch logging for high-throughput scenarios
- **Async Logging**: Consider implementing async logging for better performance
- **Connection Pooling**: Elasticsearch provider uses connection pooling
- **Retry Logic**: Built-in retry mechanisms prevent data loss
- **Compression**: Enable compression for large log volumes

## Best Practices

1. **Use Structured Logging**: Always use structured logging with meaningful fields
2. **Context Propagation**: Leverage context for distributed tracing
3. **Error Handling**: Always handle logging errors appropriately
4. **Provider Selection**: Choose the right provider for your use case
5. **Configuration**: Use appropriate log levels and formats
6. **Monitoring**: Monitor logging performance and errors
7. **Retention**: Configure appropriate log retention policies

## Examples

See the `examples/` directory for comprehensive usage examples:

- Basic logging examples
- Multiple provider setup
- Advanced features demonstration
- Performance testing
- Error handling patterns

## Dependencies

- `github.com/sirupsen/logrus` - Logging framework
- `github.com/elastic/go-elasticsearch/v8` - Elasticsearch client (for Elasticsearch provider)

## License

This library is part of the microservices-library-go project.
