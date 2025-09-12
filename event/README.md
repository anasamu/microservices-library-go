# Event Sourcing Library

A comprehensive event sourcing library for Go microservices that provides a unified interface for event storage, retrieval, and management across multiple providers.

## Features

- **Multi-Provider Support**: PostgreSQL, Kafka, NATS
- **Event Sourcing**: Complete event sourcing implementation with snapshots
- **Stream Management**: Create, delete, and manage event streams
- **Batch Operations**: Efficient batch event processing
- **Retry Logic**: Built-in retry mechanisms with configurable attempts
- **Health Checks**: Provider health monitoring
- **Statistics**: Comprehensive event sourcing statistics
- **Flexible Configuration**: Easy configuration for different environments

## Supported Providers

### PostgreSQL
- Full event sourcing support with ACID transactions
- Snapshot support for performance optimization
- Complex queries and filtering
- Stream versioning and concurrency control

### Kafka
- High-throughput event streaming
- Event partitioning and replication
- Batch processing support
- Real-time event consumption

### NATS
- Lightweight messaging with JetStream
- Event retention policies
- Real-time event streaming
- High-performance event processing

## Installation

```bash
go get github.com/anasamu/microservices-library-go/event
```

## Quick Start

### 1. Create Event Sourcing Manager

```go
package main

import (
    "context"
    "time"
    
    "github.com/anasamu/microservices-library-go/event"
    "github.com/anasamu/microservices-library-go/event/providers/postgresql"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create manager configuration
    config := &gateway.ManagerConfig{
        DefaultProvider:   "postgresql",
        RetryAttempts:     3,
        RetryDelay:        time.Second,
        Timeout:           30 * time.Second,
        MaxEventSize:      1024 * 1024, // 1MB
        MaxBatchSize:      100,
        SnapshotThreshold: 1000,
        RetentionPeriod:   365 * 24 * time.Hour, // 1 year
    }
    
    // Create manager
    manager := gateway.NewEventSourcingManager(config, logger)
    
    // Create and register PostgreSQL provider
    pgConfig := &postgresql.PostgreSQLConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "eventsourcing",
        Username: "postgres",
        Password: "password",
        SSLMode:  "disable",
    }
    
    pgProvider, err := postgresql.NewPostgreSQLProvider(pgConfig, logger)
    if err != nil {
        panic(err)
    }
    
    if err := manager.RegisterProvider(pgProvider); err != nil {
        panic(err)
    }
    
    // Connect to provider
    ctx := context.Background()
    if err := pgProvider.Connect(ctx); err != nil {
        panic(err)
    }
    
    defer manager.Close()
}
```

### 2. Create and Manage Streams

```go
// Create a stream
streamRequest := &types.CreateStreamRequest{
    StreamID:      "user-123",
    Name:          "User Events",
    AggregateID:   "user-123",
    AggregateType: "User",
    Metadata: map[string]interface{}{
        "description": "User aggregate events",
        "version":     "1.0",
    },
}

err := manager.CreateStream(ctx, streamRequest)
if err != nil {
    panic(err)
}
```

### 3. Append Events

```go
// Append an event
eventRequest := &types.AppendEventRequest{
    StreamID:      "user-123",
    EventType:     "UserCreated",
    EventData:     map[string]interface{}{"name": "John Doe", "email": "john@example.com"},
    EventMetadata: map[string]interface{}{"source": "api", "version": "1.0"},
    AggregateID:   "user-123",
    AggregateType: "User",
    UserID:        "admin",
}

response, err := manager.AppendEvent(ctx, eventRequest)
if err != nil {
    panic(err)
}

fmt.Printf("Event appended with ID: %s\n", response.EventID)
```

### 4. Retrieve Events

```go
// Get events from a stream
getEventsRequest := &types.GetEventsRequest{
    StreamID: "user-123",
    Limit:    10,
}

eventsResponse, err := manager.GetEvents(ctx, getEventsRequest)
if err != nil {
    panic(err)
}

for _, event := range eventsResponse.Events {
    fmt.Printf("Event: %s, Type: %s, Data: %v\n", 
        event.ID, event.EventType, event.EventData)
}
```

### 5. Create Snapshots

```go
// Create a snapshot
snapshotRequest := &types.CreateSnapshotRequest{
    StreamID:      "user-123",
    AggregateID:   "user-123",
    AggregateType: "User",
    Version:       10,
    Data: map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "id":    "user-123",
    },
    Metadata: map[string]interface{}{"snapshot_type": "user_state"},
}

snapshotResponse, err := manager.CreateSnapshot(ctx, snapshotRequest)
if err != nil {
    panic(err)
}

fmt.Printf("Snapshot created with ID: %s\n", snapshotResponse.SnapshotID)
```

## Advanced Usage

### Multiple Providers

```go
// Register multiple providers
kafkaProvider, _ := kafka.NewKafkaProvider(kafkaConfig, logger)
natsProvider, _ := nats.NewNATSProvider(natsConfig, logger)

manager.RegisterProvider(kafkaProvider)
manager.RegisterProvider(natsProvider)

// Use specific provider
response, err := manager.AppendEventWithProvider(ctx, "kafka", eventRequest)
```

### Batch Operations

```go
// Append multiple events in batch
batchRequest := &types.AppendEventsBatchRequest{
    StreamID: "user-123",
    Events: []*types.Event{
        {
            EventType:     "UserUpdated",
            EventData:     map[string]interface{}{"name": "John Smith"},
            EventMetadata: map[string]interface{}{"source": "api"},
            AggregateID:   "user-123",
            AggregateType: "User",
        },
        {
            EventType:     "UserEmailChanged",
            EventData:     map[string]interface{}{"email": "john.smith@example.com"},
            EventMetadata: map[string]interface{}{"source": "api"},
            AggregateID:   "user-123",
            AggregateType: "User",
        },
    },
}

batchResponse, err := manager.AppendEventsBatch(ctx, batchRequest)
if err != nil {
    panic(err)
}

fmt.Printf("Appended %d events, %d failed\n", 
    batchResponse.AppendedCount, batchResponse.FailedCount)
```

### Health Monitoring

```go
// Check health of all providers
healthResults := manager.HealthCheck(ctx)
for providerName, err := range healthResults {
    if err != nil {
        fmt.Printf("Provider %s is unhealthy: %v\n", providerName, err)
    } else {
        fmt.Printf("Provider %s is healthy\n", providerName)
    }
}

// Get statistics
stats, err := manager.GetStats(ctx)
if err != nil {
    panic(err)
}

for providerName, stat := range stats {
    fmt.Printf("Provider %s: %d events, %d streams\n", 
        providerName, stat.TotalEvents, stat.TotalStreams)
}
```

## Event Sourcing Aggregate Example

```go
type UserAggregate struct {
    ID       string
    Name     string
    Email    string
    Version  int64
    manager  *gateway.EventSourcingManager
}

func NewUserAggregate(id string, manager *gateway.EventSourcingManager) *UserAggregate {
    return &UserAggregate{
        ID:      id,
        Version: 0,
        manager: manager,
    }
}

func (u *UserAggregate) CreateUser(ctx context.Context, name, email string) error {
    eventRequest := &types.AppendEventRequest{
        StreamID:        u.ID,
        EventType:       "UserCreated",
        EventData:       map[string]interface{}{"name": name, "email": email},
        EventMetadata:   map[string]interface{}{"source": "aggregate"},
        AggregateID:     u.ID,
        AggregateType:   "User",
        ExpectedVersion: &u.Version,
    }

    response, err := u.manager.AppendEvent(ctx, eventRequest)
    if err != nil {
        return err
    }

    // Update aggregate state
    u.Name = name
    u.Email = email
    u.Version = response.Version

    return nil
}

func (u *UserAggregate) LoadFromEvents(ctx context.Context) error {
    request := &types.GetEventsByStreamRequest{
        StreamID: u.ID,
        Limit:    1000,
    }

    response, err := u.manager.GetEventsByStream(ctx, request)
    if err != nil {
        return err
    }

    // Replay events to rebuild aggregate state
    for _, event := range response.Events {
        u.Version = event.Version

        switch event.EventType {
        case "UserCreated":
            if name, ok := event.EventData["name"].(string); ok {
                u.Name = name
            }
            if email, ok := event.EventData["email"].(string); ok {
                u.Email = email
            }
        }
    }

    return nil
}
```

## Configuration

### Manager Configuration

```go
type ManagerConfig struct {
    DefaultProvider   string        `json:"default_provider"`
    RetryAttempts     int           `json:"retry_attempts"`
    RetryDelay        time.Duration `json:"retry_delay"`
    Timeout           time.Duration `json:"timeout"`
    MaxEventSize      int64         `json:"max_event_size"`
    MaxBatchSize      int           `json:"max_batch_size"`
    SnapshotThreshold int64         `json:"snapshot_threshold"`
    RetentionPeriod   time.Duration `json:"retention_period"`
    Compression       bool          `json:"compression"`
    Encryption        bool          `json:"encryption"`
    Metadata          map[string]string `json:"metadata"`
}
```

### PostgreSQL Configuration

```go
type PostgreSQLConfig struct {
    Host            string            `json:"host"`
    Port            int               `json:"port"`
    Database        string            `json:"database"`
    Username        string            `json:"username"`
    Password        string            `json:"password"`
    SSLMode         string            `json:"ssl_mode"`
    MaxConnections  int               `json:"max_connections"`
    MaxIdleConns    int               `json:"max_idle_conns"`
    ConnMaxLifetime time.Duration     `json:"conn_max_lifetime"`
    Metadata        map[string]string `json:"metadata"`
}
```

### Kafka Configuration

```go
type KafkaConfig struct {
    Brokers         []string      `json:"brokers"`
    Topic           string        `json:"topic"`
    GroupID         string        `json:"group_id"`
    Username        string        `json:"username"`
    Password        string        `json:"password"`
    SecurityProtocol string       `json:"security_protocol"`
    CompressionType string        `json:"compression_type"`
    BatchSize       int           `json:"batch_size"`
    BatchTimeout    time.Duration `json:"batch_timeout"`
    ReadTimeout     time.Duration `json:"read_timeout"`
    WriteTimeout    time.Duration `json:"write_timeout"`
    Metadata        map[string]string `json:"metadata"`
}
```

### NATS Configuration

```go
type NATSConfig struct {
    URL             string        `json:"url"`
    Subject         string        `json:"subject"`
    StreamName      string        `json:"stream_name"`
    Username        string        `json:"username"`
    Password        string        `json:"password"`
    Token           string        `json:"token"`
    MaxReconnects   int           `json:"max_reconnects"`
    ReconnectWait   time.Duration `json:"reconnect_wait"`
    Timeout         time.Duration `json:"timeout"`
    RetentionPolicy string        `json:"retention_policy"`
    MaxAge          time.Duration `json:"max_age"`
    MaxBytes        int64         `json:"max_bytes"`
    MaxEvents       int64         `json:"max_events"`
    Metadata        map[string]string `json:"metadata"`
}
```

## Error Handling

The library provides comprehensive error handling with specific error types:

```go
type EventSourcingError struct {
    Code     string `json:"code"`
    Message  string `json:"message"`
    StreamID string `json:"stream_id,omitempty"`
    EventID  string `json:"event_id,omitempty"`
}
```

Common error codes:
- `STREAM_NOT_FOUND`: Stream doesn't exist
- `EVENT_NOT_FOUND`: Event doesn't exist
- `SNAPSHOT_NOT_FOUND`: Snapshot doesn't exist
- `VERSION_MISMATCH`: Expected version doesn't match current version
- `INVALID_STREAM_ID`: Invalid stream identifier
- `INVALID_EVENT_TYPE`: Invalid event type
- `INVALID_EVENT_DATA`: Invalid event data

## Testing

Run the examples to test the library:

```bash
cd examples
go run example.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue on GitHub.
