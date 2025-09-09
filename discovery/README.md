# Service Discovery Library

A comprehensive Go library for service discovery that provides a unified interface for multiple service discovery backends including Consul, Kubernetes, etcd, and static configuration.

## Features

- **Multi-Provider Support**: Support for Consul, Kubernetes, etcd, and static configuration
- **Unified Interface**: Single API for all service discovery operations
- **Health Management**: Built-in health checking and status management
- **Service Watching**: Real-time service change notifications
- **Load Balancing**: Built-in load balancing strategies
- **Retry Logic**: Automatic retry with configurable attempts and delays
- **Fallback Support**: Automatic fallback between providers
- **Comprehensive Logging**: Detailed logging with structured fields
- **Thread-Safe**: All operations are thread-safe and concurrent

## Architecture

```
discovery/
├── gateway/                    # Core service discovery gateway
│   ├── manager.go             # Service discovery manager
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Provider implementations
│   ├── consul/                # Consul provider
│   │   ├── provider.go        # Consul implementation
│   │   └── go.mod             # Consul dependencies
│   ├── kubernetes/            # Kubernetes provider
│   │   ├── provider.go        # Kubernetes implementation
│   │   └── go.mod             # Kubernetes dependencies
│   ├── etcd/                  # etcd provider
│   │   ├── provider.go        # etcd implementation
│   │   └── go.mod             # etcd dependencies
│   └── static/                # Static config provider
│       ├── provider.go        # Static implementation
│       └── go.mod             # Static dependencies
├── types/                     # Common types and interfaces
│   ├── types.go               # Type definitions
│   └── go.mod                 # Types dependencies
├── test/                      # Test files
│   ├── integration/           # Integration tests
│   ├── unit/                  # Unit tests
│   └── mocks/                 # Mock providers
├── go.mod                     # Main module dependencies
└── README.md                  # This documentation
```

## Quick Start

### Installation

```bash
go get github.com/anasamu/microservices-library-go/discovery
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/anasamu/microservices-library-go/discovery/gateway"
    "github.com/anasamu/microservices-library-go/discovery/providers/consul"
    "github.com/anasamu/microservices-library-go/discovery/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create discovery manager
    manager := gateway.NewDiscoveryManager(nil, logger)
    
    // Create and register Consul provider
    consulConfig := &consul.ConsulConfig{
        Address: "localhost:8500",
        Timeout: 30 * time.Second,
    }
    
    consulProvider, err := consul.NewConsulProvider(consulConfig, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    err = manager.RegisterProvider(consulProvider)
    if err != nil {
        log.Fatal(err)
    }
    
    // Connect to provider
    ctx := context.Background()
    err = consulProvider.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register a service
    registration := &types.ServiceRegistration{
        ID:       "web-service-1",
        Name:     "web-service",
        Address:  "192.168.1.100",
        Port:     8080,
        Protocol: "http",
        Tags:     []string{"web", "api"},
        Health:   types.HealthPassing,
        TTL:      30 * time.Second,
    }
    
    err = manager.RegisterService(ctx, registration)
    if err != nil {
        log.Fatal(err)
    }
    
    // Discover services
    query := &types.ServiceQuery{
        Name: "web-service",
    }
    
    services, err := manager.DiscoverServices(ctx, query)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, service := range services {
        log.Printf("Found service: %s with %d instances", 
            service.Name, len(service.Instances))
    }
}
```

## Providers

### Consul Provider

The Consul provider integrates with HashiCorp Consul for service discovery.

```go
consulConfig := &consul.ConsulConfig{
    Address:    "localhost:8500",
    Token:      "your-token",
    Datacenter: "dc1",
    Namespace:  "production",
    Timeout:    30 * time.Second,
}

consulProvider, err := consul.NewConsulProvider(consulConfig, logger)
```

**Features:**
- Service registration and deregistration
- Health checking with TTL
- Service discovery with filtering
- Real-time service watching
- Load balancing support
- Multi-datacenter support

### Kubernetes Provider

The Kubernetes provider discovers services from Kubernetes clusters.

```go
k8sConfig := &kubernetes.KubernetesConfig{
    Namespace: "default",
    InCluster: true, // Set to false for external access
    Timeout:   30 * time.Second,
}

k8sProvider, err := kubernetes.NewKubernetesProvider(k8sConfig, logger)
```

**Features:**
- Automatic service discovery from Kubernetes services
- Endpoint monitoring
- Namespace filtering
- In-cluster and external access modes
- Service watching via Kubernetes API

**Note:** Service registration is not supported as services are managed by Kubernetes resources.

### etcd Provider

The etcd provider uses etcd for distributed service discovery.

```go
etcdConfig := &etcd.EtcdConfig{
    Endpoints: []string{"localhost:2379"},
    Username:  "user",
    Password:  "pass",
    Namespace: "/services",
    Timeout:   30 * time.Second,
}

etcdProvider, err := etcd.NewEtcdProvider(etcdConfig, logger)
```

**Features:**
- Distributed service registration
- TTL-based service expiration
- Real-time watching with etcd watch API
- Namespace isolation
- High availability with etcd clustering

### Static Provider

The static provider uses configuration-based service discovery.

```go
staticConfig := &static.StaticConfig{
    Services: []*types.ServiceRegistration{
        {
            ID:       "db-1",
            Name:     "database",
            Address:  "192.168.1.100",
            Port:     5432,
            Protocol: "tcp",
            Tags:     []string{"database", "postgresql"},
            Health:   types.HealthPassing,
        },
    },
    Timeout: 30 * time.Second,
}

staticProvider, err := static.NewStaticProvider(staticConfig, logger)
```

**Features:**
- Configuration-based service definitions
- Runtime service management
- No external dependencies
- Perfect for testing and development
- Fallback provider support

## Service Discovery Operations

### Service Registration

```go
registration := &types.ServiceRegistration{
    ID:       "unique-service-id",
    Name:     "service-name",
    Address:  "192.168.1.100",
    Port:     8080,
    Protocol: "http",
    Tags:     []string{"web", "api", "v1"},
    Metadata: map[string]string{
        "version": "1.0.0",
        "env":     "production",
    },
    Health: types.HealthPassing,
    TTL:    30 * time.Second,
}

err := manager.RegisterService(ctx, registration)
```

### Service Discovery

```go
// Discover all instances of a service
query := &types.ServiceQuery{
    Name: "web-service",
}

services, err := manager.DiscoverServices(ctx, query)

// Discover with filtering
query := &types.ServiceQuery{
    Name: "web-service",
    Tags: []string{"api", "v1"},
    Health: types.HealthPassing,
    Limit: 10,
}

services, err := manager.DiscoverServices(ctx, query)
```

### Health Management

```go
// Set service health
err := manager.SetHealth(ctx, "service-id", types.HealthPassing)

// Get service health
health, err := manager.GetHealth(ctx, "service-id")
```

### Service Watching

```go
watchOptions := &types.WatchOptions{
    ServiceName: "web-service",
    Tags:        []string{"api"},
    HealthOnly:  false,
    Timeout:     30 * time.Second,
}

eventChan, err := manager.WatchServices(ctx, watchOptions)
if err != nil {
    log.Fatal(err)
}

for event := range eventChan {
    switch event.Type {
    case types.EventInstanceAdded:
        log.Printf("New instance added: %s", event.Instance.ID)
    case types.EventInstanceRemoved:
        log.Printf("Instance removed: %s", event.Instance.ID)
    case types.EventHealthChanged:
        log.Printf("Health changed for: %s", event.Instance.ID)
    }
}
```

## Load Balancing

The library includes built-in load balancing strategies:

```go
// Round Robin
lb := &types.LoadBalancerRoundRobin{}

// Random
lb := &types.LoadBalancerRandom{}

// Weighted (based on service instance weight)
lb := &types.LoadBalancerWeighted{}

// Least Connections
lb := &types.LoadBalancerLeastConn{}

// Hash-based (consistent hashing)
lb := &types.LoadBalancerHash{}

// Select instance
instance, err := lb.SelectInstance(instances)
```

## Configuration

### Manager Configuration

```go
config := &gateway.ManagerConfig{
    DefaultProvider: "consul",
    RetryAttempts:   3,
    RetryDelay:      time.Second,
    Timeout:         30 * time.Second,
    FallbackEnabled: true,
    Metadata: map[string]string{
        "environment": "production",
    },
}

manager := gateway.NewDiscoveryManager(config, logger)
```

### Provider-Specific Configuration

Each provider has its own configuration structure. See the individual provider documentation for details.

## Error Handling

The library provides comprehensive error handling with specific error codes:

```go
if err != nil {
    if discoveryErr, ok := err.(*types.DiscoveryError); ok {
        switch discoveryErr.Code {
        case types.ErrCodeServiceNotFound:
            log.Printf("Service not found: %s", discoveryErr.Service)
        case types.ErrCodeConnection:
            log.Printf("Connection error: %s", discoveryErr.Message)
        case types.ErrCodeTimeout:
            log.Printf("Operation timeout: %s", discoveryErr.Message)
        default:
            log.Printf("Discovery error: %s", discoveryErr.Message)
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Monitoring and Statistics

```go
// Get provider statistics
stats, err := manager.GetStats(ctx)
for providerName, stat := range stats {
    log.Printf("Provider %s: %d services, %d instances", 
        providerName, stat.ServicesRegistered, stat.InstancesActive)
}

// Get provider information
providerInfo := manager.GetProviderInfo()
for name, info := range providerInfo {
    log.Printf("Provider %s: Connected=%v, Features=%d", 
        name, info.IsConnected, len(info.SupportedFeatures))
}

// List all services
allServices, err := manager.ListServices(ctx)
for providerName, services := range allServices {
    log.Printf("Provider %s has %d services", providerName, len(services))
}
```

## Testing

The library includes comprehensive testing support:

```go
// Unit tests
go test ./discovery/...

// Integration tests
go test ./discovery/test/integration/...

// Run with coverage
go test -cover ./discovery/...
```

### Mock Providers

For testing, you can use mock providers:

```go
// Create mock provider for testing
mockProvider := &mocks.MockDiscoveryProvider{
    Services: make(map[string]*types.Service),
}

// Register mock provider
err := manager.RegisterProvider(mockProvider)
```

## Best Practices

1. **Provider Selection**: Choose the appropriate provider based on your infrastructure
2. **Health Checks**: Always implement proper health checks for your services
3. **TTL Management**: Set appropriate TTL values for service registrations
4. **Error Handling**: Implement proper error handling and retry logic
5. **Monitoring**: Monitor service discovery metrics and health
6. **Security**: Use proper authentication and encryption for service discovery
7. **Fallback**: Implement fallback strategies for service discovery failures

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the examples in the `gateway/example.go` file
