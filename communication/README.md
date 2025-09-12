# Communication Gateway Library

A comprehensive, modular, and production-ready communication library for Go microservices. This library provides a unified interface for multiple communication providers including HTTP, WebSocket, gRPC, QUIC, GraphQL, and SSE.

## 🚀 Features

### 🔧 Multi-Provider Support
- **HTTP**: Full HTTP server support with middleware, CORS, and health checks
- **WebSocket**: Complete WebSocket integration with real-time messaging
- **gRPC**: gRPC server support with reflection and service registration
- **QUIC**: QUIC protocol support for high-performance communication
- **GraphQL**: GraphQL server support with query handling
- **SSE**: Server-Sent Events for real-time streaming

### 📊 Core Operations
- **Request Handling**: Handle HTTP requests with unified interface
- **WebSocket Connections**: Manage WebSocket connections and real-time communication
- **Message Operations**: Send and broadcast messages across connections
- **Connection Management**: Monitor and manage active connections
- **Server Management**: Start, stop, and configure communication servers

### 🔗 Advanced Features
- **Message Headers**: Custom headers and metadata support
- **Message Filtering**: Filter messages based on criteria
- **Connection Filtering**: Target specific connections or users
- **Real-time Broadcasting**: Broadcast messages to multiple connections
- **Connection Statistics**: Detailed connection and message statistics
- **Health Monitoring**: Real-time server health monitoring

### 🏥 Production Features
- **Server Management**: Automatic server lifecycle management
- **Retry Logic**: Configurable retry with exponential backoff
- **Health Monitoring**: Real-time communication server health monitoring
- **Statistics**: Detailed communication statistics and metrics
- **Error Handling**: Comprehensive error reporting with context
- **Logging**: Structured logging with detailed context

## 📁 Project Structure

```
communication/
├── gateway/                    # Core communication gateway
│   ├── manager.go             # Communication manager implementation
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Communication provider implementations
│   ├── http/                  # HTTP provider
│   │   ├── provider.go        # HTTP implementation
│   │   └── go.mod             # HTTP dependencies
│   ├── websocket/             # WebSocket provider
│   │   ├── provider.go        # WebSocket implementation
│   │   └── go.mod             # WebSocket dependencies
│   ├── grpc/                  # gRPC provider
│   │   ├── provider.go        # gRPC implementation
│   │   └── go.mod             # gRPC dependencies
│   ├── quic/                  # QUIC provider
│   │   ├── provider.go        # QUIC implementation
│   │   └── go.mod             # QUIC dependencies
│   ├── graphql/               # GraphQL provider
│   │   ├── provider.go        # GraphQL implementation
│   │   └── go.mod             # GraphQL dependencies
│   └── sse/                   # Server-Sent Events provider
│       ├── provider.go        # SSE implementation
│       └── go.mod             # SSE dependencies
├── go.mod                     # Main module dependencies
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
cd microservices-library-go/communication

# Install dependencies
go mod tidy
```

### Using Specific Providers

```bash
# For HTTP support
go get github.com/anasamu/microservices-library-go/communication/providers/http

# For WebSocket support
go get github.com/anasamu/microservices-library-go/communication/providers/websocket

# For gRPC support
go get github.com/anasamu/microservices-library-go/communication/providers/grpc

# For QUIC support
go get github.com/anasamu/microservices-library-go/communication/providers/quic

# For GraphQL support
go get github.com/anasamu/microservices-library-go/communication/providers/graphql

# For SSE support
go get github.com/anasamu/microservices-library-go/communication/providers/sse
```

## 📖 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/communication"
    "github.com/anasamu/microservices-library-go/communication/providers/http"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create communication manager
    config := gateway.DefaultManagerConfig()
    config.DefaultProvider = "http"
    config.MaxConnections = 1000
    
    communicationManager := gateway.NewCommunicationManager(config, logger)
    
    // Register HTTP provider
    httpProvider := http.NewProvider(logger)
    httpConfig := map[string]interface{}{
        "host": "0.0.0.0",
        "port": 8080,
    }
    
    if err := httpProvider.Configure(httpConfig); err != nil {
        log.Fatal(err)
    }
    
    communicationManager.RegisterProvider(httpProvider)
    
    // Use the communication manager...
}
```

### Start Communication Server

```go
// Start HTTP server
ctx := context.Background()
httpConfig := map[string]interface{}{
    "host": "0.0.0.0",
    "port": 8080,
}

err := communicationManager.Start(ctx, "http", httpConfig)
if err != nil {
    log.Fatal(err)
}

// Check if running
if communicationManager.IsProviderRunning("http") {
    log.Println("HTTP server started successfully")
}
```

### Handle HTTP Requests

```go
// Create HTTP request
request := &gateway.Request{
    Method: "GET",
    Path:   "/api/users",
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    QueryParams: map[string]string{
        "page":  "1",
        "limit": "10",
    },
    RemoteAddr: "127.0.0.1:12345",
    UserAgent:  "Example Client",
}

// Handle request
response, err := communicationManager.HandleRequest(ctx, "http", request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Response: %d - %s", response.StatusCode, string(response.Body))
```

### Handle WebSocket Connections

```go
// Create WebSocket connection request
wsRequest := &gateway.WebSocketRequest{
    Path: "/ws",
    Headers: map[string]string{
        "Upgrade":               "websocket",
        "Connection":            "Upgrade",
        "Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
        "Sec-WebSocket-Version": "13",
    },
    RemoteAddr: "127.0.0.1:12345",
    UserAgent:  "WebSocket Client",
    UserID:     "user123",
    Metadata: map[string]interface{}{
        "room": "general",
    },
}

// Handle WebSocket connection
response, err := communicationManager.HandleWebSocket(ctx, "websocket", wsRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("WebSocket connection established: %s", response.ConnectionID)
```

### Send Messages

```go
// Create a message
message := gateway.CreateMessage("chat.message", map[string]interface{}{
    "text":      "Hello, World!",
    "timestamp": time.Now().Unix(),
})
message.SetFrom("user123")

// Send message to specific connection
sendRequest := &gateway.SendMessageRequest{
    ConnectionID: "client_1234567890",
    Message:      message,
    Metadata: map[string]interface{}{
        "priority": "normal",
    },
}

response, err := communicationManager.SendMessage(ctx, "websocket", sendRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Message sent: %s", response.MessageID)
```

### Broadcast Messages

```go
// Create broadcast message
broadcastMessage := gateway.CreateMessage("system.announcement", map[string]interface{}{
    "title":   "System Maintenance",
    "content": "The system will be under maintenance in 30 minutes",
    "type":    "warning",
})

// Broadcast to all connections
broadcastRequest := &gateway.BroadcastRequest{
    Message: broadcastMessage,
    Metadata: map[string]interface{}{
        "priority": "high",
    },
}

response, err := communicationManager.BroadcastMessage(ctx, "websocket", broadcastRequest)
if err != nil {
    log.Fatal(err)
}

log.Printf("Message broadcasted: %d sent, %d failed", response.SentCount, response.FailedCount)
```

### QUIC Provider Example

```go
import (
    "github.com/anasamu/microservices-library-go/communication/providers/quic"
    "github.com/anasamu/microservices-library-go/communication/providers/graphql"
    "github.com/anasamu/microservices-library-go/communication/providers/sse"
)

// Register QUIC provider
quicProvider := quic.NewProvider(logger)
quicConfig := map[string]interface{}{
    "host":              "0.0.0.0",
    "port":              8443,
    "max_streams":       100,
    "max_idle_timeout":  "30s",
    "keep_alive_period": "10s",
}

if err := quicProvider.Configure(quicConfig); err != nil {
    log.Fatal(err)
}

communicationManager.RegisterProvider(quicProvider)

// Start QUIC server
err := communicationManager.Start(ctx, "quic", quicConfig)
if err != nil {
    log.Fatal(err)
}
```

### GraphQL Provider Example

```go
// Register GraphQL provider
graphqlProvider := graphql.NewProvider(logger)
graphqlConfig := map[string]interface{}{
    "host":                "0.0.0.0",
    "port":                8080,
    "enable_introspection": true,
    "enable_playground":   true,
}

if err := graphqlProvider.Configure(graphqlConfig); err != nil {
    log.Fatal(err)
}

communicationManager.RegisterProvider(graphqlProvider)

// Start GraphQL server
err := communicationManager.Start(ctx, "graphql", graphqlConfig)
if err != nil {
    log.Fatal(err)
}
```

### SSE Provider Example

```go
// Register SSE provider
sseProvider := sse.NewProvider(logger)
sseConfig := map[string]interface{}{
    "host":               "0.0.0.0",
    "port":               8080,
    "heartbeat_interval": "30s",
    "max_connections":    1000,
}

if err := sseProvider.Configure(sseConfig); err != nil {
    log.Fatal(err)
}

communicationManager.RegisterProvider(sseProvider)

// Start SSE server
err := communicationManager.Start(ctx, "sse", sseConfig)
if err != nil {
    log.Fatal(err)
}
```

### Connection Management

```go
// Get connections
connections, err := communicationManager.GetConnections(ctx, "websocket")
if err != nil {
    log.Fatal(err)
}

log.Printf("WebSocket connections: %d found", len(connections))
for _, conn := range connections {
    log.Printf("  - %s (%s) - %s", conn.ID, conn.Type, conn.UserID)
}

// Get connection count
count, err := communicationManager.GetConnectionCount(ctx, "websocket")
if err != nil {
    log.Fatal(err)
}

log.Printf("WebSocket connection count: %d", count)

// Close connection
if len(connections) > 0 {
    connectionID := connections[0].ID
    err = communicationManager.CloseConnection(ctx, "websocket", connectionID)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Connection closed: %s", connectionID)
}
```

### Health Checks

```go
// Perform health checks
results := communicationManager.HealthCheck(ctx)

for provider, err := range results {
    if err != nil {
        log.Printf("%s: ❌ %v", provider, err)
    } else {
        log.Printf("%s: ✅ Healthy", provider)
    }
}
```

### Communication Statistics

```go
// Get communication statistics
stats, err := communicationManager.GetStats(ctx, "websocket")
if err != nil {
    log.Fatal(err)
}

log.Printf("Total Connections: %d", stats.TotalConnections)
log.Printf("Active Connections: %d", stats.ActiveConnections)
log.Printf("Total Requests: %d", stats.TotalRequests)
log.Printf("Total Messages: %d", stats.TotalMessages)
log.Printf("Failed Requests: %d", stats.FailedRequests)
log.Printf("Failed Messages: %d", stats.FailedMessages)
log.Printf("Average Response Time: %v", stats.AverageResponseTime)
```

## 🔧 Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Communication Manager Configuration
export COMMUNICATION_DEFAULT_PROVIDER="http"
export COMMUNICATION_MAX_CONNECTIONS="1000"
export COMMUNICATION_RETRY_ATTEMPTS="3"
export COMMUNICATION_RETRY_DELAY="5s"
export COMMUNICATION_TIMEOUT="30s"

# HTTP Configuration
export HTTP_HOST="0.0.0.0"
export HTTP_PORT="8080"
export HTTP_READ_TIMEOUT="30s"
export HTTP_WRITE_TIMEOUT="30s"
export HTTP_IDLE_TIMEOUT="120s"

# WebSocket Configuration
export WEBSOCKET_READ_BUFFER_SIZE="1024"
export WEBSOCKET_WRITE_BUFFER_SIZE="1024"
export WEBSOCKET_CHECK_ORIGIN="false"
export WEBSOCKET_PING_PERIOD="54s"
export WEBSOCKET_PONG_WAIT="60s"
export WEBSOCKET_WRITE_WAIT="10s"
export WEBSOCKET_MAX_MESSAGE_SIZE="512"

# gRPC Configuration
export GRPC_HOST="0.0.0.0"
export GRPC_PORT="9090"
export GRPC_MAX_RECV_MSG_SIZE="4194304"
export GRPC_MAX_SEND_MSG_SIZE="4194304"
export GRPC_ENABLE_REFLECTION="true"

# QUIC Configuration
export QUIC_HOST="0.0.0.0"
export QUIC_PORT="8443"
export QUIC_MAX_STREAMS="100"
export QUIC_MAX_IDLE_TIMEOUT="30s"
export QUIC_KEEP_ALIVE_PERIOD="10s"

# GraphQL Configuration
export GRAPHQL_HOST="0.0.0.0"
export GRAPHQL_PORT="8080"
export GRAPHQL_ENABLE_INTROSPECTION="true"
export GRAPHQL_ENABLE_PLAYGROUND="true"

# SSE Configuration
export SSE_HOST="0.0.0.0"
export SSE_PORT="8080"
export SSE_HEARTBEAT_INTERVAL="30s"
export SSE_MAX_CONNECTIONS="1000"
```

### Configuration Files

You can also use configuration files:

```json
{
  "communication": {
    "default_provider": "http",
    "max_connections": 1000,
    "retry_attempts": 3,
    "retry_delay": "5s",
    "timeout": "30s"
  },
  "providers": {
    "http": {
      "host": "0.0.0.0",
      "port": 8080,
      "read_timeout": "30s",
      "write_timeout": "30s",
      "idle_timeout": "120s"
    },
    "websocket": {
      "read_buffer_size": 1024,
      "write_buffer_size": 1024,
      "check_origin": false,
      "ping_period": "54s",
      "pong_wait": "60s",
      "write_wait": "10s",
      "max_message_size": 512
    },
    "grpc": {
      "host": "0.0.0.0",
      "port": 9090,
      "max_recv_msg_size": 4194304,
      "max_send_msg_size": 4194304,
      "enable_reflection": true
    },
    "quic": {
      "host": "0.0.0.0",
      "port": 8443,
      "max_streams": 100,
      "max_idle_timeout": "30s",
      "keep_alive_period": "10s"
    },
    "graphql": {
      "host": "0.0.0.0",
      "port": 8080,
      "enable_introspection": true,
      "enable_playground": true
    },
    "sse": {
      "host": "0.0.0.0",
      "port": 8080,
      "heartbeat_interval": "30s",
      "max_connections": 1000
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
go test ./providers/http/...
go test ./providers/websocket/...
go test ./providers/grpc/...
go test ./providers/quic/...
go test ./providers/graphql/...
go test ./providers/sse/...

# Run gateway tests
go test ./...
```

## 📚 API Documentation

### Communication Manager API

- `NewCommunicationManager(config, logger)` - Create communication manager
- `RegisterProvider(provider)` - Register a communication provider
- `Start(ctx, provider, config)` - Start a communication server
- `Stop(ctx, provider)` - Stop a communication server
- `HandleRequest(ctx, provider, request)` - Handle an HTTP request
- `HandleWebSocket(ctx, provider, request)` - Handle a WebSocket connection
- `SendMessage(ctx, provider, request)` - Send a message
- `BroadcastMessage(ctx, provider, request)` - Broadcast a message
- `GetConnections(ctx, provider)` - Get connections
- `GetConnectionCount(ctx, provider)` - Get connection count
- `CloseConnection(ctx, provider, connectionID)` - Close a connection
- `HealthCheck(ctx)` - Check provider health
- `GetStats(ctx, provider)` - Get communication statistics

### Provider Interface

All providers implement the `CommunicationProvider` interface:

```go
type CommunicationProvider interface {
    GetName() string
    GetSupportedFeatures() []CommunicationFeature
    GetConnectionInfo() *ConnectionInfo
    Start(ctx context.Context, config map[string]interface{}) error
    Stop(ctx context.Context) error
    IsRunning() bool
    HandleRequest(ctx context.Context, request *Request) (*Response, error)
    HandleWebSocket(ctx context.Context, request *WebSocketRequest) (*WebSocketResponse, error)
    SendMessage(ctx context.Context, request *SendMessageRequest) (*SendMessageResponse, error)
    BroadcastMessage(ctx context.Context, request *BroadcastRequest) (*BroadcastResponse, error)
    GetConnections(ctx context.Context) ([]ConnectionInfo, error)
    GetConnectionCount(ctx context.Context) (int, error)
    CloseConnection(ctx context.Context, connectionID string) error
    HealthCheck(ctx context.Context) error
    GetStats(ctx context.Context) (*CommunicationStats, error)
    Configure(config map[string]interface{}) error
    IsConfigured() bool
    Close() error
}
```

### Supported Features

- `FeatureHTTP` - HTTP server support
- `FeatureHTTPS` - HTTPS server support
- `FeatureWebSocket` - WebSocket support
- `FeatureServerSentEvents` - Server-sent events
- `FeatureLongPolling` - Long polling
- `FeatureCORS` - Cross-origin resource sharing
- `FeatureCompression` - Message compression
- `FeatureAuthentication` - Authentication support
- `FeatureRateLimiting` - Rate limiting
- `FeatureLoadBalancing` - Load balancing
- `FeatureMetrics` - Metrics collection
- `FeatureLogging` - Request logging
- `FeatureHealthChecks` - Health check endpoints
- `FeatureMiddleware` - Middleware support
- `FeatureStaticFiles` - Static file serving
- `FeatureTemplates` - Template rendering
- `FeatureSessions` - Session management
- `FeatureCookies` - Cookie support
- `FeatureHeaders` - Custom headers
- `FeatureQueryParams` - Query parameters
- `FeaturePathParams` - Path parameters
- `FeatureBodyParsing` - Request body parsing
- `FeatureFileUpload` - File upload support
- `FeatureStreaming` - Streaming support
- `FeatureRealTime` - Real-time communication
- `FeatureBroadcasting` - Message broadcasting
- `FeatureMulticasting` - Message multicasting
- `FeatureUnicasting` - Message unicasting

## 🔒 Security Considerations

### Connection Security

- **SSL/TLS**: All providers support encrypted connections
- **Authentication**: Multiple authentication methods per provider
- **Connection Validation**: Validate connections and origins
- **Rate Limiting**: Prevent abuse and DoS attacks

### Message Security

- **Message Validation**: Validate message content and structure
- **Access Control**: Connection and user-based access control
- **Message Encryption**: Encrypted message transmission
- **Message Signing**: Message integrity verification

## 🚀 Performance

### Optimization Features

- **Connection Pooling**: Efficient connection management
- **Message Batching**: Efficient bulk operations
- **Retry Logic**: Automatic retry with backoff
- **Connection Reuse**: Minimize connection overhead
- **Message Compression**: Reduce message size

### Monitoring

- **Health Checks**: Real-time server health monitoring
- **Statistics**: Detailed performance metrics
- **Connection Monitoring**: Connection pool statistics
- **Message Performance**: Message processing monitoring

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

- [Go HTTP Server](https://golang.org/pkg/net/http/) for HTTP server functionality
- [Gorilla WebSocket](https://github.com/gorilla/websocket) for WebSocket support
- [gRPC Go](https://github.com/grpc/grpc-go) for gRPC server functionality
- [QUIC Go](https://github.com/quic-go/quic-go) for QUIC protocol support
- [GraphQL Go](https://github.com/graphql-go/graphql) for GraphQL server functionality
- [Logrus](https://github.com/sirupsen/logrus) for structured logging

---

Made with ❤️ for the Go microservices community
