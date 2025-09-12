# Storage Gateway Library

A comprehensive, modular, and production-ready storage library for Go microservices. This library provides a unified interface for multiple cloud storage providers including AWS S3, Google Cloud Storage, Azure Blob Storage, and MinIO.

## 🚀 Features

### 🔧 Multi-Provider Support
- **AWS S3**: Full S3 API compatibility with presigned URLs, versioning, and encryption
- **Google Cloud Storage**: Complete GCS integration with service account authentication
- **Azure Blob Storage**: Full Azure Blob Storage support with SAS URLs and managed identity
- **MinIO**: S3-compatible object storage for on-premises deployments

### 📁 Core Operations
- **Object Management**: Upload, download, delete, copy, and move objects
- **Batch Operations**: Delete multiple objects in a single operation
- **Object Listing**: List objects with filtering, pagination, and prefix support
- **Object Information**: Get detailed metadata and properties
- **Existence Checks**: Verify object and bucket existence

### 🔗 URL Generation
- **Presigned URLs**: Generate secure, time-limited URLs for direct access
- **Public URLs**: Create public access URLs for objects
- **Custom Expiration**: Configurable URL expiration times
- **Method Support**: GET, PUT, DELETE operations

### 🪣 Bucket Management
- **Bucket Operations**: Create, delete, and list buckets/containers
- **Bucket Policies**: Set and retrieve bucket policies
- **Region Support**: Multi-region bucket creation and management
- **ACL Management**: Access control list configuration

### 🔒 Security Features
- **Encryption**: Server-side encryption support
- **Access Control**: Fine-grained permissions and ACLs
- **Authentication**: Multiple authentication methods per provider
- **Metadata**: Custom metadata and tagging support

### 🏥 Health & Monitoring
- **Health Checks**: Provider health monitoring
- **Retry Logic**: Configurable retry attempts with backoff
- **Error Handling**: Comprehensive error reporting
- **Logging**: Structured logging with context

## 📁 Project Structure

```
libs/storage/
├── providers/                 # Storage provider implementations
│   ├── s3/                    # AWS S3 provider
│   │   ├── provider.go        # S3 implementation
│   │   └── go.mod             # S3 dependencies
│   ├── gcs/                   # Google Cloud Storage provider
│   │   ├── provider.go        # GCS implementation
│   │   └── go.mod             # GCS dependencies
│   ├── azure/                 # Azure Blob Storage provider
│   │   ├── provider.go        # Azure implementation
│   │   └── go.mod             # Azure dependencies
│   └── minio/                 # MinIO provider
│       ├── provider.go        # MinIO implementation
│       └── go.mod             # MinIO dependencies
├── go.mod                     # Main module dependencies
├── manager.go                 # core
└── README.md                  # This file
```

## 🛠️ Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Basic Installation

```bash
# Clone the repository
git clone https://github.com/anasamu/microservices-library-go.git
cd microservices-library-go/libs/storage

# Install dependencies
go mod tidy
```

### Using Specific Providers

```bash
# For S3 support
go get github.com/anasamu/microservices-library-go/libs/storage/providers/s3

# For Google Cloud Storage support
go get github.com/anasamu/microservices-library-go/libs/storage/providers/gcs

# For Azure Blob Storage support
go get github.com/anasamu/microservices-library-go/libs/storage/providers/azure

# For MinIO support
go get github.com/anasamu/microservices-library-go/libs/storage/providers/minio
```

## 📖 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "strings"
    
    "github.com/anasamu/microservices-library-go/libs/storage"
    "github.com/anasamu/microservices-library-go/libs/storage/providers/s3"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create storage manager
    config := gateway.DefaultManagerConfig()
    config.DefaultProvider = "s3"
    config.MaxFileSize = 100 * 1024 * 1024 // 100MB
    
    storageManager := gateway.NewStorageManager(config, logger)
    
    // Register S3 provider
    s3Provider := s3.NewProvider(logger)
    s3Config := map[string]interface{}{
        "region":            "us-east-1",
        "access_key_id":     "your_aws_access_key_id",
        "secret_access_key": "your_aws_secret_access_key",
    }
    
    if err := s3Provider.Configure(s3Config); err != nil {
        log.Fatal(err)
    }
    
    storageManager.RegisterProvider(s3Provider)
    
    // Use the storage manager...
}
```

### Upload a File

```go
// Upload a file
content := strings.NewReader("Hello, World!")
request := &gateway.PutObjectRequest{
    Bucket:      "my-bucket",
    Key:         "path/to/file.txt",
    Content:     content,
    Size:        int64(content.Len()),
    ContentType: "text/plain",
    Metadata: map[string]string{
        "uploaded-by": "my-app",
        "environment": "production",
    },
}

response, err := storageManager.PutObject(ctx, "s3", request)
if err != nil {
    log.Fatal(err)
}

log.Printf("File uploaded: %s", response.Key)
```

### Download a File

```go
// Download a file
request := &gateway.GetObjectRequest{
    Bucket: "my-bucket",
    Key:    "path/to/file.txt",
}

response, err := storageManager.GetObject(ctx, "s3", request)
if err != nil {
    log.Fatal(err)
}
defer response.Content.Close()

// Read the content
data, err := io.ReadAll(response.Content)
if err != nil {
    log.Fatal(err)
}

log.Printf("Downloaded %d bytes", len(data))
```

### List Objects

```go
// List objects
request := &gateway.ListObjectsRequest{
    Bucket:  "my-bucket",
    Prefix:  "path/to/",
    MaxKeys: 100,
}

response, err := storageManager.ListObjects(ctx, "s3", request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Found %d objects:", len(response.Objects))
for _, obj := range response.Objects {
    log.Printf("  - %s (%d bytes)", obj.Key, obj.Size)
}
```

### Generate Presigned URLs

```go
// Generate presigned URL for upload
request := &gateway.PresignedURLRequest{
    Bucket:  "my-bucket",
    Key:     "uploads/user-file.txt",
    Method:  "PUT",
    Expires: 1 * time.Hour,
    Headers: map[string]string{
        "Content-Type": "text/plain",
    },
}

url, err := storageManager.GeneratePresignedURL(ctx, "s3", request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Presigned URL: %s", url)
```

### Copy and Move Objects

```go
// Copy object
copyRequest := &gateway.CopyObjectRequest{
    SourceBucket: "source-bucket",
    SourceKey:    "source/file.txt",
    DestBucket:   "dest-bucket",
    DestKey:      "dest/file.txt",
}

copyResponse, err := storageManager.CopyObject(ctx, "s3", copyRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Object copied: %s", copyResponse.Key)

// Move object
moveRequest := &gateway.MoveObjectRequest{
    SourceBucket: "source-bucket",
    SourceKey:    "source/file.txt",
    DestBucket:   "dest-bucket",
    DestKey:      "dest/file.txt",
}

moveResponse, err := storageManager.MoveObject(ctx, "s3", moveRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Object moved: %s", moveResponse.Key)
```

## 🔧 Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Storage Manager Configuration
export STORAGE_DEFAULT_PROVIDER="s3"
export STORAGE_MAX_FILE_SIZE="104857600"  # 100MB
export STORAGE_RETRY_ATTEMPTS="3"
export STORAGE_RETRY_DELAY="5s"
export STORAGE_TIMEOUT="30s"

# S3 Configuration
export S3_REGION="us-east-1"
export S3_ACCESS_KEY_ID="your_access_key"
export S3_SECRET_ACCESS_KEY="your_secret_key"
export S3_ENDPOINT=""  # Leave empty for AWS S3

# Google Cloud Storage Configuration
export GCS_PROJECT_ID="your-project-id"
export GCS_CREDENTIALS_PATH="/path/to/service-account.json"

# Azure Blob Storage Configuration
export AZURE_ACCOUNT_NAME="your_storage_account"
export AZURE_ACCOUNT_KEY="your_storage_key"

# MinIO Configuration
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY_ID="minioadmin"
export MINIO_SECRET_ACCESS_KEY="minioadmin"
export MINIO_USE_SSL="false"
```

### Configuration Files

You can also use configuration files:

```json
{
  "storage": {
    "default_provider": "s3",
    "max_file_size": 104857600,
    "retry_attempts": 3,
    "retry_delay": "5s",
    "timeout": "30s",
    "allowed_types": ["image/*", "application/pdf", "text/*"]
  },
  "providers": {
    "s3": {
      "region": "us-east-1",
      "access_key_id": "your_access_key",
      "secret_access_key": "your_secret_key"
    },
    "gcs": {
      "project_id": "your-project-id",
      "credentials_path": "/path/to/service-account.json"
    },
    "azure": {
      "account_name": "your_storage_account",
      "account_key": "your_storage_key"
    },
    "minio": {
      "endpoint": "localhost:9000",
      "access_key_id": "minioadmin",
      "secret_access_key": "minioadmin",
      "use_ssl": false
    }
  }
}
```

## 🧪 Testing

Run tests for all modules:

```bash
# Run all tests
go test ./...

# Run tests for specific provider
go test ./providers/s3/...
go test ./providers/gcs/...
go test ./providers/azure/...
go test ./providers/minio/...

# Run gateway tests
go test ./...
```

## 📚 API Documentation

### Storage Manager API

- `NewStorageManager(config, logger)` - Create storage manager
- `RegisterProvider(provider)` - Register a storage provider
- `PutObject(ctx, provider, request)` - Upload an object
- `GetObject(ctx, provider, request)` - Download an object
- `DeleteObject(ctx, provider, request)` - Delete an object
- `ListObjects(ctx, provider, request)` - List objects
- `CopyObject(ctx, provider, request)` - Copy an object
- `MoveObject(ctx, provider, request)` - Move an object
- `GeneratePresignedURL(ctx, provider, request)` - Generate presigned URL
- `HealthCheck(ctx)` - Check provider health

### Provider Interface

All providers implement the `StorageProvider` interface:

```go
type StorageProvider interface {
    GetName() string
    GetSupportedFeatures() []StorageFeature
    Configure(config map[string]interface{}) error
    IsConfigured() bool
    PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error)
    GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error)
    DeleteObject(ctx context.Context, request *DeleteObjectRequest) error
    // ... other methods
}
```

### Supported Features

- `FeaturePresignedURLs` - Presigned URL generation
- `FeaturePublicURLs` - Public URL generation
- `FeatureMultipart` - Multipart upload support
- `FeatureVersioning` - Object versioning
- `FeatureEncryption` - Server-side encryption
- `FeatureLifecycle` - Lifecycle management
- `FeatureCORS` - Cross-origin resource sharing
- `FeatureCDN` - Content delivery network integration

## 🔒 Security Considerations

### Authentication

Each provider supports multiple authentication methods:

- **AWS S3**: Access keys, IAM roles, instance profiles
- **Google Cloud Storage**: Service accounts, workload identity
- **Azure Blob Storage**: Account keys, service principals, managed identity
- **MinIO**: Access keys, LDAP, OIDC

### Encryption

- **Server-side encryption** for all providers
- **Client-side encryption** support
- **Encryption key management** integration
- **TLS/SSL** for all communications

### Access Control

- **Bucket policies** and ACLs
- **IAM integration** for cloud providers
- **Presigned URLs** with time limits
- **Metadata-based access control**

## 🚀 Performance

### Optimization Features

- **Concurrent uploads** and downloads
- **Multipart upload** for large files
- **Connection pooling** and reuse
- **Retry logic** with exponential backoff
- **Compression** support

### Monitoring

- **Health checks** for all providers
- **Metrics collection** for operations
- **Error tracking** and reporting
- **Performance monitoring**

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our Discord](https://discord.gg/example)
- 📖 Documentation: [Full Documentation](https://docs.example.com)
- 🐛 Issues: [GitHub Issues](https://github.com/anasamu/microservices-library-go/issues)

## 🙏 Acknowledgments

- [AWS S3](https://aws.amazon.com/s3/) for object storage
- [Google Cloud Storage](https://cloud.google.com/storage) for cloud storage
- [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/) for blob storage
- [MinIO](https://min.io/) for S3-compatible storage
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go-v2) for S3 integration
- [Google Cloud Go](https://github.com/googleapis/google-cloud-go) for GCS integration
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) for Azure integration

---

Made with ❤️ for the Go microservices community
