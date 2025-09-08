# Microservices Library Go

Library Go yang komprehensif untuk pengembangan microservices dengan dukungan berbagai provider dan layanan terintegrasi.

## 📋 Daftar Isi

- [Gambaran Umum](#gambaran-umum)
- [Arsitektur](#arsitektur)
- [Fitur Utama](#fitur-utama)
- [Instalasi](#instalasi)
- [Penggunaan](#penggunaan)
- [Komponen Library](#komponen-library)
- [Contoh Implementasi](#contoh-implementasi)
- [Konfigurasi](#konfigurasi)
- [Monitoring & Observability](#monitoring--observability)
- [Kontribusi](#kontribusi)
- [Lisensi](#lisensi)

## 🎯 Gambaran Umum

Microservices Library Go adalah kumpulan library yang dirancang untuk mempermudah pengembangan aplikasi microservices dengan arsitektur yang modular dan dapat diperluas. Library ini menyediakan abstraksi yang konsisten untuk berbagai layanan cloud dan provider eksternal.

### Keunggulan

- **Modular Design**: Setiap komponen dapat digunakan secara independen
- **Provider Agnostic**: Mendukung multiple provider untuk setiap layanan
- **Fallback Mechanism**: Automatic failover antar provider
- **Comprehensive Logging**: Built-in logging dan monitoring
- **Type Safety**: Strong typing dengan Go generics
- **Production Ready**: Siap untuk production dengan error handling yang robust

## 🏗️ Arsitektur

Library ini mengikuti pola **Gateway Pattern** dengan struktur sebagai berikut:

```
microservices-library-go/
├── ai/                    # AI/ML Services
├── auth/                  # Authentication & Authorization
├── cache/                 # Caching Services
├── communication/         # Communication Protocols
├── config/                # Configuration Management
├── database/              # Database Services
├── logging/               # Logging Services
├── messaging/             # Message Queue Services
├── monitoring/            # Monitoring & Observability
├── payment/               # Payment Processing
├── ratelimit/             # Rate Limiting
├── resilience/            # Circuit Breaker & Resilience
├── storage/               # Object Storage Services
├── tracing/               # Distributed Tracing
└── utils/                 # Utility Functions
```

### Pola Arsitektur

Setiap modul mengikuti pola yang konsisten:

1. **Gateway Manager**: Mengelola multiple provider
2. **Provider Interface**: Abstraksi untuk implementasi spesifik
3. **Types Package**: Definisi tipe data dan struktur
4. **Examples**: Contoh implementasi dan penggunaan

## ✨ Fitur Utama

### 🤖 AI Services
- **Multi-Provider Support**: OpenAI, Anthropic, Google, DeepSeek, X.AI
- **Unified Interface**: Chat, Text Generation, Embeddings
- **Fallback Mechanism**: Automatic provider switching
- **Usage Statistics**: Token tracking dan performance metrics

### 🔐 Authentication & Authorization
- **Multiple Auth Methods**: JWT, OAuth2, 2FA, LDAP, SAML
- **Authorization Models**: RBAC, ABAC, ACL
- **Session Management**: Token refresh, revocation
- **Security Features**: Rate limiting, brute force protection

### 💾 Storage Services
- **Cloud Storage**: AWS S3, Google Cloud Storage, Azure Blob
- **Self-Hosted**: MinIO
- **Advanced Features**: Presigned URLs, multipart upload, versioning
- **File Management**: Copy, move, metadata operations

### 🗄️ Database Services
- **SQL Databases**: PostgreSQL, MySQL
- **NoSQL Databases**: MongoDB, Redis, Elasticsearch
- **Connection Pooling**: Built-in connection management
- **Migration Support**: Database schema management

### 📨 Messaging Services
- **Message Queues**: Kafka, RabbitMQ, NATS, AWS SQS
- **Event Streaming**: Real-time event processing
- **Dead Letter Queues**: Error handling dan retry logic

### 🔧 Configuration Management
- **Multiple Sources**: Environment variables, files, Consul, Vault
- **Hot Reloading**: Dynamic configuration updates
- **Validation**: Schema validation untuk konfigurasi

### 📊 Monitoring & Observability
- **Metrics**: Prometheus integration
- **Tracing**: Jaeger distributed tracing
- **Logging**: Structured logging dengan Elasticsearch
- **Health Checks**: Comprehensive health monitoring

## 🚀 Instalasi

### Prerequisites
- Go 1.21 atau lebih baru
- Git

### Install Dependencies

```bash
# Clone repository
git clone https://github.com/anasamu/microservices-library-go.git
cd microservices-library-go

# Install dependencies untuk semua modul
./tidy-all.sh
```

### Individual Module Installation

```bash
# AI Services
cd ai && go mod tidy

# Authentication
cd auth && go mod tidy

# Storage
cd storage && go mod tidy

# Database
cd database && go mod tidy
```

## 💡 Penggunaan

### AI Services

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/anasamu/microservices-library-go/ai/gateway"
    "github.com/anasamu/microservices-library-go/ai/types"
)

func main() {
    // Create AI manager
    manager := gateway.NewAIManager()
    
    // Add OpenAI provider
    config := &types.ProviderConfig{
        Name:         "openai",
        APIKey:       "your-api-key",
        DefaultModel: "gpt-4",
        Timeout:      30 * time.Second,
    }
    
    if err := manager.AddProvider(config); err != nil {
        log.Fatal(err)
    }
    
    // Chat request
    chatReq := &types.ChatRequest{
        Messages: []types.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
        Model:       "gpt-4",
        Temperature: 0.7,
    }
    
    ctx := context.Background()
    resp, err := manager.Chat(ctx, "openai", chatReq)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", resp.Choices[0].Message.Content)
}
```

### Authentication Services

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/auth/gateway"
    "github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
    "github.com/anasamu/microservices-library-go/auth/types"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create auth manager
    authManager := gateway.NewAuthManager(gateway.DefaultManagerConfig(), logger)
    
    // Register JWT provider
    jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
    authManager.RegisterProvider(jwtProvider)
    
    // Authenticate user
    authReq := &types.AuthRequest{
        Username: "john_doe",
        Password: "secure_password",
    }
    
    ctx := context.Background()
    resp, err := authManager.Authenticate(ctx, "jwt", authReq)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Authentication successful: %+v", resp)
}
```

### Storage Services

```go
package main

import (
    "context"
    "log"
    "strings"
    
    "github.com/anasamu/microservices-library-go/storage/gateway"
    "github.com/anasamu/microservices-library-go/storage/providers/s3"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create storage manager
    storageManager := gateway.NewStorageManager(gateway.DefaultManagerConfig(), logger)
    
    // Register S3 provider
    s3Provider := s3.NewS3Provider(s3Config, logger)
    storageManager.RegisterProvider(s3Provider)
    
    // Upload file
    content := strings.NewReader("Hello, World!")
    putReq := &gateway.PutObjectRequest{
        Bucket:      "my-bucket",
        Key:         "test/file.txt",
        Content:     content,
        Size:        13,
        ContentType: "text/plain",
    }
    
    ctx := context.Background()
    resp, err := storageManager.PutObject(ctx, "s3", putReq)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("File uploaded: %+v", resp)
}
```

## 🧩 Komponen Library

### AI Services (`ai/`)
- **Providers**: OpenAI, Anthropic, Google, DeepSeek, X.AI
- **Features**: Chat, Text Generation, Embeddings, Model Information
- **Fallback**: Automatic provider switching on failure

### Authentication (`auth/`)
- **Authentication**: JWT, OAuth2, 2FA, LDAP, SAML
- **Authorization**: RBAC, ABAC, ACL
- **Security**: Rate limiting, session management, audit logging

### Cache (`cache/`)
- **Providers**: Redis, Memcached, In-Memory
- **Features**: TTL, eviction policies, clustering

### Communication (`communication/`)
- **Protocols**: HTTP, gRPC, WebSocket, GraphQL, SSE, QUIC
- **Features**: Load balancing, circuit breaker, retry logic

### Configuration (`config/`)
- **Sources**: Environment, Files, Consul, Vault
- **Features**: Hot reloading, validation, encryption

### Database (`database/`)
- **SQL**: PostgreSQL, MySQL
- **NoSQL**: MongoDB, Redis, Elasticsearch
- **Features**: Connection pooling, migrations, transactions

### Logging (`logging/`)
- **Outputs**: Console, File, Elasticsearch
- **Features**: Structured logging, log levels, correlation IDs

### Messaging (`messaging/`)
- **Queues**: Kafka, RabbitMQ, NATS, AWS SQS
- **Features**: Event streaming, dead letter queues, partitioning

### Monitoring (`monitoring/`)
- **Metrics**: Prometheus
- **Tracing**: Jaeger
- **Logging**: Elasticsearch
- **Features**: Health checks, alerting, dashboards

### Payment (`payment/`)
- **Providers**: Stripe, PayPal, Midtrans, Xendit
- **Features**: Payment processing, webhooks, refunds

### Storage (`storage/`)
- **Cloud**: AWS S3, Google Cloud Storage, Azure Blob
- **Self-hosted**: MinIO
- **Features**: Presigned URLs, multipart upload, versioning

## 📝 Contoh Implementasi

### Microservice dengan Multiple Services

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/anasamu/microservices-library-go/ai/gateway"
    "github.com/anasamu/microservices-library-go/auth/gateway"
    "github.com/anasamu/microservices-library-go/storage/gateway"
)

func main() {
    // Initialize services
    aiManager := ai_gateway.NewAIManager()
    authManager := auth_gateway.NewAuthManager(auth_gateway.DefaultManagerConfig(), logger)
    storageManager := storage_gateway.NewStorageManager(storage_gateway.DefaultManagerConfig(), logger)
    
    // Setup HTTP routes
    http.HandleFunc("/chat", chatHandler(aiManager))
    http.HandleFunc("/upload", uploadHandler(storageManager))
    http.HandleFunc("/login", loginHandler(authManager))
    
    log.Println("Starting microservice on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Health Check Endpoint

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // Check all services
    aiHealth, _ := aiManager.HealthCheck(ctx)
    authHealth := authManager.HealthCheck(ctx)
    storageHealth := storageManager.HealthCheck(ctx)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    fmt.Fprintf(w, `{
        "status": "healthy",
        "services": {
            "ai": %v,
            "auth": %v,
            "storage": %v
        },
        "timestamp": "%s"
    }`, aiHealth, authHealth, storageHealth, time.Now().Format(time.RFC3339))
}
```

## ⚙️ Konfigurasi

### Environment Variables

```bash
# AI Services
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
GOOGLE_API_KEY=your-google-key

# Authentication
JWT_SECRET=your-jwt-secret
OAUTH_CLIENT_ID=your-oauth-client-id
OAUTH_CLIENT_SECRET=your-oauth-client-secret

# Storage
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
AWS_REGION=us-east-1

# Database
DATABASE_URL=postgres://user:pass@localhost/db
REDIS_URL=redis://localhost:6379

# Monitoring
PROMETHEUS_ENDPOINT=http://localhost:9090
JAEGER_ENDPOINT=http://localhost:14268
```

### Configuration Files

```yaml
# config.yaml
ai:
  providers:
    openai:
      api_key: "${OPENAI_API_KEY}"
      default_model: "gpt-4"
      timeout: 30s
    anthropic:
      api_key: "${ANTHROPIC_API_KEY}"
      default_model: "claude-3-sonnet"

auth:
  providers:
    jwt:
      secret: "${JWT_SECRET}"
      expiration: 24h
    oauth:
      client_id: "${OAUTH_CLIENT_ID}"
      client_secret: "${OAUTH_CLIENT_SECRET}"

storage:
  providers:
    s3:
      access_key: "${AWS_ACCESS_KEY_ID}"
      secret_key: "${AWS_SECRET_ACCESS_KEY}"
      region: "${AWS_REGION}"
```

## 📊 Monitoring & Observability

### Metrics

Library ini menyediakan metrics yang komprehensif:

- **Request Count**: Total requests per provider
- **Response Time**: Latency percentiles
- **Error Rate**: Success/failure ratios
- **Token Usage**: AI token consumption
- **Storage Operations**: Upload/download metrics

### Health Checks

```go
// Health check untuk semua services
func healthCheck() map[string]interface{} {
    ctx := context.Background()
    
    return map[string]interface{}{
        "ai":      aiManager.HealthCheck(ctx),
        "auth":    authManager.HealthCheck(ctx),
        "storage": storageManager.HealthCheck(ctx),
        "database": dbManager.HealthCheck(ctx),
    }
}
```

### Logging

Structured logging dengan correlation IDs:

```go
logger.WithFields(logrus.Fields{
    "service": "user-service",
    "operation": "create_user",
    "user_id": userID,
    "correlation_id": correlationID,
}).Info("User created successfully")
```

## 🤝 Kontribusi

Kontribusi sangat diterima! Silakan ikuti langkah-langkah berikut:

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit perubahan (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

### Development Guidelines

- Ikuti Go coding standards
- Tambahkan tests untuk fitur baru
- Update dokumentasi
- Pastikan semua tests pass

### Adding New Providers

Untuk menambahkan provider baru:

1. Buat implementasi interface di `providers/`
2. Tambahkan konfigurasi di gateway manager
3. Update examples dan dokumentasi
4. Tambahkan tests

## 📄 Lisensi

Project ini dilisensikan di bawah MIT License - lihat file [LICENSE](LICENSE) untuk detail.

## 🙏 Acknowledgments

- Terima kasih kepada semua kontributor
- Inspirasi dari berbagai open source projects
- Community feedback dan suggestions

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/anasamu/microservices-library-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/anasamu/microservices-library-go/discussions)
- **Documentation**: [Wiki](https://github.com/anasamu/microservices-library-go/wiki)

---

**Happy Coding! 🚀**
