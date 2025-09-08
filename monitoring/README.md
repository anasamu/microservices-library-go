# Monitoring Library

A comprehensive monitoring library for microservices that provides unified interfaces for metrics, logging, tracing, and alerting across multiple monitoring backends.

## Features

- **Unified Interface**: Single API for multiple monitoring backends
- **Multiple Providers**: Support for Prometheus, Jaeger, Elasticsearch, and more
- **Comprehensive Monitoring**: Metrics, logs, traces, alerts, and health checks
- **High Performance**: Batch operations, retry logic, and connection pooling
- **Flexible Configuration**: Environment variables and programmatic configuration
- **Production Ready**: Error handling, logging, and monitoring capabilities

## Supported Providers

### Metrics
- **Prometheus**: Metrics collection and querying
- **Elasticsearch**: Metrics indexing and search

### Logging
- **Elasticsearch**: Log aggregation and search
- **File**: Local file logging

### Tracing
- **Jaeger**: Distributed tracing
- **Elasticsearch**: Trace indexing and search

### Alerting
- **Prometheus Alertmanager**: Alert management
- **Elasticsearch**: Alert indexing and search

## Installation

```bash
go get github.com/anasamu/microservices-library-go/monitoring
```

## Quick Start

```go
package main

import (
    "context"
    "time"
    
    "github.com/anasamu/microservices-library-go/monitoring/gateway"
    "github.com/anasamu/microservices-library-go/monitoring/providers/prometheus"
    "github.com/anasamu/microservices-library-go/monitoring/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create monitoring manager
    config := &gateway.ManagerConfig{
        DefaultProvider: "prometheus",
        RetryAttempts:   3,
        RetryDelay:      5 * time.Second,
        Timeout:         30 * time.Second,
    }
    
    manager := gateway.NewMonitoringManager(config, logger)
    
    // Create and register Prometheus provider
    prometheusProvider := prometheus.NewPrometheusProvider(nil, logger)
    manager.RegisterProvider(prometheusProvider)
    
    // Connect to provider
    ctx := context.Background()
    manager.Connect(ctx, "prometheus")
    
    // Submit metrics
    metrics := []types.Metric{
        *prometheus.CreateCounterMetric("http_requests_total", 1, map[string]string{
            "method": "GET",
            "path":   "/api/users",
            "status": "200",
        }),
    }
    
    metricRequest := &types.MetricRequest{
        Metrics: metrics,
    }
    
    manager.SubmitMetrics(ctx, "prometheus", metricRequest)
    
    // Clean up
    manager.Close()
}
```

## Configuration

### Environment Variables

The library supports configuration through environment variables:

#### Prometheus
- `PROMETHEUS_HOST`: Prometheus server host (default: localhost)
- `PROMETHEUS_PORT`: Prometheus server port (default: 9090)
- `PROMETHEUS_PROTOCOL`: Protocol (http/https) (default: http)
- `PROMETHEUS_PATH`: API path (default: /api/v1)
- `PROMETHEUS_TIMEOUT`: Request timeout (default: 30s)
- `PROMETHEUS_RETRY_ATTEMPTS`: Retry attempts (default: 3)
- `PROMETHEUS_RETRY_DELAY`: Retry delay (default: 5s)
- `PROMETHEUS_BATCH_SIZE`: Batch size (default: 1000)
- `PROMETHEUS_FLUSH_INTERVAL`: Flush interval (default: 10s)
- `PROMETHEUS_USERNAME`: Username for authentication
- `PROMETHEUS_PASSWORD`: Password for authentication
- `PROMETHEUS_TOKEN`: Token for authentication
- `PROMETHEUS_NAMESPACE`: Namespace for metrics
- `PROMETHEUS_ENVIRONMENT`: Environment name

#### Jaeger
- `JAEGER_HOST`: Jaeger server host (default: localhost)
- `JAEGER_PORT`: Jaeger server port (default: 14268)
- `JAEGER_PROTOCOL`: Protocol (http/https) (default: http)
- `JAEGER_PATH`: API path (default: /api/traces)
- `JAEGER_TIMEOUT`: Request timeout (default: 30s)
- `JAEGER_RETRY_ATTEMPTS`: Retry attempts (default: 3)
- `JAEGER_RETRY_DELAY`: Retry delay (default: 5s)
- `JAEGER_BATCH_SIZE`: Batch size (default: 1000)
- `JAEGER_FLUSH_INTERVAL`: Flush interval (default: 10s)
- `JAEGER_USERNAME`: Username for authentication
- `JAEGER_PASSWORD`: Password for authentication
- `JAEGER_TOKEN`: Token for authentication
- `JAEGER_SERVICE_NAME`: Service name
- `JAEGER_ENVIRONMENT`: Environment name
- `JAEGER_SAMPLING_RATE`: Sampling rate (default: 1.0)

#### Elasticsearch
- `ELASTICSEARCH_HOST`: Elasticsearch server host (default: localhost)
- `ELASTICSEARCH_PORT`: Elasticsearch server port (default: 9200)
- `ELASTICSEARCH_PROTOCOL`: Protocol (http/https) (default: http)
- `ELASTICSEARCH_PATH`: API path (default: "")
- `ELASTICSEARCH_TIMEOUT`: Request timeout (default: 30s)
- `ELASTICSEARCH_RETRY_ATTEMPTS`: Retry attempts (default: 3)
- `ELASTICSEARCH_RETRY_DELAY`: Retry delay (default: 5s)
- `ELASTICSEARCH_BATCH_SIZE`: Batch size (default: 1000)
- `ELASTICSEARCH_FLUSH_INTERVAL`: Flush interval (default: 10s)
- `ELASTICSEARCH_USERNAME`: Username for authentication
- `ELASTICSEARCH_PASSWORD`: Password for authentication
- `ELASTICSEARCH_TOKEN`: Token for authentication
- `ELASTICSEARCH_INDEX_PREFIX`: Index prefix
- `ELASTICSEARCH_ENVIRONMENT`: Environment name
- `ELASTICSEARCH_MAX_RETRIES`: Maximum retries (default: 3)

### Programmatic Configuration

```go
// Prometheus configuration
prometheusConfig := &prometheus.PrometheusConfig{
    Host:          "localhost",
    Port:          9090,
    Protocol:      "http",
    Path:          "/api/v1",
    Timeout:       30 * time.Second,
    RetryAttempts: 3,
    RetryDelay:    5 * time.Second,
    BatchSize:     1000,
    FlushInterval: 10 * time.Second,
    Namespace:     "myapp",
    Environment:   "production",
}

prometheusProvider := prometheus.NewPrometheusProvider(prometheusConfig, logger)
```

## Usage Examples

### Metrics

```go
// Create counter metric
counter := prometheus.CreateCounterMetric("http_requests_total", 1, map[string]string{
    "method": "GET",
    "path":   "/api/users",
    "status": "200",
})

// Create gauge metric
gauge := prometheus.CreateGaugeMetric("memory_usage_bytes", 1024*1024*100, map[string]string{
    "instance": "server-1",
})

// Create histogram metric
histogram := prometheus.CreateHistogramMetric("request_duration_seconds", 0.5, map[string]string{
    "method": "POST",
    "path":   "/api/orders",
})

// Submit metrics
metricRequest := &types.MetricRequest{
    Metrics: []types.Metric{*counter, *gauge, *histogram},
}

manager.SubmitMetrics(ctx, "prometheus", metricRequest)
```

### Tracing

```go
// Create trace
trace := jaeger.CreateTrace("user-service", "get-user")

// Create spans
span1 := jaeger.CreateSpan(trace.TraceID, "", "database-query")
span2 := jaeger.CreateSpan(trace.TraceID, span1.SpanID, "cache-lookup")

// Add tags and logs
span1.AddTag("db.type", "postgresql")
span1.AddTag("db.table", "users")
span1.AddLog(types.LogLevelInfo, "Executing user query", map[string]interface{}{
    "query": "SELECT * FROM users WHERE id = ?",
    "params": []interface{}{123},
})

// Finish spans
span1.FinishSpan("ok")
span2.FinishSpan("ok")

// Add spans to trace
trace.AddSpan(span1)
trace.AddSpan(span2)

// Finish trace
trace.FinishTrace("ok")

// Submit trace
traceRequest := &types.TraceRequest{
    Traces: []types.Trace{*trace},
}

manager.SubmitTraces(ctx, "jaeger", traceRequest)
```

### Logging

```go
// Create log entry
logEntry := elasticsearch.CreateLogEntry(types.LogLevelInfo, "User login successful", "auth-service", "auth-service")

// Add fields
logEntry.AddField("user_id", "12345")
logEntry.AddField("ip_address", "192.168.1.100")
logEntry.SetTraceInfo(traceID, spanID)

// Submit logs
logRequest := &types.LogRequest{
    Logs: []types.LogEntry{*logEntry},
}

manager.SubmitLogs(ctx, "elasticsearch", logRequest)
```

### Alerting

```go
// Create alert
alert := prometheus.CreateAlert("HighMemoryUsage", "Memory usage is above 90%", types.AlertSeverityHigh, "monitoring", "user-service")

// Add labels and annotations
alert.Labels["instance"] = "server-1"
alert.Labels["severity"] = "high"
alert.Annotations["summary"] = "High memory usage detected"
alert.Annotations["description"] = "Memory usage has exceeded 90% for more than 5 minutes"

// Submit alert
alertRequest := &types.AlertRequest{
    Alerts: []types.Alert{*alert},
}

manager.SubmitAlerts(ctx, "prometheus", alertRequest)
```

### Health Checks

```go
// Health check for specific provider
healthRequest := &types.HealthCheckRequest{
    Service: "prometheus",
    Timeout: 10 * time.Second,
}

healthResponse, err := manager.HealthCheck(ctx, "prometheus", healthRequest)
if err != nil {
    log.Printf("Health check failed: %v", err)
} else {
    fmt.Printf("Health status: %s\n", healthResponse.Status)
    for _, check := range healthResponse.Checks {
        fmt.Printf("  - %s: %s (%s)\n", check.Name, check.Status, check.Message)
    }
}

// Health check for all providers
allHealthResponses := manager.HealthCheckAll(ctx)
for provider, response := range allHealthResponses {
    fmt.Printf("%s health status: %s\n", provider, response.Status)
}
```

### Querying Data

```go
// Query metrics
metricQueryRequest := &types.QueryRequest{
    Query: "http_requests_total",
    Start: time.Now().Add(-1 * time.Hour),
    End:   time.Now(),
    Step:  5 * time.Minute,
}

metricResponse, err := manager.QueryMetrics(ctx, "prometheus", metricQueryRequest)
if err != nil {
    log.Printf("Failed to query metrics: %v", err)
} else {
    fmt.Printf("Metrics query result: %+v\n", metricResponse.Data)
}

// Query traces
traceQueryRequest := &types.QueryRequest{
    Query: "service:user-service",
    Start: time.Now().Add(-1 * time.Hour),
    End:   time.Now(),
}

traceResponse, err := manager.QueryTraces(ctx, "jaeger", traceQueryRequest)
if err != nil {
    log.Printf("Failed to query traces: %v", err)
} else {
    fmt.Printf("Traces query result: %+v\n", traceResponse.Data)
}

// Query logs
logQueryRequest := &types.QueryRequest{
    Query: "level:error",
    Start: time.Now().Add(-1 * time.Hour),
    End:   time.Now(),
}

logResponse, err := manager.QueryLogs(ctx, "elasticsearch", logQueryRequest)
if err != nil {
    log.Printf("Failed to query logs: %v", err)
} else {
    fmt.Printf("Logs query result: %+v\n", logResponse.Data)
}
```

## Architecture

The monitoring library follows a provider pattern with the following components:

1. **Gateway Manager**: Central manager that coordinates multiple providers
2. **Types Package**: Common types and interfaces for all providers
3. **Providers**: Specific implementations for different monitoring backends
4. **Configuration**: Flexible configuration system with environment variable support

### Provider Interface

All providers implement the `MonitoringProvider` interface:

```go
type MonitoringProvider interface {
    // Provider information
    GetName() string
    GetSupportedFeatures() []types.MonitoringFeature
    GetConnectionInfo() *types.ConnectionInfo

    // Connection management
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Ping(ctx context.Context) error
    IsConnected() bool

    // Metrics operations
    SubmitMetrics(ctx context.Context, request *types.MetricRequest) error
    QueryMetrics(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

    // Logging operations
    SubmitLogs(ctx context.Context, request *types.LogRequest) error
    QueryLogs(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

    // Tracing operations
    SubmitTraces(ctx context.Context, request *types.TraceRequest) error
    QueryTraces(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

    // Alerting operations
    SubmitAlerts(ctx context.Context, request *types.AlertRequest) error
    QueryAlerts(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

    // Health check operations
    HealthCheck(ctx context.Context, request *types.HealthCheckRequest) (*types.HealthCheckResponse, error)

    // Statistics and monitoring
    GetStats(ctx context.Context) (*types.MonitoringStats, error)

    // Configuration
    Configure(config map[string]interface{}) error
    IsConfigured() bool
    Close() error
}
```

## Error Handling

The library provides comprehensive error handling with custom error types:

```go
type MonitoringError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Source  string `json:"source,omitempty"`
}
```

Common error codes:
- `CONNECTION_ERROR`: Connection to monitoring backend failed
- `TIMEOUT`: Request timeout
- `INVALID_METRIC`: Invalid metric data
- `INVALID_LOG`: Invalid log data
- `INVALID_TRACE`: Invalid trace data
- `INVALID_ALERT`: Invalid alert data
- `SERIALIZATION_ERROR`: Data serialization failed
- `DESERIALIZATION_ERROR`: Data deserialization failed
- `QUOTA_EXCEEDED`: Quota limit exceeded
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded
- `NOT_FOUND`: Resource not found
- `UNAUTHORIZED`: Authentication failed
- `FORBIDDEN`: Access denied

## Performance Considerations

- **Batch Operations**: Use batch operations for better performance
- **Connection Pooling**: Providers implement connection pooling
- **Retry Logic**: Automatic retry with exponential backoff
- **Sampling**: Configurable sampling rates for traces
- **Buffering**: Configurable buffer sizes for batch operations

## Best Practices

1. **Use Appropriate Providers**: Choose providers based on your monitoring needs
2. **Configure Properly**: Set appropriate timeouts, retry attempts, and batch sizes
3. **Handle Errors**: Always handle errors and implement proper fallback mechanisms
4. **Monitor Performance**: Monitor the monitoring system itself
5. **Use Structured Logging**: Use structured logging with consistent fields
6. **Implement Health Checks**: Regular health checks for all monitoring components
7. **Set Up Alerting**: Configure appropriate alerts for critical issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
