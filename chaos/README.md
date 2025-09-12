# Chaos Engineering Library

A comprehensive chaos engineering library for Go microservices that provides controlled failure injection and resilience testing capabilities.

## Features

- **Kubernetes Chaos**: Integration with Chaos Mesh for container orchestration chaos
- **HTTP Chaos**: Network-level chaos for HTTP services (latency, errors, timeouts)
- **Messaging Chaos**: Message queue chaos (delays, loss, reordering)
- **Unified Interface**: Single API for all chaos types
- **Experiment Management**: Start, stop, and monitor chaos experiments
- **Configurable**: Flexible configuration for different chaos scenarios

## Architecture

```
chaos/
├── gateway/                    # Core chaos gateway
│   ├── manager.go             # Chaos engineering manager
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Chaos provider implementations
│   ├── kubernetes/            # Kubernetes-based chaos (Chaos Mesh)
│   │   ├── provider.go        # Kubernetes implementation
│   │   └── go.mod             # Kubernetes dependencies
│   ├── http/                  # HTTP-based chaos (latency, errors)
│   │   ├── provider.go        # HTTP implementation
│   │   └── go.mod             # HTTP dependencies
│   └── messaging/             # Messaging-based chaos
│       ├── provider.go        # Messaging implementation
│       └── go.mod             # Messaging dependencies
├── test/                      # Test files
│   ├── integration/           # Integration tests
│   ├── unit/                  # Unit tests
│   └── mocks/                 # Mock providers
├── go.mod                     # Main module dependencies
└── README.md                  # Documentation
```

## Quick Start

### Installation

```bash
go get github.com/anasamu/microservices-library-go/chaos
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/chaos"
)

func main() {
    // Create a chaos manager with default providers
    manager := gateway.DefaultManager()
    
    // Initialize the manager
    ctx := context.Background()
    if err := manager.Initialize(ctx); err != nil {
        log.Fatalf("Failed to initialize chaos manager: %v", err)
    }
    defer manager.Cleanup(ctx)
    
    // Start a pod failure experiment
    config := gateway.ExperimentConfig{
        Type:       gateway.ChaosTypeKubernetes,
        Experiment: gateway.PodFailure,
        Duration:   "5m",
        Target:     "my-app-pod",
        Parameters: map[string]interface{}{
            "namespace": "default",
            "selector":  "app=my-app",
        },
    }
    
    result, err := manager.StartExperiment(ctx, config)
    if err != nil {
        log.Printf("Failed to start experiment: %v", err)
        return
    }
    
    log.Printf("Experiment started: %s", result.ID)
    
    // Stop the experiment
    defer manager.StopExperiment(ctx, gateway.ChaosTypeKubernetes, result.ID)
}
```

## Chaos Types

### Kubernetes Chaos

Kubernetes chaos experiments using Chaos Mesh:

- **Pod Failure**: Randomly kill pods
- **Network Latency**: Inject network delays
- **CPU Stress**: Stress test CPU resources
- **Memory Stress**: Stress test memory resources

```go
// Pod failure experiment
config := gateway.ExperimentConfig{
    Type:       gateway.ChaosTypeKubernetes,
    Experiment: gateway.PodFailure,
    Duration:   "5m",
    Target:     "my-app-pod",
    Parameters: map[string]interface{}{
        "namespace": "default",
        "selector":  "app=my-app",
    },
}
```

### HTTP Chaos

HTTP-level chaos experiments:

- **HTTP Latency**: Inject request/response delays
- **HTTP Error**: Return error responses
- **HTTP Timeout**: Simulate request timeouts

```go
// HTTP latency experiment
config := gateway.ExperimentConfig{
    Type:       gateway.ChaosTypeHTTP,
    Experiment: gateway.HTTPLatency,
    Duration:   "2m",
    Intensity:  0.5,
    Target:     "http://api.example.com",
    Parameters: map[string]interface{}{
        "latency_ms": 1000,
        "probability": 0.3,
    },
}
```

### Messaging Chaos

Message queue chaos experiments:

- **Message Delay**: Delay message processing
- **Message Loss**: Drop messages
- **Message Reorder**: Reorder message delivery

```go
// Message delay experiment
config := gateway.ExperimentConfig{
    Type:       gateway.ChaosTypeMessaging,
    Experiment: gateway.MessageDelay,
    Duration:   "3m",
    Target:     "order-queue",
    Parameters: map[string]interface{}{
        "delay_ms": 5000,
        "probability": 0.2,
    },
}
```

## Configuration

### Kubernetes Provider

The Kubernetes provider requires:
- Kubernetes cluster access
- Chaos Mesh installed
- Proper RBAC permissions

```go
// Custom Kubernetes configuration
config := map[string]interface{}{
    "kubeconfig": "/path/to/kubeconfig",
    "namespace":  "chaos-mesh",
}
```

### HTTP Provider

The HTTP provider can be configured with:
- Custom transport settings
- Timeout configurations
- Connection pool settings

```go
// Custom HTTP configuration
config := map[string]interface{}{
    "max_idle_conns":     100,
    "idle_conn_timeout":  "90s",
    "timeout":            "30s",
}
```

### Messaging Provider

The messaging provider supports:
- Multiple topics
- Custom message handlers
- Queue configurations

```go
// Custom messaging configuration
config := map[string]interface{}{
    "topics": []string{"orders", "payments", "notifications"},
}
```

## Experiment Management

### Starting Experiments

```go
result, err := manager.StartExperiment(ctx, config)
if err != nil {
    log.Printf("Failed to start experiment: %v", err)
    return
}
log.Printf("Experiment started: %s", result.ID)
```

### Monitoring Experiments

```go
status, err := manager.GetExperimentStatus(ctx, chaosType, experimentID)
if err != nil {
    log.Printf("Failed to get status: %v", err)
    return
}
log.Printf("Status: %s - %s", status.Status, status.Message)
```

### Stopping Experiments

```go
err := manager.StopExperiment(ctx, chaosType, experimentID)
if err != nil {
    log.Printf("Failed to stop experiment: %v", err)
    return
}
log.Println("Experiment stopped")
```

### Listing Experiments

```go
experiments, err := manager.ListExperiments(ctx, chaosType)
if err != nil {
    log.Printf("Failed to list experiments: %v", err)
    return
}

for _, exp := range experiments {
    log.Printf("Experiment %s: %s", exp.ID, exp.Status)
}
```

## Best Practices

### 1. Start Small
Begin with low-intensity experiments and gradually increase complexity.

### 2. Monitor Everything
Ensure comprehensive monitoring before running chaos experiments.

### 3. Have Rollback Plans
Always have a way to quickly stop experiments and restore normal operation.

### 4. Test in Staging First
Never run chaos experiments in production without thorough testing.

### 5. Document Experiments
Keep records of experiments and their outcomes for learning.

## Safety Features

- **Automatic Cleanup**: Experiments are automatically cleaned up on provider shutdown
- **Error Handling**: Comprehensive error handling and logging
- **Resource Limits**: Built-in resource limits to prevent system overload
- **Status Monitoring**: Real-time experiment status monitoring

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the examples in the `gateway/example.go` file
