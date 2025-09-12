# Microservices Middleware Library

A comprehensive, flexible, and extensible middleware library for Go microservices that provides a unified interface for common middleware patterns including authentication, logging, monitoring, rate limiting, circuit breaking, caching, storage, communication, messaging, chaos engineering, and failover.

## Features

- **Multiple Middleware Providers**: Authentication, Logging, Monitoring, Rate Limiting, Circuit Breaker, Caching, Storage, Communication, Messaging, Chaos Engineering, Failover
- **Unified Interface**: Consistent API across all middleware providers
- **HTTP Middleware Support**: Easy integration with HTTP handlers
- **Middleware Chains**: Compose multiple middleware into chains
- **Dynamic Configuration**: Runtime provider registration and configuration
- **Comprehensive Logging**: Structured logging with context
- **Health Monitoring**: Built-in health checks and statistics
- **Type Safety**: Strong typing with comprehensive interfaces
- **Extensible**: Easy to add custom middleware providers

## Architecture

```
middleware/
├── gateway/           # Core middleware manager and gateway
├── types/            # Shared types and interfaces
├── providers/        # Middleware provider implementations
│   ├── auth/         # Authentication middleware (JWT, OAuth, Basic Auth)
│   ├── logging/      # Logging middleware (structured, request/response)
│   ├── monitoring/   # Monitoring middleware (metrics, tracing, health)
│   ├── ratelimit/    # Rate limiting middleware (token bucket, sliding window)
│   ├── circuitbreaker/ # Circuit breaker middleware
│   ├── cache/        # Caching middleware
│   ├── storage/      # Storage middleware
│   ├── communication/ # Communication middleware (HTTP, gRPC, WebSocket, etc.)
│   ├── messaging/    # Messaging middleware (Kafka, NATS, RabbitMQ, etc.)
│   ├── chaos/        # Chaos engineering middleware
│   └── failover/     # Failover middleware
└── test/             # Test files
    ├── integration/  # Integration tests
    ├── unit/         # Unit tests
    └── mocks/        # Mock providers
```

## Quick Start

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Basic Usage

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/sirupsen/logrus"
    "github.com/anasamu/microservices-library-go/middleware"
    "github.com/anasamu/microservices-library-go/middleware/providers/auth"
    "github.com/anasamu/microservices-library-go/middleware/providers/logging"
    "github.com/anasamu/microservices-library-go/middleware/providers/storage"
    "github.com/anasamu/microservices-library-go/middleware/providers/communication"
    "github.com/anasamu/microservices-library-go/middleware/providers/messaging"
    "github.com/anasamu/microservices-library-go/middleware/types"
)

func main() {
    // Initialize logger
    logger := logrus.New()
    
    // Create middleware manager
    manager := gateway.NewMiddlewareManager(gateway.DefaultManagerConfig(), logger)
    
    // Register authentication provider
    authProvider := auth.NewAuthProvider(auth.DefaultAuthConfig(), logger)
    manager.RegisterProvider(authProvider)
    
    // Register logging provider
    loggingProvider := logging.NewLoggingProvider(logging.DefaultLoggingConfig(), logger)
    manager.RegisterProvider(loggingProvider)
    
    // Register storage provider
    storageProvider := storage.NewStorageProvider(storage.DefaultStorageConfig(), logger)
    manager.RegisterProvider(storageProvider)
    
    // Register communication provider
    communicationProvider := communication.NewCommunicationProvider(communication.DefaultCommunicationConfig(), logger)
    manager.RegisterProvider(communicationProvider)
    
    // Register messaging provider
    messagingProvider := messaging.NewMessagingProvider(messaging.DefaultMessagingConfig(), logger)
    manager.RegisterProvider(messagingProvider)
    
    // Create HTTP middleware
    authMiddleware, err := manager.CreateHTTPMiddleware("auth", &types.MiddlewareConfig{
        Name:    "auth-middleware",
        Type:    "authentication",
        Enabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    loggingMiddleware, err := manager.CreateHTTPMiddleware("logging", &types.MiddlewareConfig{
        Name:    "logging-middleware",
        Type:    "logging",
        Enabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    storageMiddleware, err := manager.CreateHTTPMiddleware("storage", &types.MiddlewareConfig{
        Name:    "storage-middleware",
        Type:    "storage",
        Enabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    communicationMiddleware, err := manager.CreateHTTPMiddleware("communication", &types.MiddlewareConfig{
        Name:    "communication-middleware",
        Type:    "communication",
        Enabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    messagingMiddleware, err := manager.CreateHTTPMiddleware("messaging", &types.MiddlewareConfig{
        Name:    "messaging-middleware",
        Type:    "messaging",
        Enabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Setup HTTP server with middleware
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    // Chain middlewares
    wrappedHandler := authMiddleware(loggingMiddleware(storageMiddleware(communicationMiddleware(messagingMiddleware(handler)))))
    
    http.Handle("/api/", wrappedHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3. Middleware Chain Usage

```go
// Create middleware chain
config := &types.MiddlewareConfig{
    Name:     "api-chain",
    Type:     "api",
    Enabled:  true,
    Priority: 1,
    Timeout:  30 * time.Second,
    Rules: []types.MiddlewareRule{
        {
            ID:          "auth-rule",
            Name:        "Authentication Rule",
            Description: "Require authentication for API endpoints",
            Enabled:     true,
            Priority:    1,
            Conditions: []types.MiddlewareCondition{
                {
                    ID:       "path-condition",
                    Type:     "path",
                    Field:    "path",
                    Operator: "starts_with",
                    Value:    "/api/",
                },
            },
            Actions: []types.MiddlewareAction{
                {
                    ID:   "auth-action",
                    Type: "authenticate",
                    Parameters: map[string]interface{}{
                        "provider": "jwt",
                    },
                },
            },
        },
    },
}

chain, err := manager.CreateChain(context.Background(), "auth", config)
if err != nil {
    log.Fatal(err)
}

// Execute chain
request := &types.MiddlewareRequest{
    ID:        "req-123",
    Type:      "http",
    Method:    "GET",
    Path:      "/api/users",
    Headers:   map[string]string{"Authorization": "Bearer token123"},
    Context:   make(map[string]interface{}),
    Timestamp: time.Now(),
}

response, err := manager.ExecuteChain(context.Background(), "auth", chain, request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Chain response: %+v", response)
```

## Middleware Providers

### Authentication Provider

Provides JWT, OAuth2, Basic Auth, and API Key authentication.

```go
// Configure authentication
authConfig := map[string]interface{}{
    "jwt_secret_key": "your-secret-key",
    "jwt_expiry":     15 * time.Minute,
    "jwt_issuer":     "your-service",
    "jwt_audience":   "your-clients",
    "excluded_paths": []string{"/health", "/metrics"},
    "required_roles": []string{"user", "admin"},
}

authProvider.Configure(authConfig)
```

**Features:**
- JWT token validation
- OAuth2 integration
- Basic authentication
- API key authentication
- Role-based access control
- Scope-based authorization
- Path exclusions

### Logging Provider

Provides structured logging for requests and responses.

```go
// Configure logging
loggingConfig := map[string]interface{}{
    "log_level":       "info",
    "log_format":      "json",
    "include_headers": true,
    "include_body":    false,
    "include_query":   true,
    "max_body_size":   1024,
    "excluded_paths":  []string{"/health", "/metrics"},
}

loggingProvider.Configure(loggingConfig)
```

**Features:**
- Structured JSON logging
- Request/response logging
- Configurable log levels
- Header/body filtering
- Performance metrics
- Audit logging

### Rate Limiting Provider

Implements token bucket algorithm for rate limiting.

```go
// Configure rate limiting
rateLimitConfig := map[string]interface{}{
    "requests_per_second": 100,
    "burst_size":         200,
    "window_size":        time.Minute,
    "key_extractor":      "ip", // ip, user, custom
    "excluded_paths":     []string{"/health", "/metrics"},
}

rateLimitProvider.Configure(rateLimitConfig)
```

**Features:**
- Token bucket algorithm
- Sliding window rate limiting
- Fixed window rate limiting
- Leaky bucket algorithm
- Per-IP/user rate limiting
- Custom key extraction

### Circuit Breaker Provider

Implements circuit breaker pattern for fault tolerance.

```go
// Configure circuit breaker
circuitBreakerConfig := map[string]interface{}{
    "failure_threshold": 5,
    "success_threshold": 3,
    "timeout":          30 * time.Second,
    "max_requests":     10,
    "excluded_paths":   []string{"/health", "/metrics"},
}

circuitBreakerProvider.Configure(circuitBreakerConfig)
```

**Features:**
- Circuit breaker pattern
- Retry logic
- Timeout handling
- Bulkhead isolation
- Fallback mechanisms
- Health monitoring

### Monitoring Provider

Provides metrics, tracing, and health monitoring.

```go
// Configure monitoring
monitoringConfig := map[string]interface{}{
    "metrics_enabled": true,
    "tracing_enabled": true,
    "health_checks":   true,
    "prometheus_port": 9090,
    "jaeger_endpoint": "http://jaeger:14268/api/traces",
}

monitoringProvider.Configure(monitoringConfig)
```

**Features:**
- Prometheus metrics
- Jaeger tracing
- Health checks
- Performance profiling
- Custom metrics
- Dashboard integration

### Cache Provider

Provides caching middleware for improved performance.

```go
// Configure caching
cacheConfig := map[string]interface{}{
	"cache_type":      "memory", // memory, redis
	"ttl":            5 * time.Minute,
	"max_size":       1000,
	"excluded_paths": []string{"/health", "/metrics"},
}

cacheProvider.Configure(cacheConfig)
```

**Features:**
- In-memory caching
- Redis caching
- Distributed caching
- Cache invalidation
- Cache warming
- TTL management

### Storage Provider

Provides file storage middleware for handling file uploads, downloads, and metadata management.

```go
// Configure storage
storageConfig := map[string]interface{}{
	"storage_type":        "s3", // local, s3, gcs, azure
	"base_path":           "/uploads",
	"max_file_size":       int64(50 * 1024 * 1024), // 50MB
	"allowed_mime_types":  []string{"image/*", "application/pdf"},
	"excluded_paths":      []string{"/health", "/metrics"},
	"cache_enabled":       true,
	"cache_ttl":          1 * time.Hour,
	"compression_enabled": true,
	"encryption_enabled":  false,
}

storageProvider.Configure(storageConfig)
```

**Features:**
- File upload/download handling
- Multiple storage backends (local, S3, GCS, Azure)
- File metadata management
- MIME type validation
- File size limits
- File caching
- Compression support
- Encryption support
- File hash calculation
- Path-based file organization

### Communication Provider

Provides communication middleware for handling different protocols and communication patterns.

```go
// Configure communication
communicationConfig := map[string]interface{}{
	"protocols":         []string{"http", "grpc", "websocket", "graphql", "sse", "quic"},
	"default_protocol":  "http",
	"timeout":           30 * time.Second,
	"retry_attempts":    3,
	"retry_delay":       1 * time.Second,
	"max_connections":   100,
	"keep_alive":        true,
	"compression":       true,
	"encryption":        false,
	"load_balancing":    false,
	"circuit_breaker":   false,
	"rate_limiting":     false,
	"excluded_paths":    []string{"/health", "/metrics", "/admin"},
}

communicationProvider.Configure(communicationConfig)
```

**Features:**
- Multiple protocol support (HTTP, gRPC, WebSocket, GraphQL, SSE, QUIC)
- Protocol detection and routing
- Connection pooling and management
- Load balancing capabilities
- Circuit breaker integration
- Rate limiting support
- Compression and encryption
- Retry logic with exponential backoff
- Timeout management
- Custom headers and metadata

### Messaging Provider

Provides messaging middleware for handling message queues and event-driven communication.

```go
// Configure messaging
messagingConfig := map[string]interface{}{
	"queues":            []string{"kafka", "nats", "rabbitmq", "sqs", "redis"},
	"default_queue":     "kafka",
	"broker_url":        "localhost:9092",
	"topic_prefix":      "microservice",
	"group_id":          "default-group",
	"auto_commit":       true,
	"batch_size":        100,
	"batch_timeout":     1 * time.Second,
	"retry_attempts":    3,
	"retry_delay":       1 * time.Second,
	"max_message_size":  1024 * 1024, // 1MB
	"compression":       true,
	"encryption":        false,
	"dead_letter_queue": true,
	"message_ttl":       24 * time.Hour,
	"priority":          0,
	"excluded_topics":   []string{"health", "metrics", "admin"},
}

messagingProvider.Configure(messagingConfig)
```

**Features:**
- Multiple queue support (Kafka, NATS, RabbitMQ, SQS, Redis)
- Message publishing and consuming
- Batch processing capabilities
- Dead letter queue handling
- Message compression and encryption
- Retry logic with configurable attempts
- Message TTL and priority support
- Consumer group management
- Topic-based routing
- Event sourcing and CQRS patterns
- Message acknowledgment and rejection

### Chaos Engineering Provider

Provides chaos engineering capabilities for testing resilience.

```go
// Configure chaos engineering
chaosConfig := map[string]interface{}{
    "latency_injection": true,
    "error_injection":   true,
    "resource_exhaustion": false,
    "network_partition": false,
    "failure_rate":     0.1, // 10% failure rate
}

chaosProvider.Configure(chaosConfig)
```

**Features:**
- Latency injection
- Error injection
- Resource exhaustion
- Network partitioning
- Service failure simulation
- Configurable chaos patterns

### Failover Provider

Provides failover and load balancing capabilities.

```go
// Configure failover
failoverConfig := map[string]interface{}{
    "load_balancing":     "round_robin", // round_robin, least_connections
    "health_checking":    true,
    "service_discovery":  true,
    "failover_enabled":   true,
    "graceful_shutdown":  true,
}

failoverProvider.Configure(failoverConfig)
```

**Features:**
- Load balancing
- Health checking
- Service discovery
- Failover mechanisms
- Graceful shutdown
- Circuit breaker integration

## Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Authentication Configuration
export JWT_SECRET_KEY="your-secret-key"
export JWT_ACCESS_EXPIRY="15m"
export JWT_REFRESH_EXPIRY="168h"
export JWT_ISSUER="your-service"
export JWT_AUDIENCE="your-clients"

# Logging Configuration
export LOG_LEVEL="info"
export LOG_FORMAT="json"
export LOG_INCLUDE_HEADERS="true"
export LOG_INCLUDE_BODY="false"

# Rate Limiting Configuration
export RATE_LIMIT_RPS="100"
export RATE_LIMIT_BURST="200"
export RATE_LIMIT_WINDOW="1m"

# Monitoring Configuration
export METRICS_ENABLED="true"
export TRACING_ENABLED="true"
export PROMETHEUS_PORT="9090"
```

### Dynamic Configuration

```go
// Configure providers dynamically
configs := map[string]map[string]interface{}{
    "auth": {
        "jwt_secret_key": "your-secret-key",
        "jwt_expiry":     time.Minute * 15,
        "excluded_paths": []string{"/health", "/metrics"},
    },
    "logging": {
        "log_level":      "debug",
        "log_format":     "json",
        "include_headers": true,
    },
    "ratelimit": {
        "requests_per_second": 200,
        "burst_size":         400,
        "key_extractor":      "ip",
    },
}

for providerName, config := range configs {
    provider, err := manager.GetProvider(providerName)
    if err != nil {
        log.Printf("Provider %s not found: %v", providerName, err)
        continue
    }
    
    if err := provider.Configure(config); err != nil {
        log.Printf("Failed to configure provider %s: %v", providerName, err)
    }
}
```

## Advanced Usage

### Custom Middleware Provider

```go
type CustomProvider struct {
    name   string
    logger *logrus.Logger
    config map[string]interface{}
}

func (cp *CustomProvider) GetName() string {
    return cp.name
}

func (cp *CustomProvider) GetType() string {
    return "custom"
}

func (cp *CustomProvider) GetSupportedFeatures() []types.MiddlewareFeature {
    return []types.MiddlewareFeature{
        types.FeatureCustom,
    }
}

func (cp *CustomProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
    // Custom request processing logic
    return &types.MiddlewareResponse{
        ID:        request.ID,
        Success:   true,
        Timestamp: time.Now(),
    }, nil
}

// Implement other required methods...

// Register custom provider
customProvider := &CustomProvider{
    name:   "custom",
    logger: logger,
}
manager.RegisterProvider(customProvider)
```

### Middleware Chain Composition

```go
// Create complex middleware chain
chain := types.NewMiddlewareChain(
    // Authentication
    func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
        return authProvider.ProcessRequest(ctx, req)
    },
    // Rate limiting
    func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
        return rateLimitProvider.ProcessRequest(ctx, req)
    },
    // Logging
    func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
        return loggingProvider.ProcessRequest(ctx, req)
    },
    // Custom logic
    func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
        // Custom middleware logic
        return nil, nil
    },
)

// Execute chain
response, err := chain.Execute(context.Background(), request)
```

### HTTP Middleware Integration

```go
// Create HTTP middleware chain
httpChain := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create middleware request
        request := &types.MiddlewareRequest{
            ID:        generateRequestID(),
            Type:      "http",
            Method:    r.Method,
            Path:      r.URL.Path,
            Headers:   extractHeaders(r.Header),
            Query:     extractQuery(r.URL.Query()),
            Context:   make(map[string]interface{}),
            Timestamp: time.Now(),
        }
        
        // Execute middleware chain
        response, err := chain.Execute(r.Context(), request)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        
        if !response.Success {
            w.WriteHeader(response.StatusCode)
            w.Write(response.Body)
            return
        }
        
        // Add response headers
        for name, value := range response.Headers {
            w.Header().Set(name, value)
        }
        
        // Call next handler
        next.ServeHTTP(w, r)
    })
}

// Use with HTTP server
http.Handle("/api/", httpChain(yourHandler))
```

## Monitoring and Health Checks

```go
// Health check
healthResults := manager.HealthCheck(context.Background())
for provider, err := range healthResults {
    if err != nil {
        log.Printf("Provider %s is unhealthy: %v", provider, err)
    } else {
        log.Printf("Provider %s is healthy", provider)
    }
}

// Get statistics
stats := manager.GetStats(context.Background())
for provider, stat := range stats {
    log.Printf("Provider %s statistics: %+v", provider, stat)
}
```

## Error Handling

The library provides comprehensive error handling:

```go
response, err := manager.ProcessRequest(ctx, "auth", request)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "provider not found"):
        // Handle provider not found
    case strings.Contains(err.Error(), "invalid configuration"):
        // Handle configuration error
    case strings.Contains(err.Error(), "rate limit exceeded"):
        // Handle rate limiting
    case strings.Contains(err.Error(), "circuit breaker open"):
        // Handle circuit breaker
    default:
        // Handle other errors
    }
}
```

## Best Practices

1. **Provider Registration**: Register providers early in application startup
2. **Configuration**: Use environment variables for configuration in production
3. **Error Handling**: Always handle errors from middleware operations
4. **Logging**: Use structured logging for better observability
5. **Health Checks**: Implement health checks for all providers
6. **Monitoring**: Enable metrics and tracing for production deployments
7. **Testing**: Use mocks for unit testing and integration tests for end-to-end validation
8. **Security**: Keep secrets secure and rotate them regularly
9. **Performance**: Monitor middleware performance and adjust configurations as needed
10. **Documentation**: Document custom middleware providers and configurations

## Examples

See the `examples/` directory for complete working examples:

- HTTP middleware integration
- Multiple provider usage
- Custom middleware development
- Error handling patterns
- Configuration management

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
