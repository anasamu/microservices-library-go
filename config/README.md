# Configuration Library

A flexible and extensible configuration management library for Go microservices. This library provides a unified interface for managing configuration across different sources including files, environment variables, HashiCorp Vault, and Consul.

## Features

- **Multiple Configuration Sources**: Support for file-based, environment variable, Vault, and Consul configuration
- **Unified Interface**: Single API for all configuration providers
- **Configuration Watching**: Real-time configuration change detection
- **Type Safety**: Strongly typed configuration structures
- **Extensible**: Easy to add new configuration providers
- **Default Values**: Built-in sensible defaults for all configuration options
- **Dynamic Configuration**: Support for custom configuration fields that can be added at runtime
- **Microservices Configuration**: Built-in support for microservices configuration with health checks, circuit breakers, and load balancing
- **Service Discovery**: Easy service URL generation and health check URL management
- **Custom Services**: Ability to add custom services dynamically

## Architecture

The configuration library follows a provider pattern with the following components:

- **Types**: Core configuration structures and interfaces
- **Gateway**: Configuration manager that orchestrates different providers
- **Providers**: Implementation of different configuration sources
- **Examples**: Usage examples and best practices

## Supported Providers

### File Provider
Loads configuration from YAML, JSON, or TOML files with automatic file watching.

```go
provider := file.NewProvider("config.yaml", "yaml")
```

### Environment Provider
Loads configuration from environment variables with customizable prefixes.

```go
provider := env.NewProvider("MYAPP_")
```

### Vault Provider
Loads configuration from HashiCorp Vault with automatic secret management.

```go
provider, err := vault.NewProvider("http://localhost:8200", "token", "secret/config")
```

### Consul Provider
Loads configuration from HashiCorp Consul KV store.

```go
provider, err := consul.NewProvider("localhost:8500", "config/")
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
    "log"
    
    "github.com/anasamu/microservices-library-go/config"
    "github.com/anasamu/microservices-library-go/config/providers/file"
)

func main() {
    // Create configuration manager
    manager := gateway.NewManager()
    
    // Register file provider
    fileProvider := file.NewProvider("config.yaml", "yaml")
    manager.RegisterProvider("file", fileProvider)
    manager.SetCurrentProvider("file")
    
    // Load configuration
    config, err := manager.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use configuration
    fmt.Printf("Server running on port: %s\n", config.Server.Port)
}
```

### 3. Configuration File Example

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  environment: "development"
  service_name: "my-service"
  version: "1.0.0"

database:
  postgresql:
    host: "localhost"
    port: 5432
    user: "postgres"
    password: "password"
    dbname: "mydb"
    sslmode: "disable"
    max_conns: 25
    min_conns: 5

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

logging:
  level: "info"
  format: "json"
  output: "stdout"

monitoring:
  prometheus:
    enabled: true
    port: "9090"
    path: "/metrics"
  jaeger:
    enabled: false
    endpoint: ""
    service: "my-service"
```

## Advanced Usage

### Configuration Watching

```go
// Watch for configuration changes
err := manager.Watch(func(newConfig *types.Config) {
    fmt.Printf("Configuration changed! New port: %s\n", newConfig.Server.Port)
})
```

### Multiple Providers

```go
// Register multiple providers
manager.RegisterProvider("file", fileProvider)
manager.RegisterProvider("env", envProvider)
manager.RegisterProvider("vault", vaultProvider)

// Switch between providers
manager.SetCurrentProvider("vault")
config, err := manager.Load()
```

### Configuration Updates

```go
// Update configuration programmatically
err := manager.UpdateConfig(func(config *types.Config) {
    config.Server.Port = "9090"
    config.Server.Environment = "production"
})
```

## Configuration Structure

The configuration structure includes the following sections:

- **Server**: HTTP server configuration
- **Database**: Database connection settings (PostgreSQL, MongoDB)
- **Redis**: Redis cache configuration
- **Vault**: HashiCorp Vault settings
- **Logging**: Logging configuration
- **Monitoring**: Prometheus and Jaeger settings
- **Storage**: Object storage configuration (MinIO, S3)
- **Search**: Elasticsearch configuration
- **Auth**: JWT authentication settings
- **RabbitMQ**: Message queue configuration
- **Kafka**: Apache Kafka settings
- **gRPC**: gRPC server configuration
- **Services**: Microservices configuration with health checks, circuit breakers, and load balancing
- **Custom**: Dynamic custom configuration fields

## Environment Variables

The library supports loading configuration from environment variables. Use the following naming convention:

- `SERVER_PORT` for server port
- `DB_POSTGRESQL_HOST` for database host
- `REDIS_PASSWORD` for Redis password
- `AUTH_JWT_SECRET_KEY` for JWT secret
- `SERVICES_USER_SERVICE_HOST` for user service host
- `CUSTOM_FEATURE_FLAGS_NEW_UI` for custom configuration

### Custom Configuration via Environment Variables

You can add custom configuration using environment variables with the `CUSTOM_` prefix:

```bash
export CUSTOM_FEATURE_FLAGS_NEW_UI=true
export CUSTOM_API_RATE_LIMIT=1000
export CUSTOM_CACHE_TTL=3600
```

These will be automatically parsed and available in the `Custom` field of the configuration.

## Microservices Configuration

The library provides comprehensive support for microservices configuration including:

### Service Configuration

Each service can be configured with:

- **Basic Settings**: Name, host, port, protocol, version, environment
- **Health Checks**: Enabled/disabled, path, interval, timeout, retries, grace period
- **Retry Logic**: Max attempts, delays, exponential backoff, jitter
- **Timeouts**: Connect, read, write, and total timeouts
- **Circuit Breaker**: Failure thresholds, success thresholds, timeout, max requests
- **Load Balancing**: Strategy (round_robin, least_conn, random, weighted), servers, weights
- **Custom Fields**: Service-specific configuration

### Predefined Services

The library comes with predefined configurations for common services:

- `user_service` - User management service
- `auth_service` - Authentication service
- `payment_service` - Payment processing service
- `notification_service` - Notification service
- `file_service` - File management service
- `report_service` - Reporting service

### Custom Services

You can add custom services dynamically:

```go
customService := types.ServiceConfig{
    Name:     "analytics-service",
    Host:     "analytics.example.com",
    Port:     "8080",
    Protocol: "https",
    HealthCheck: types.HealthCheckConfig{
        Enabled: true,
        Path:    "/health",
    },
    Custom: map[string]interface{}{
        "api_key":    "secret-key",
        "batch_size": 100,
    },
}
config.SetServiceConfig("analytics_service", customService)
```

### Service URL Generation

The library provides helper methods for service URL generation:

```go
// Get service URL
userServiceURL, err := config.GetServiceURL("user_service")
// Returns: http://localhost:8001

// Get health check URL
healthURL, err := config.GetServiceHealthCheckURL("user_service")
// Returns: http://localhost:8001/health

// Get all services
allServices := config.GetAllServices()
```

## Dynamic Configuration

The library supports dynamic configuration through the `Custom` field:

### Setting Custom Values

```go
config.SetCustomValue("feature_flags.new_ui", true)
config.SetCustomValue("api_rate_limit", 1000)
config.SetCustomValue("cache_ttl", 3600)
```

### Getting Custom Values

```go
// Type-safe getters with defaults
newUI := config.GetCustomBool("feature_flags.new_ui", false)
rateLimit := config.GetCustomInt("api_rate_limit", 100)
cacheTTL := config.GetCustomInt("cache_ttl", 300)

// Generic getter
value, exists := config.GetCustomValue("any_key")
```

### Custom Configuration in YAML

```yaml
# Custom configuration section
feature_flags:
  new_ui: true
  beta_features: false

api_settings:
  rate_limit: 1000
  timeout: 30

business_rules:
  max_file_size: 10485760
  max_users_per_organization: 1000
```

## Best Practices

1. **Use Environment Variables for Secrets**: Store sensitive information like passwords and API keys in environment variables or Vault.

2. **Provide Sensible Defaults**: The library provides sensible defaults for all configuration options.

3. **Use Configuration Watching**: Enable configuration watching for dynamic configuration updates without service restarts.

4. **Validate Configuration**: Always validate configuration after loading to ensure all required fields are present.

5. **Use Type-Safe Access**: Use the provided getter methods for type-safe configuration access.

## Examples

See the `examples/` directory for complete usage examples:

- Basic file-based configuration
- Environment variable configuration
- Vault integration
- Configuration watching
- Multiple provider usage

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
