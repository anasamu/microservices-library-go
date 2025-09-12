# Backup Microservice Library

A comprehensive Go library for backup and restore operations with support for multiple storage providers including AWS S3, Google Cloud Storage, and local file system.

## Features

- **Multiple Storage Providers**: Support for S3, GCS, and local file system
- **Unified Interface**: Consistent API across all providers
- **Metadata Management**: Rich metadata support with tags and descriptions
- **Health Checks**: Built-in health monitoring for all providers
- **Comprehensive Testing**: Unit tests, integration tests, and mocks
- **Context Support**: Full context.Context support for cancellation and timeouts
- **Error Handling**: Detailed error messages and proper error propagation

## Architecture

```
backup/
├── gateway/                    # Core backup gateway
│   ├── manager.go             # Backup and restore manager
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Backup provider implementations
│   ├── s3/                    # S3-based backup
│   │   ├── provider.go        # S3 implementation
│   │   └── go.mod             # S3 dependencies
│   ├── gcs/                   # GCS-based backup
│   │   ├── provider.go        # GCS implementation
│   │   └── go.mod             # GCS dependencies
│   └── local/                 # Local file-based backup
│       ├── provider.go        # Local implementation
│       └── go.mod             # Local dependencies
├── test/                      # Test files
│   ├── integration/           # Integration tests
│   ├── unit/                  # Unit tests
│   └── mocks/                 # Mock providers
├── types/                     # Common types and interfaces
│   └── provider.go            # Provider interface definitions
├── go.mod                     # Main module dependencies
└── README.md                  # This documentation
```

## Quick Start

### Installation

```bash
go get github.com/anasamu/microservices-library-go/backup
```

### Basic Usage

```go
package main

import (
    "context"
    "strings"
    
    "github.com/anasamu/microservices-library-go/backup"
    "github.com/anasamu/microservices-library-go/backup/providers/local"
    "github.com/anasamu/microservices-library-go/backup/types"
)

func main() {
    // Create backup manager
    manager := gateway.NewBackupManager()
    
    // Set up local provider
    localProvider := local.NewLocalProvider("/tmp/backups")
    manager.SetProvider(localProvider)
    
    // Create a backup
    ctx := context.Background()
    data := strings.NewReader("Hello, World!")
    opts := &types.BackupOptions{
        Compression: true,
        Tags: map[string]string{
            "environment": "development",
        },
        Description: "My first backup",
    }
    
    metadata, err := manager.CreateBackup(ctx, "hello-backup", data, opts)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Backup created: %s\n", metadata.ID)
}
```

## Providers

### Local File Provider

The local file provider stores backups in the local file system.

```go
import "github.com/anasamu/microservices-library-go/backup/providers/local"

// Create local provider
provider := local.NewLocalProvider("/path/to/backup/directory")
manager.SetProvider(provider)
```

**Features:**
- File-based storage
- JSON metadata files
- Automatic directory creation
- Health checks for directory access

### AWS S3 Provider

The S3 provider stores backups in Amazon S3 or S3-compatible services.

```go
import "github.com/anasamu/microservices-library-go/backup/providers/s3"

// Create S3 provider
config := &s3.S3Config{
    Bucket: "my-backup-bucket",
    Prefix: "backups/",
    Region: "us-west-2",
    // Endpoint: "https://s3.amazonaws.com", // Optional for custom endpoints
}
provider, err := s3.NewS3Provider(config)
if err != nil {
    panic(err)
}
manager.SetProvider(provider)
```

**Features:**
- S3 and S3-compatible storage
- Metadata stored as separate objects
- Configurable prefixes and regions
- Support for custom endpoints

**Prerequisites:**
- AWS credentials configured (via environment variables, IAM roles, or AWS config)
- S3 bucket with appropriate permissions

### Google Cloud Storage Provider

The GCS provider stores backups in Google Cloud Storage.

```go
import "github.com/anasamu/microservices-library-go/backup/providers/gcs"

// Create GCS provider
config := &gcs.GCSConfig{
    Bucket: "my-backup-bucket",
    Prefix: "backups/",
}
provider, err := gcs.NewGCSProvider(ctx, config)
if err != nil {
    panic(err)
}
manager.SetProvider(provider)
```

**Features:**
- Google Cloud Storage integration
- Metadata stored as separate objects
- Configurable prefixes
- Automatic client creation

**Prerequisites:**
- Google Cloud credentials configured
- GCS bucket with appropriate permissions

## API Reference

### BackupManager

The main interface for backup operations.

```go
type Manager interface {
    SetProvider(provider Provider)
    GetProvider() Provider
    CreateBackup(ctx context.Context, name string, data io.Reader, opts *BackupOptions) (*BackupMetadata, error)
    RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *RestoreOptions) error
    ListBackups(ctx context.Context) ([]*BackupMetadata, error)
    GetBackup(ctx context.Context, backupID string) (*BackupMetadata, error)
    DeleteBackup(ctx context.Context, backupID string) error
    HealthCheck(ctx context.Context) error
}
```

### Types

#### BackupMetadata

```go
type BackupMetadata struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Size        int64             `json:"size"`
    CreatedAt   time.Time         `json:"created_at"`
    Tags        map[string]string `json:"tags,omitempty"`
    Description string            `json:"description,omitempty"`
}
```

#### BackupOptions

```go
type BackupOptions struct {
    Compression bool              `json:"compression"`
    Encryption  bool              `json:"encryption"`
    Tags        map[string]string `json:"tags,omitempty"`
    Description string            `json:"description,omitempty"`
}
```

#### RestoreOptions

```go
type RestoreOptions struct {
    Overwrite bool `json:"overwrite"`
}
```

## Examples

### Complete Backup and Restore Workflow

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    
    "github.com/anasamu/microservices-library-go/backup"
    "github.com/anasamu/microservices-library-go/backup/providers/local"
    "github.com/anasamu/microservices-library-go/backup/types"
)

func main() {
    // Setup
    manager := gateway.NewBackupManager()
    provider := local.NewLocalProvider("/tmp/backups")
    manager.SetProvider(provider)
    
    ctx := context.Background()
    
    // 1. Create backup
    data := strings.NewReader("Important application data")
    opts := &types.BackupOptions{
        Compression: true,
        Tags: map[string]string{
            "app": "myapp",
            "env": "prod",
        },
        Description: "Production data backup",
    }
    
    metadata, err := manager.CreateBackup(ctx, "app-backup", data, opts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Backup created: %s\n", metadata.ID)
    
    // 2. List all backups
    backups, err := manager.ListBackups(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Total backups: %d\n", len(backups))
    
    // 3. Restore backup
    var restoredData strings.Builder
    restoreOpts := &types.RestoreOptions{Overwrite: true}
    
    err = manager.RestoreBackup(ctx, metadata.ID, &restoredData, restoreOpts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Restored data: %s\n", restoredData.String())
    
    // 4. Health check
    err = manager.HealthCheck(ctx)
    if err != nil {
        log.Fatal("Health check failed:", err)
    }
    
    fmt.Println("Backup system is healthy!")
}
```

### Using S3 Provider

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/backup"
    "github.com/anasamu/microservices-library-go/backup/providers/s3"
)

func main() {
    // S3 configuration
    config := &s3.S3Config{
        Bucket: "my-backup-bucket",
        Prefix: "app-backups/",
        Region: "us-west-2",
    }
    
    provider, err := s3.NewS3Provider(config)
    if err != nil {
        log.Fatal(err)
    }
    
    manager := gateway.NewBackupManager()
    manager.SetProvider(provider)
    
    // Use manager as usual...
}
```

## Testing

The library includes comprehensive testing support:

### Unit Tests

```bash
cd backup/test/unit
go test -v
```

### Integration Tests

```bash
cd backup/test/integration
go test -v
```

### Mock Provider

For testing your application without actual storage:

```go
import "github.com/anasamu/microservices-library-go/backup/test/mocks"

mockProvider := mocks.NewMockProvider()
manager.SetProvider(mockProvider)
```

## Error Handling

The library provides detailed error messages for common scenarios:

- **Provider not set**: "no backup provider set"
- **Backup not found**: "backup not found: {backupID}"
- **Storage errors**: Provider-specific error messages
- **Configuration errors**: Detailed validation messages

## Best Practices

1. **Always use context**: Pass context for cancellation and timeouts
2. **Handle errors**: Check and handle all returned errors
3. **Health checks**: Regularly check provider health
4. **Resource cleanup**: Close readers/writers properly
5. **Testing**: Use mock providers for unit tests

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the examples in the `gateway/example.go` file
- Review the test files for usage patterns
