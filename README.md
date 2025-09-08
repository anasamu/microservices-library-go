# Microservices Library for Go

A comprehensive, modular, and production-ready library for building microservices in Go. This library provides essential components for authentication, communication, payment processing, service discovery, health checking, and more.

## 🚀 Features

### 🔐 Authentication & Authorization
- **JWT Management**: Token generation, validation, and refresh
- **OAuth2 Support**: Multiple providers (Google, GitHub, etc.)
- **RBAC (Role-Based Access Control)**: Role and permission management
- **ACL (Access Control List)**: Fine-grained access control
- **ABAC (Attribute-Based Access Control)**: Context-aware authorization
- **Dynamic Policy Engine**: Flexible policy evaluation
- **Comprehensive Audit Logging**: Track all authentication events

### 💳 Payment Gateway
- **Multi-Provider Support**: Stripe, PayPal, Xendit, Midtrans
- **Modular Architecture**: Easy to add new payment providers
- **Webhook Handling**: Secure webhook processing
- **Refund Management**: Full refund lifecycle support
- **Currency Support**: Multiple currencies per provider

### 🌐 Communication
- **HTTP Server**: Production-ready HTTP server with middleware
- **WebSocket Support**: Real-time bidirectional communication
- **gRPC**: High-performance RPC communication
- **GraphQL**: Flexible API querying (existing implementation)

### 🔍 Service Discovery & Registry
- **Service Registration**: Automatic service discovery
- **Health Monitoring**: Real-time health checks
- **Load Balancing**: Service instance management
- **Metadata Support**: Rich service metadata

### 🏥 Health Checking
- **Multiple Check Types**: HTTP, Database, Custom checks
- **Concurrent Execution**: Parallel health check execution
- **Detailed Reporting**: Comprehensive health status reporting
- **Configurable Timeouts**: Flexible timeout management

### ⚙️ Configuration Management
- **Multiple Sources**: File, Environment, Custom sources
- **Hot Reloading**: Runtime configuration updates
- **Type Safety**: Strongly typed configuration access
- **Validation**: Schema-based configuration validation

### 🚨 Error Handling
- **Structured Errors**: Consistent error format across services
- **Error Codes**: Standardized error categorization
- **Severity Levels**: Error severity classification
- **Stack Traces**: Detailed error context

## 📁 Project Structure

```
microservices-library-go/
├── auth/                          # Authentication & Authorization
│   ├── jwt.go                     # JWT token management
│   ├── oauth2.go                  # OAuth2 provider support
│   ├── rbac.go                    # Role-based access control
│   ├── acl.go                     # Access control lists
│   ├── abac.go                    # Attribute-based access control
│   ├── policy_engine.go           # Dynamic policy evaluation
│   ├── middleware.go              # HTTP/gRPC middleware
│   ├── audit.go                   # Audit logging
│   ├── config.go                  # Auth configuration
│   ├── manager.go                 # Main auth manager
│   └── example.go                 # Usage examples
├── libs/                          # Core libraries
│   ├── communication/             # Communication protocols
│   │   ├── http/                  # HTTP server
│   │   ├── websocket/             # WebSocket server
│   │   ├── grpc/                  # gRPC (existing)
│   │   └── graphql/               # GraphQL (existing)
│   ├── payment/                   # Payment processing
│   │   ├── gateway/               # Payment gateway manager
│   │   └── providers/             # Payment providers
│   │       ├── stripe/            # Stripe integration
│   │       ├── paypal/            # PayPal integration
│   │       ├── xendit/            # Xendit integration
│   │       └── midtrans/          # Midtrans integration
│   ├── core/                      # Core utilities
│   │   ├── registry/              # Service registry
│   │   ├── health/                # Health checking
│   │   ├── config/                # Configuration management
│   │   └── errors/                # Error handling
│   └── infrastructure/            # Infrastructure components
│       ├── database/              # Database connections
│       ├── cache/                 # Caching layer
│       ├── messaging/             # Message queues
│       ├── monitoring/            # Metrics & monitoring
│       └── storage/               # File storage
├── middleware/                    # HTTP middleware
├── discovery/                     # Service discovery
├── events/                        # Event handling
├── validation/                    # Input validation
└── utils/                         # Utility functions
```

## 🛠️ Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Basic Installation

```bash
# Clone the repository
git clone https://github.com/anasamu/microservices-library-go.git
cd microservices-library-go

# Install dependencies
go mod tidy
```

### Using Specific Modules

Each module can be used independently:

```bash
# For authentication
go get github.com/anasamu/microservices-library-go/auth

# For payment processing
go get github.com/anasamu/microservices-library-go/libs/payment/gateway

# For HTTP communication
go get github.com/anasamu/microservices-library-go/libs/communication/http

# For service registry
go get github.com/anasamu/microservices-library-go/libs/core/registry
```

## 📖 Usage Examples

### Authentication

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/auth"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Load configuration from environment
    config := auth.LoadAuthConfigFromEnv()
    
    // Create authentication manager
    authManager, err := auth.NewAuthManager(config, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate token
    userID := uuid.New()
    tenantID := uuid.New()
    roles := []string{"user", "admin"}
    permissions := []string{"read:all", "write:own"}
    
    tokenPair, err := authManager.AuthenticateUser(
        context.Background(),
        userID,
        tenantID,
        "user@example.com",
        roles,
        permissions,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Access Token: %s", tokenPair.AccessToken)
}
```

### Payment Processing

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/libs/payment/gateway"
    "github.com/anasamu/microservices-library-go/libs/payment/providers/stripe"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create payment manager
    config := gateway.DefaultManagerConfig()
    paymentManager := gateway.NewPaymentManager(config, logger)
    
    // Register Stripe provider
    stripeProvider := stripe.NewProvider(logger)
    stripeConfig := map[string]interface{}{
        "api_key": "sk_test_your_stripe_secret_key",
    }
    stripeProvider.Configure(stripeConfig)
    paymentManager.RegisterProvider(stripeProvider)
    
    // Create payment
    request := &gateway.PaymentRequest{
        Amount:      2000, // $20.00 in cents
        Currency:    "USD",
        Description: "Test payment",
        Customer: &gateway.Customer{
            Email: "customer@example.com",
            Name:  "John Doe",
        },
        PaymentMethod: gateway.PaymentMethodCard,
        ReturnURL:     "https://example.com/success",
        CancelURL:     "https://example.com/cancel",
    }
    
    response, err := paymentManager.CreatePayment(
        context.Background(),
        "stripe",
        request,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Payment URL: %s", response.PaymentURL)
}
```

### HTTP Server

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/anasamu/microservices-library-go/libs/communication/http"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create HTTP server
    config := http.DefaultServerConfig()
    config.Port = 8080
    server := http.NewServer(config, logger)
    
    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    // Start server
    log.Fatal(server.Start(handler))
}
```

### Service Registry

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/libs/core/registry"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create service registry
    config := registry.DefaultRegistryConfig()
    registry := registry.NewServiceRegistry(config, logger)
    
    // Register a service
    service := &registry.Service{
        Name:     "user-service",
        Version:  "1.0.0",
        Host:     "localhost",
        Port:     8080,
        Protocol: "http",
        Tags:     []string{"api", "user"},
        Metadata: map[string]interface{}{
            "environment": "production",
        },
    }
    
    err := registry.RegisterService(context.Background(), service)
    if err != nil {
        log.Fatal(err)
    }
    
    // Find services
    services, err := registry.GetServicesByName(context.Background(), "user-service")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d services", len(services))
}
```

### Health Checking

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/libs/core/health"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create health checker
    config := health.DefaultHealthConfig()
    checker := health.NewHealthChecker(config, logger)
    
    // Register HTTP health check
    httpCheck := health.NewHTTPHealthCheck(
        "api_health",
        "http://localhost:8080/health",
        5*time.Second,
        30*time.Second,
    )
    checker.RegisterCheck(httpCheck)
    
    // Register custom health check
    customCheck := health.NewCustomHealthCheck(
        "database_health",
        5*time.Second,
        30*time.Second,
        func(ctx context.Context) *health.CheckResult {
            // Your custom health check logic
            return &health.CheckResult{
                Status:  health.HealthStatusHealthy,
                Message: "Database is healthy",
            }
        },
    )
    checker.RegisterCheck(customCheck)
    
    // Perform health checks
    overallHealth := checker.CheckAll(context.Background())
    log.Printf("Overall health: %s", overallHealth.Status)
}
```

## 🔧 Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Authentication
export AUTH_JWT_SECRET_KEY="your-secret-key"
export AUTH_JWT_ACCESS_EXPIRY="15m"
export AUTH_JWT_REFRESH_EXPIRY="7d"
export AUTH_RBAC_ENABLED="true"
export AUTH_ACL_ENABLED="true"
export AUTH_ABAC_ENABLED="true"

# Payment Gateway
export PAYMENT_DEFAULT_PROVIDER="stripe"
export PAYMENT_STRIPE_API_KEY="sk_test_..."
export PAYMENT_PAYPAL_CLIENT_ID="your-paypal-client-id"
export PAYMENT_PAYPAL_CLIENT_SECRET="your-paypal-client-secret"

# Service Registry
export REGISTRY_CLEANUP_INTERVAL="30s"
export REGISTRY_DEFAULT_TTL="60s"
export REGISTRY_ENABLE_HEALTH_CHECKS="true"

# Health Checks
export HEALTH_DEFAULT_TIMEOUT="5s"
export HEALTH_DEFAULT_INTERVAL="30s"
export HEALTH_MAX_CONCURRENCY="10"
```

### Configuration Files

You can also use configuration files:

```json
{
  "auth": {
    "jwt": {
      "secret_key": "your-secret-key",
      "access_expiry": "15m",
      "refresh_expiry": "7d"
    },
    "rbac": {
      "enabled": true,
      "default_roles": ["user"],
      "default_permissions": ["read:own"]
    }
  },
  "payment": {
    "default_provider": "stripe",
    "providers": {
      "stripe": {
        "api_key": "sk_test_...",
        "enabled": true
      }
    }
  }
}
```

## 🧪 Testing

Run tests for all modules:

```bash
# Run all tests
go test ./...

# Run tests for specific module
go test ./auth/...
go test ./libs/payment/gateway/...
go test ./libs/communication/http/...
```

## 📚 API Documentation

### Authentication API

- `NewAuthManager(config, logger)` - Create authentication manager
- `AuthenticateUser(ctx, userID, tenantID, email, roles, permissions, metadata)` - Generate tokens
- `ValidateUserToken(ctx, token)` - Validate JWT token
- `CheckUserAccess(ctx, userID, tenantID, resource, action, context)` - Check permissions

### Payment API

- `NewPaymentManager(config, logger)` - Create payment manager
- `RegisterProvider(provider)` - Register payment provider
- `CreatePayment(ctx, provider, request)` - Create payment
- `GetPayment(ctx, provider, paymentID)` - Get payment status
- `RefundPayment(ctx, provider, request)` - Process refund

### Communication API

- `NewServer(config, logger)` - Create HTTP server
- `Start(handler)` - Start server
- `Stop(ctx)` - Stop server gracefully

### Service Registry API

- `NewServiceRegistry(config, logger)` - Create service registry
- `RegisterService(ctx, service)` - Register service
- `FindServices(ctx, query)` - Find services
- `GetServicesByName(ctx, name)` - Get services by name

### Health Check API

- `NewHealthChecker(config, logger)` - Create health checker
- `RegisterCheck(check)` - Register health check
- `CheckAll(ctx)` - Perform all health checks
- `CheckSpecific(ctx, names)` - Perform specific health checks

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

- [Stripe](https://stripe.com) for payment processing
- [PayPal](https://paypal.com) for payment solutions
- [Xendit](https://xendit.co) for Southeast Asian payments
- [Midtrans](https://midtrans.com) for Indonesian payments
- [gRPC](https://grpc.io) for high-performance RPC
- [Logrus](https://github.com/sirupsen/logrus) for structured logging

---

Made with ❤️ for the Go microservices community
