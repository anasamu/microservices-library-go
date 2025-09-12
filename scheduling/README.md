# Scheduling Library

Library untuk task scheduling yang mendukung berbagai provider dan fitur scheduling yang komprehensif.

## Fitur

- **Multi-Provider Support**: Mendukung berbagai provider scheduling (Cron, Redis, dll)
- **Flexible Scheduling**: Mendukung cron expressions, one-time tasks, dan interval-based scheduling
- **Task Management**: Create, update, cancel, dan monitor tasks
- **Retry Policy**: Konfigurasi retry dengan berbagai backoff strategies
- **Health Monitoring**: Health checks dan metrics untuk semua providers
- **Batch Operations**: Schedule dan cancel multiple tasks sekaligus
- **Extensible**: Mudah untuk menambah provider baru

## Struktur

```
scheduling/
├── providers/             # Scheduling provider implementations
│   ├── cron/              # Cron-based scheduling
│   └── redis/             # Redis-based scheduling
├── test/                  # Test files
│   ├── integration/       # Integration tests
│   ├── unit/              # Unit tests
│   └── mocks/             # Mock providers
├── types/                 # Data structures dan types
├── manager.go             
├── go.mod                 # Main module dependencies
└── README.md              # Dokumentasi library
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
    "time"
    
    "github.com/anasamu/microservices-library-go/scheduling"
    "github.com/anasamu/microservices-library-go/scheduling/providers/cron"
    "github.com/anasamu/microservices-library-go/scheduling/types"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create manager configuration
    config := &gateway.ManagerConfig{
        DefaultProvider: "cron",
        RetryAttempts:   3,
        RetryDelay:      time.Second,
        Timeout:         30 * time.Second,
        FallbackEnabled: true,
    }
    
    // Create scheduling manager
    manager := gateway.NewSchedulingManager(config, logger)
    
    // Create and register cron provider
    cronConfig := &cron.CronConfig{
        Timezone:     "UTC",
        WithSeconds:  false,
        WithLocation: true,
    }
    
    cronProvider, err := cron.NewCronProvider(cronConfig, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    err = manager.RegisterProvider("cron", cronProvider)
    if err != nil {
        log.Fatal(err)
    }
    
    // Connect provider
    ctx := context.Background()
    err = cronProvider.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer cronProvider.Disconnect(ctx)
    
    // Schedule a task
    task := &types.Task{
        ID:          "daily-report",
        Name:        "Daily Report Generation",
        Description: "Generate daily sales report",
        Schedule: &types.Schedule{
            Type:     types.ScheduleTypeCron,
            CronExpr: "0 9 * * *", // Every day at 9 AM
        },
        Handler: &types.TaskHandler{
            Type: types.HandlerTypeHTTP,
            HTTP: &types.HTTPHandler{
                URL:    "http://api.example.com/reports/daily",
                Method: "POST",
                Headers: map[string]string{
                    "Content-Type": "application/json",
                    "Authorization": "Bearer token123",
                },
                Body: `{"report_type": "daily", "date": "{{.Date}}"}`,
            },
        },
        RetryPolicy: &types.RetryPolicy{
            MaxAttempts: 3,
            Delay:       time.Minute,
            Backoff:     types.BackoffTypeExponential,
        },
        Timeout: time.Minute * 5,
        Metadata: map[string]string{
            "department": "sales",
            "priority":   "high",
        },
    }
    
    result, err := manager.ScheduleTask(ctx, task, "cron")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Task scheduled successfully: %s", result.TaskID)
}
```

## Task Types

### 1. Cron Tasks

```go
task := &types.Task{
    ID:   "cron-task",
    Name: "Cron Task",
    Schedule: &types.Schedule{
        Type:     types.ScheduleTypeCron,
        CronExpr: "0 */6 * * *", // Every 6 hours
        Timezone: "UTC",
    },
    Handler: &types.TaskHandler{
        Type: types.HandlerTypeHTTP,
        HTTP: &types.HTTPHandler{
            URL:    "http://api.example.com/process",
            Method: "POST",
        },
    },
}
```

### 2. One-Time Tasks

```go
task := &types.Task{
    ID:   "one-time-task",
    Name: "One-Time Task",
    Schedule: &types.Schedule{
        Type:      types.ScheduleTypeOnce,
        StartTime: &[]time.Time{time.Now().Add(time.Hour)}[0], // Run in 1 hour
    },
    Handler: &types.TaskHandler{
        Type: types.HandlerTypeCommand,
        Command: &types.CommandHandler{
            Command: "python",
            Args:    []string{"script.py"},
            Env: map[string]string{
                "ENV": "production",
            },
        },
    },
}
```

### 3. Interval Tasks

```go
task := &types.Task{
    ID:   "interval-task",
    Name: "Interval Task",
    Schedule: &types.Schedule{
        Type:     types.ScheduleTypeInterval,
        Interval: time.Minute * 30, // Every 30 minutes
    },
    Handler: &types.TaskHandler{
        Type: types.HandlerTypeFunction,
        Function: &types.FunctionHandler{
            FunctionName: "cleanup",
            Parameters: map[string]interface{}{
                "cleanup_temp": true,
                "optimize_db":  true,
            },
        },
    },
}
```

## Task Handlers

### 1. HTTP Handler

```go
handler := &types.TaskHandler{
    Type: types.HandlerTypeHTTP,
    HTTP: &types.HTTPHandler{
        URL:    "http://api.example.com/webhook",
        Method: "POST",
        Headers: map[string]string{
            "Content-Type": "application/json",
            "Authorization": "Bearer token",
        },
        Body:    `{"message": "Hello World"}`,
        Timeout: time.Minute * 5,
    },
}
```

### 2. Command Handler

```go
handler := &types.TaskHandler{
    Type: types.HandlerTypeCommand,
    Command: &types.CommandHandler{
        Command: "curl",
        Args:    []string{"-X", "POST", "http://api.example.com/trigger"},
        Env: map[string]string{
            "API_KEY": "secret",
        },
        WorkDir: "/tmp",
    },
}
```

### 3. Function Handler

```go
handler := &types.TaskHandler{
    Type: types.HandlerTypeFunction,
    Function: &types.FunctionHandler{
        FunctionName: "processData",
        Parameters: map[string]interface{}{
            "batch_size": 100,
            "timeout":    300,
        },
    },
}
```

### 4. Message Handler

```go
handler := &types.TaskHandler{
    Type: types.HandlerTypeMessage,
    Message: &types.MessageHandler{
        Topic:   "notifications",
        Message: `{"type": "alert", "message": "System check"}`,
        Headers: map[string]string{
            "priority": "high",
        },
    },
}
```

## Retry Policy

```go
retryPolicy := &types.RetryPolicy{
    MaxAttempts: 3,
    Delay:       time.Minute,
    Backoff:     types.BackoffTypeExponential,
    MaxDelay:    time.Minute * 10,
}
```

### Backoff Types

- `BackoffTypeFixed`: Fixed delay between retries
- `BackoffTypeLinear`: Linear increase in delay
- `BackoffTypeExponential`: Exponential increase in delay

## Task Management

### List Tasks

```go
filter := &types.TaskFilter{
    Status: []types.TaskStatus{
        types.TaskStatusScheduled,
        types.TaskStatusRunning,
    },
    Tags:  []string{"production"},
    Limit: 10,
}

tasks, err := manager.ListTasks(ctx, filter, "cron")
if err != nil {
    log.Fatal(err)
}

for _, task := range tasks {
    log.Printf("Task: %s - %s (Status: %s)", task.ID, task.Name, task.Status)
}
```

### Update Task

```go
task, err := manager.GetTask(ctx, "task-id", "cron")
if err != nil {
    log.Fatal(err)
}

task.Description = "Updated description"
task.Metadata["updated"] = "true"

err = manager.UpdateTask(ctx, task, "cron")
if err != nil {
    log.Fatal(err)
}
```

### Cancel Task

```go
err := manager.CancelTask(ctx, "task-id", "cron")
if err != nil {
    log.Fatal(err)
}
```

### Batch Operations

```go
// Schedule multiple tasks
tasks := []*types.Task{
    {ID: "task1", Name: "Task 1", /* ... */},
    {ID: "task2", Name: "Task 2", /* ... */},
}

results, err := manager.ScheduleMultiple(ctx, tasks, "cron")
if err != nil {
    log.Fatal(err)
}

// Cancel multiple tasks
taskIDs := []string{"task1", "task2"}
err = manager.CancelMultiple(ctx, taskIDs, "cron")
if err != nil {
    log.Fatal(err)
}
```

## Health Monitoring

### Get Health Status

```go
health, err := manager.GetHealth(ctx)
if err != nil {
    log.Fatal(err)
}

for provider, status := range health {
    log.Printf("Provider %s: %s - %s", provider, status.Status, status.Message)
}
```

### Get Metrics

```go
metrics, err := manager.GetMetrics(ctx)
if err != nil {
    log.Fatal(err)
}

for provider, metric := range metrics {
    log.Printf("Provider %s: %d total tasks, %.2f%% success rate", 
        provider, metric.TotalTasks, metric.SuccessRate)
}
```

## Providers

### Cron Provider

Cron provider menggunakan library `github.com/robfig/cron/v3` untuk scheduling berbasis cron expressions.

```go
cronConfig := &cron.CronConfig{
    Timezone:     "UTC",
    WithSeconds:  false,
    WithLocation: true,
    Metadata: map[string]string{
        "environment": "production",
    },
}

cronProvider, err := cron.NewCronProvider(cronConfig, logger)
```

**Features:**
- Cron expressions
- Timezone support
- Recurring tasks
- Retry policies
- Health monitoring

### Redis Provider

Redis provider menggunakan Redis untuk persistence dan clustering.

```go
redisConfig := &redis.RedisConfig{
    Host:         "localhost",
    Port:         6379,
    Password:     "",
    Database:     0,
    PoolSize:     10,
    MinIdleConns: 5,
    MaxRetries:   3,
    DialTimeout:  time.Second * 5,
    ReadTimeout:  time.Second * 3,
    WriteTimeout: time.Second * 3,
}

redisProvider, err := redis.NewRedisProvider(redisConfig, logger)
```

**Features:**
- Persistence
- Clustering support
- All schedule types
- Batch operations
- Advanced filtering

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

```go
import "github.com/anasamu/microservices-library-go/scheduling/test/mocks"

mockProvider := mocks.NewMockSchedulingProvider("test")
mockProvider.SetConnectError(nil)
mockProvider.SetScheduleTaskError(nil)

manager.RegisterProvider("test", mockProvider)
```

## Error Handling

Library ini menggunakan error handling yang konsisten:

```go
result, err := manager.ScheduleTask(ctx, task, "cron")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "provider not found"):
        // Handle provider not found
    case strings.Contains(err.Error(), "task not found"):
        // Handle task not found
    case strings.Contains(err.Error(), "connection failed"):
        // Handle connection issues
    default:
        // Handle other errors
    }
}
```

## Best Practices

1. **Always use context**: Pass context untuk timeout dan cancellation
2. **Handle errors**: Selalu handle errors dengan proper error checking
3. **Use retry policies**: Configure retry policies untuk tasks yang critical
4. **Monitor health**: Regularly check provider health status
5. **Use batch operations**: Untuk multiple tasks, gunakan batch operations
6. **Set timeouts**: Set appropriate timeouts untuk tasks
7. **Use metadata**: Gunakan metadata untuk tagging dan filtering
8. **Clean up**: Always disconnect providers ketika aplikasi shutdown

## Examples

Lihat file `gateway/example.go` untuk contoh penggunaan yang lebih lengkap.

## Contributing

1. Fork repository
2. Create feature branch
3. Add tests untuk fitur baru
4. Ensure all tests pass
5. Submit pull request

## License

MIT License
