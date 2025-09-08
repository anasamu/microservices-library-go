package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/libs/communication/providers/grpc"
	"github.com/anasamu/microservices-library-go/libs/communication/providers/http"
	"github.com/anasamu/microservices-library-go/libs/communication/providers/websocket"
	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the communication gateway system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create communication manager
	config := DefaultManagerConfig()
	config.DefaultProvider = "http"
	config.MaxConnections = 1000
	config.RetryAttempts = 3
	config.RetryDelay = 5 * time.Second

	communicationManager := NewCommunicationManager(config, logger)

	// Register communication providers
	registerProviders(communicationManager, logger)

	// Example 1: Start HTTP server
	exampleHTTPServer(communicationManager)

	// Example 2: Start WebSocket server
	exampleWebSocketServer(communicationManager)

	// Example 3: Start gRPC server
	exampleGRPCServer(communicationManager)

	// Example 4: Handle HTTP requests
	exampleHTTPRequests(communicationManager)

	// Example 5: Handle WebSocket connections
	exampleWebSocketConnections(communicationManager)

	// Example 6: Send messages
	exampleSendMessages(communicationManager)

	// Example 7: Broadcast messages
	exampleBroadcastMessages(communicationManager)

	// Example 8: Connection management
	exampleConnectionManagement(communicationManager)

	// Example 9: Health checks
	exampleHealthChecks(communicationManager)

	// Example 10: Communication statistics
	exampleCommunicationStats(communicationManager)
}

// registerProviders registers all communication providers
func registerProviders(communicationManager *CommunicationManager, logger *logrus.Logger) {
	// Register HTTP provider
	httpProvider := http.NewProvider(logger)
	httpConfig := map[string]interface{}{
		"host":          "0.0.0.0",
		"port":          8080,
		"read_timeout":  30 * time.Second,
		"write_timeout": 30 * time.Second,
		"idle_timeout":  120 * time.Second,
	}
	if err := httpProvider.Configure(httpConfig); err != nil {
		log.Printf("Failed to configure HTTP: %v", err)
	} else {
		communicationManager.RegisterProvider(httpProvider)
	}

	// Register WebSocket provider
	websocketProvider := websocket.NewProvider(logger)
	websocketConfig := map[string]interface{}{
		"read_buffer_size":  1024,
		"write_buffer_size": 1024,
		"check_origin":      false,
		"ping_period":       54 * time.Second,
		"pong_wait":         60 * time.Second,
		"write_wait":        10 * time.Second,
		"max_message_size":  512,
	}
	if err := websocketProvider.Configure(websocketConfig); err != nil {
		log.Printf("Failed to configure WebSocket: %v", err)
	} else {
		communicationManager.RegisterProvider(websocketProvider)
	}

	// Register gRPC provider
	grpcProvider := grpc.NewProvider(logger)
	grpcConfig := map[string]interface{}{
		"host":              "0.0.0.0",
		"port":              9090,
		"max_recv_msg_size": 4 * 1024 * 1024, // 4MB
		"max_send_msg_size": 4 * 1024 * 1024, // 4MB
		"enable_reflection": true,
	}
	if err := grpcProvider.Configure(grpcConfig); err != nil {
		log.Printf("Failed to configure gRPC: %v", err)
	} else {
		communicationManager.RegisterProvider(grpcProvider)
	}
}

// exampleHTTPServer demonstrates HTTP server usage
func exampleHTTPServer(communicationManager *CommunicationManager) {
	fmt.Println("=== HTTP Server Example ===")

	ctx := context.Background()

	// Start HTTP server
	httpConfig := map[string]interface{}{
		"host": "0.0.0.0",
		"port": 8080,
	}

	err := communicationManager.Start(ctx, "http", httpConfig)
	if err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
		return
	}

	// Check if running
	if communicationManager.IsProviderRunning("http") {
		fmt.Println("HTTP server started successfully")
	}

	// Wait a bit for server to start
	time.Sleep(1 * time.Second)
}

// exampleWebSocketServer demonstrates WebSocket server usage
func exampleWebSocketServer(communicationManager *CommunicationManager) {
	fmt.Println("\n=== WebSocket Server Example ===")

	ctx := context.Background()

	// Start WebSocket server
	websocketConfig := map[string]interface{}{
		"read_buffer_size":  1024,
		"write_buffer_size": 1024,
		"check_origin":      false,
	}

	err := communicationManager.Start(ctx, "websocket", websocketConfig)
	if err != nil {
		log.Printf("Failed to start WebSocket server: %v", err)
		return
	}

	// Check if running
	if communicationManager.IsProviderRunning("websocket") {
		fmt.Println("WebSocket server started successfully")
	}
}

// exampleGRPCServer demonstrates gRPC server usage
func exampleGRPCServer(communicationManager *CommunicationManager) {
	fmt.Println("\n=== gRPC Server Example ===")

	ctx := context.Background()

	// Start gRPC server
	grpcConfig := map[string]interface{}{
		"host":              "0.0.0.0",
		"port":              9090,
		"enable_reflection": true,
	}

	err := communicationManager.Start(ctx, "grpc", grpcConfig)
	if err != nil {
		log.Printf("Failed to start gRPC server: %v", err)
		return
	}

	// Check if running
	if communicationManager.IsProviderRunning("grpc") {
		fmt.Println("gRPC server started successfully")
	}
}

// exampleHTTPRequests demonstrates HTTP request handling
func exampleHTTPRequests(communicationManager *CommunicationManager) {
	fmt.Println("\n=== HTTP Request Example ===")

	ctx := context.Background()

	// Example 1: GET request
	getRequest := &Request{
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

	response, err := communicationManager.HandleRequest(ctx, "http", getRequest)
	if err != nil {
		log.Printf("Failed to handle GET request: %v", err)
	} else {
		fmt.Printf("GET request handled: Status %d\n", response.StatusCode)
	}

	// Example 2: POST request
	postRequest := &Request{
		Method: "POST",
		Path:   "/api/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:       []byte(`{"name":"John Doe","email":"john@example.com"}`),
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "Example Client",
	}

	response, err = communicationManager.HandleRequest(ctx, "http", postRequest)
	if err != nil {
		log.Printf("Failed to handle POST request: %v", err)
	} else {
		fmt.Printf("POST request handled: Status %d\n", response.StatusCode)
	}

	// Example 3: Health check request
	healthRequest := &Request{
		Method:     "GET",
		Path:       "/health",
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "Health Checker",
	}

	response, err = communicationManager.HandleRequest(ctx, "http", healthRequest)
	if err != nil {
		log.Printf("Failed to handle health request: %v", err)
	} else {
		fmt.Printf("Health check handled: Status %d\n", response.StatusCode)
	}
}

// exampleWebSocketConnections demonstrates WebSocket connection handling
func exampleWebSocketConnections(communicationManager *CommunicationManager) {
	fmt.Println("\n=== WebSocket Connection Example ===")

	ctx := context.Background()

	// Example 1: WebSocket connection request
	wsRequest := &WebSocketRequest{
		Path: "/ws",
		Headers: map[string]string{
			"Upgrade":               "websocket",
			"Connection":            "Upgrade",
			"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
			"Sec-WebSocket-Version": "13",
		},
		QueryParams: map[string]string{
			"token": "user123",
		},
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "WebSocket Client",
		UserID:     "user123",
		Metadata: map[string]interface{}{
			"room": "general",
		},
	}

	response, err := communicationManager.HandleWebSocket(ctx, "websocket", wsRequest)
	if err != nil {
		log.Printf("Failed to handle WebSocket connection: %v", err)
	} else {
		fmt.Printf("WebSocket connection established: %s\n", response.ConnectionID)
	}

	// Example 2: Another WebSocket connection
	wsRequest2 := &WebSocketRequest{
		Path: "/ws/chat",
		Headers: map[string]string{
			"Upgrade":               "websocket",
			"Connection":            "Upgrade",
			"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
			"Sec-WebSocket-Version": "13",
		},
		RemoteAddr: "127.0.0.1:12346",
		UserAgent:  "WebSocket Client 2",
		UserID:     "user456",
		Metadata: map[string]interface{}{
			"room": "chat",
		},
	}

	response, err = communicationManager.HandleWebSocket(ctx, "websocket", wsRequest2)
	if err != nil {
		log.Printf("Failed to handle WebSocket connection 2: %v", err)
	} else {
		fmt.Printf("WebSocket connection 2 established: %s\n", response.ConnectionID)
	}
}

// exampleSendMessages demonstrates message sending
func exampleSendMessages(communicationManager *CommunicationManager) {
	fmt.Println("\n=== Send Messages Example ===")

	ctx := context.Background()

	// Example 1: Send message to specific connection
	message := CreateMessage("chat.message", map[string]interface{}{
		"text":      "Hello, World!",
		"timestamp": time.Now().Unix(),
	})
	message.SetFrom("user123")

	sendRequest := &SendMessageRequest{
		ConnectionID: "client_1234567890",
		Message:      message,
		Metadata: map[string]interface{}{
			"priority": "normal",
		},
	}

	response, err := communicationManager.SendMessage(ctx, "websocket", sendRequest)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	} else {
		fmt.Printf("Message sent: %s (Status: %s)\n", response.MessageID, response.Status)
	}

	// Example 2: Send message to specific user
	userMessage := CreateMessage("notification", map[string]interface{}{
		"title":   "New Message",
		"content": "You have a new message",
		"type":    "info",
	})

	userSendRequest := &SendMessageRequest{
		UserID:  "user456",
		Message: userMessage,
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	response, err = communicationManager.SendMessage(ctx, "websocket", userSendRequest)
	if err != nil {
		log.Printf("Failed to send user message: %v", err)
	} else {
		fmt.Printf("User message sent: %s (Status: %s)\n", response.MessageID, response.Status)
	}
}

// exampleBroadcastMessages demonstrates message broadcasting
func exampleBroadcastMessages(communicationManager *CommunicationManager) {
	fmt.Println("\n=== Broadcast Messages Example ===")

	ctx := context.Background()

	// Example 1: Broadcast to all connections
	broadcastMessage := CreateMessage("system.announcement", map[string]interface{}{
		"title":   "System Maintenance",
		"content": "The system will be under maintenance in 30 minutes",
		"type":    "warning",
	})

	broadcastRequest := &BroadcastRequest{
		Message: broadcastMessage,
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	response, err := communicationManager.BroadcastMessage(ctx, "websocket", broadcastRequest)
	if err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	} else {
		fmt.Printf("Message broadcasted: %d sent, %d failed\n", response.SentCount, response.FailedCount)
	}

	// Example 2: Broadcast with filter
	filteredMessage := CreateMessage("room.message", map[string]interface{}{
		"text":      "Hello everyone in the room!",
		"room":      "general",
		"timestamp": time.Now().Unix(),
	})

	filteredBroadcastRequest := &BroadcastRequest{
		Message: filteredMessage,
		Filter: map[string]interface{}{
			"room": "general",
		},
		Metadata: map[string]interface{}{
			"priority": "normal",
		},
	}

	response, err = communicationManager.BroadcastMessage(ctx, "websocket", filteredBroadcastRequest)
	if err != nil {
		log.Printf("Failed to broadcast filtered message: %v", err)
	} else {
		fmt.Printf("Filtered message broadcasted: %d sent, %d failed\n", response.SentCount, response.FailedCount)
	}
}

// exampleConnectionManagement demonstrates connection management
func exampleConnectionManagement(communicationManager *CommunicationManager) {
	fmt.Println("\n=== Connection Management Example ===")

	ctx := context.Background()

	// Example 1: Get connections
	connections, err := communicationManager.GetConnections(ctx, "websocket")
	if err != nil {
		log.Printf("Failed to get connections: %v", err)
	} else {
		fmt.Printf("WebSocket connections: %d found\n", len(connections))
		for _, conn := range connections {
			fmt.Printf("  - %s (%s) - %s\n", conn.ID, conn.Type, conn.UserID)
		}
	}

	// Example 2: Get connection count
	count, err := communicationManager.GetConnectionCount(ctx, "websocket")
	if err != nil {
		log.Printf("Failed to get connection count: %v", err)
	} else {
		fmt.Printf("WebSocket connection count: %d\n", count)
	}

	// Example 3: Close connection
	if len(connections) > 0 {
		connectionID := connections[0].ID
		err = communicationManager.CloseConnection(ctx, "websocket", connectionID)
		if err != nil {
			log.Printf("Failed to close connection: %v", err)
		} else {
			fmt.Printf("Connection closed: %s\n", connectionID)
		}
	}

	// Example 4: Get HTTP connections
	httpConnections, err := communicationManager.GetConnections(ctx, "http")
	if err != nil {
		log.Printf("Failed to get HTTP connections: %v", err)
	} else {
		fmt.Printf("HTTP connections: %d found\n", len(httpConnections))
	}
}

// exampleHealthChecks demonstrates health checks
func exampleHealthChecks(communicationManager *CommunicationManager) {
	fmt.Println("\n=== Health Checks Example ===")

	ctx := context.Background()

	// Perform health checks on all providers
	results := communicationManager.HealthCheck(ctx)

	fmt.Printf("Health Check Results:\n")
	for provider, err := range results {
		if err != nil {
			fmt.Printf("  %s: ❌ %v\n", provider, err)
		} else {
			fmt.Printf("  %s: ✅ Healthy\n", provider)
		}
	}

	// Get running providers
	running := communicationManager.GetRunningProviders()
	fmt.Printf("\nRunning Providers: %v\n", running)
}

// exampleCommunicationStats demonstrates communication statistics
func exampleCommunicationStats(communicationManager *CommunicationManager) {
	fmt.Println("\n=== Communication Statistics Example ===")

	ctx := context.Background()

	// Get statistics for each provider
	providers := communicationManager.GetSupportedProviders()
	for _, providerName := range providers {
		if communicationManager.IsProviderRunning(providerName) {
			stats, err := communicationManager.GetStats(ctx, providerName)
			if err != nil {
				log.Printf("Failed to get stats for %s: %v", providerName, err)
				continue
			}

			fmt.Printf("%s Statistics:\n", providerName)
			fmt.Printf("  Total Connections: %d\n", stats.TotalConnections)
			fmt.Printf("  Active Connections: %d\n", stats.ActiveConnections)
			fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
			fmt.Printf("  Total Messages: %d\n", stats.TotalMessages)
			fmt.Printf("  Failed Requests: %d\n", stats.FailedRequests)
			fmt.Printf("  Failed Messages: %d\n", stats.FailedMessages)
			fmt.Printf("  Average Response Time: %v\n", stats.AverageResponseTime)
		}
	}
}

// ExampleProviderCapabilities demonstrates provider capabilities
func ExampleProviderCapabilities() {
	fmt.Println("\n=== Provider Capabilities Example ===")

	// Create logger and communication manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	communicationManager := NewCommunicationManager(config, logger)

	// Register providers
	registerProviders(communicationManager, logger)

	// Get capabilities for each provider
	providers := communicationManager.GetSupportedProviders()
	for _, providerName := range providers {
		features, connInfo, err := communicationManager.GetProviderCapabilities(providerName)
		if err != nil {
			log.Printf("Failed to get capabilities for %s: %v", providerName, err)
			continue
		}

		fmt.Printf("%s Capabilities:\n", providerName)
		fmt.Printf("  Features: %v\n", features)
		fmt.Printf("  Connection Info: %+v\n", connInfo)
	}
}

// ExampleServerManagement demonstrates server management
func ExampleServerManagement() {
	fmt.Println("\n=== Server Management Example ===")

	// Create logger and communication manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	communicationManager := NewCommunicationManager(config, logger)

	// Register providers
	registerProviders(communicationManager, logger)

	ctx := context.Background()

	// Start multiple servers
	servers := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "http",
			config: map[string]interface{}{
				"host": "0.0.0.0",
				"port": 8080,
			},
		},
		{
			name: "websocket",
			config: map[string]interface{}{
				"read_buffer_size":  1024,
				"write_buffer_size": 1024,
			},
		},
		{
			name: "grpc",
			config: map[string]interface{}{
				"host": "0.0.0.0",
				"port": 9090,
			},
		},
	}

	// Start servers
	for _, server := range servers {
		if err := communicationManager.Start(ctx, server.name, server.config); err != nil {
			log.Printf("Failed to start %s server: %v", server.name, err)
		} else {
			fmt.Printf("Started %s server\n", server.name)
		}
	}

	// Check server status
	fmt.Printf("Running servers: %v\n", communicationManager.GetRunningProviders())

	// Stop servers
	for _, server := range servers {
		if err := communicationManager.Stop(ctx, server.name); err != nil {
			log.Printf("Failed to stop %s server: %v", server.name, err)
		} else {
			fmt.Printf("Stopped %s server\n", server.name)
		}
	}

	// Close all connections
	if err := communicationManager.Close(); err != nil {
		log.Printf("Failed to close communication manager: %v", err)
	} else {
		fmt.Println("Communication manager closed successfully")
	}
}

// ExampleAdvancedFeatures demonstrates advanced communication features
func ExampleAdvancedFeatures() {
	fmt.Println("\n=== Advanced Features Example ===")

	// Create logger and communication manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	communicationManager := NewCommunicationManager(config, logger)

	// Register providers
	registerProviders(communicationManager, logger)

	ctx := context.Background()

	// Start WebSocket server
	websocketConfig := map[string]interface{}{
		"read_buffer_size":  1024,
		"write_buffer_size": 1024,
		"check_origin":      false,
	}

	if err := communicationManager.Start(ctx, "websocket", websocketConfig); err != nil {
		log.Printf("Failed to start WebSocket server: %v", err)
		return
	}

	// Example 1: Message with headers and metadata
	message := CreateMessage("advanced.test", map[string]interface{}{
		"data": "advanced message",
	})

	// Add headers
	message.AddHeader("content-type", "application/json")
	message.AddHeader("version", "2.0")

	// Add metadata
	message.AddMetadata("priority", "high")
	message.AddMetadata("retry-count", 0)

	// Set sender and recipient
	message.SetFrom("advanced-service")
	message.SetTo("consumer-service")

	// Send message
	sendRequest := &SendMessageRequest{
		ConnectionID: "client_1234567890",
		Message:      message,
		Metadata: map[string]interface{}{
			"custom-header": "custom-value",
		},
	}

	response, err := communicationManager.SendMessage(ctx, "websocket", sendRequest)
	if err != nil {
		log.Printf("Failed to send advanced message: %v", err)
	} else {
		fmt.Printf("Advanced message sent: %s\n", response.MessageID)
	}

	// Example 2: Broadcast with complex filter
	broadcastMessage := CreateMessage("room.announcement", map[string]interface{}{
		"title":   "Room Announcement",
		"content": "Welcome to the room!",
		"room":    "general",
	})

	broadcastRequest := &BroadcastRequest{
		Message: broadcastMessage,
		Filter: map[string]interface{}{
			"room":    "general",
			"status":  "active",
			"version": "2.0",
		},
		Metadata: map[string]interface{}{
			"priority": "normal",
			"ttl":      300, // 5 minutes
		},
	}

	response, err = communicationManager.BroadcastMessage(ctx, "websocket", broadcastRequest)
	if err != nil {
		log.Printf("Failed to broadcast advanced message: %v", err)
	} else {
		fmt.Printf("Advanced message broadcasted: %d sent, %d failed\n", response.SentCount, response.FailedCount)
	}
}
