package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/logging"
	"github.com/anasamu/microservices-library-go/logging/providers/console"
	"github.com/anasamu/microservices-library-go/logging/providers/elasticsearch"
	"github.com/anasamu/microservices-library-go/logging/providers/file"
	"github.com/anasamu/microservices-library-go/logging/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a logger for the manager itself
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create logging manager
	managerConfig := &gateway.ManagerConfig{
		DefaultProvider: "console",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewLoggingManager(managerConfig, logger)

	// Example 1: Console Provider
	fmt.Println("=== Console Provider Example ===")
	consoleExample(manager)

	// Example 2: File Provider
	fmt.Println("\n=== File Provider Example ===")
	fileExample(manager)

	// Example 3: Elasticsearch Provider (if available)
	fmt.Println("\n=== Elasticsearch Provider Example ===")
	elasticsearchExample(manager)

	// Example 4: Multiple Providers
	fmt.Println("\n=== Multiple Providers Example ===")
	multipleProvidersExample(manager)

	// Example 5: Advanced Features
	fmt.Println("\n=== Advanced Features Example ===")
	advancedFeaturesExample(manager)

	// Clean up
	if err := manager.Close(); err != nil {
		log.Printf("Error closing manager: %v", err)
	}
}

func consoleExample(manager *gateway.LoggingManager) {
	// Create console provider
	consoleConfig := &types.LoggingConfig{
		Level:   types.LevelInfo,
		Format:  types.FormatJSON,
		Output:  types.OutputStdout,
		Service: "example-service",
		Version: "1.0.0",
	}

	consoleProvider := console.NewConsoleProvider(consoleConfig)
	if err := manager.RegisterProvider(consoleProvider); err != nil {
		log.Printf("Failed to register console provider: %v", err)
		return
	}

	// Set as default
	ctx := context.Background()

	// Basic logging
	manager.Info(ctx, "This is an info message")
	manager.Debug(ctx, "This is a debug message")
	manager.Warn(ctx, "This is a warning message")
	manager.Error(ctx, "This is an error message")

	// Formatted logging
	manager.Infof(ctx, "User %s logged in at %s", "john_doe", time.Now().Format(time.RFC3339))

	// Structured logging
	manager.Info(ctx, "User action", map[string]interface{}{
		"user_id":    "123",
		"action":     "login",
		"ip_address": "192.168.1.1",
		"user_agent": "Mozilla/5.0...",
	})

	// Context logging
	ctxWithTrace := context.WithValue(ctx, "trace_id", "trace-123")
	ctxWithTrace = context.WithValue(ctxWithTrace, "user_id", "user-456")
	manager.Info(ctxWithTrace, "Request processed", map[string]interface{}{
		"method":      "GET",
		"path":        "/api/users",
		"status_code": 200,
		"duration":    150,
	})
}

func fileExample(manager *gateway.LoggingManager) {
	// Create file provider
	fileConfig := &types.LoggingConfig{
		Level:   types.LevelDebug,
		Format:  types.FormatJSON,
		Output:  types.OutputFile,
		Service: "file-service",
		Version: "1.0.0",
		Metadata: map[string]interface{}{
			"file_path": "./logs/application.log",
		},
	}

	fileProvider := file.NewFileProvider(fileConfig)
	if err := manager.RegisterProvider(fileProvider); err != nil {
		log.Printf("Failed to register file provider: %v", err)
		return
	}

	ctx := context.Background()

	// Connect to file
	if err := fileProvider.Connect(ctx); err != nil {
		log.Printf("Failed to connect to file provider: %v", err)
		return
	}

	// Log to file
	manager.LogWithProvider(ctx, "file", types.LevelInfo, "This message goes to file", map[string]interface{}{})
	manager.LogWithProvider(ctx, "file", types.LevelError, "This error goes to file", map[string]interface{}{
		"error_code": "FILE_001",
		"details":    "Failed to process file",
	})

	// Batch logging
	entries := []types.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     types.LevelInfo,
			Message:   "Batch entry 1",
			Service:   "file-service",
			Version:   "1.0.0",
			Fields: map[string]interface{}{
				"batch_id": "batch-001",
			},
		},
		{
			Timestamp: time.Now(),
			Level:     types.LevelInfo,
			Message:   "Batch entry 2",
			Service:   "file-service",
			Version:   "1.0.0",
			Fields: map[string]interface{}{
				"batch_id": "batch-001",
			},
		},
	}

	if err := manager.LogBatchWithProvider(ctx, "file", entries); err != nil {
		log.Printf("Failed to log batch: %v", err)
	}

	// Disconnect
	if err := fileProvider.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect from file provider: %v", err)
	}
}

func elasticsearchExample(manager *gateway.LoggingManager) {
	// Create Elasticsearch provider (only if Elasticsearch is available)
	elasticConfig := &types.LoggingConfig{
		Level:   types.LevelInfo,
		Format:  types.FormatJSON,
		Service: "elasticsearch-service",
		Version: "1.0.0",
		Index:   "application-logs",
		Metadata: map[string]interface{}{
			"elastic_url": "http://localhost:9200",
		},
	}

	elasticProvider, err := elasticsearch.NewElasticsearchProvider(elasticConfig)
	if err != nil {
		log.Printf("Failed to create Elasticsearch provider: %v", err)
		return
	}

	if err := manager.RegisterProvider(elasticProvider); err != nil {
		log.Printf("Failed to register Elasticsearch provider: %v", err)
		return
	}

	ctx := context.Background()

	// Try to connect
	if err := elasticProvider.Connect(ctx); err != nil {
		log.Printf("Failed to connect to Elasticsearch: %v", err)
		return
	}

	// Log to Elasticsearch
	manager.LogWithProvider(ctx, "elasticsearch", types.LevelInfo, "This message goes to Elasticsearch", map[string]interface{}{})
	manager.LogWithProvider(ctx, "elasticsearch", types.LevelWarn, "This warning goes to Elasticsearch", map[string]interface{}{
		"warning_type": "PERFORMANCE",
		"threshold":    1000,
		"actual":       1500,
	})

	// Search example (if supported)
	query := types.LogQuery{
		Levels:   []types.LogLevel{types.LevelWarn, types.LevelError},
		Services: []string{"elasticsearch-service"},
		Limit:    10,
	}

	if entries, err := manager.SearchWithProvider(ctx, "elasticsearch", query); err != nil {
		log.Printf("Search failed: %v", err)
	} else {
		fmt.Printf("Found %d log entries\n", len(entries))
	}

	// Get stats
	if stats, err := manager.GetStats(ctx); err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		fmt.Printf("Logging stats: %+v\n", stats)
	}

	// Disconnect
	if err := elasticProvider.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect from Elasticsearch: %v", err)
	}
}

func multipleProvidersExample(manager *gateway.LoggingManager) {
	// Register multiple providers
	consoleConfig := &types.LoggingConfig{
		Level:   types.LevelInfo,
		Format:  types.FormatText,
		Output:  types.OutputStdout,
		Service: "multi-service",
		Version: "1.0.0",
	}

	fileConfig := &types.LoggingConfig{
		Level:   types.LevelDebug,
		Format:  types.FormatJSON,
		Output:  types.OutputFile,
		Service: "multi-service",
		Version: "1.0.0",
		Metadata: map[string]interface{}{
			"file_path": "./logs/multi-service.log",
		},
	}

	consoleProvider := console.NewConsoleProvider(consoleConfig)
	fileProvider := file.NewFileProvider(fileConfig)

	if err := manager.RegisterProvider(consoleProvider); err != nil {
		log.Printf("Failed to register console provider: %v", err)
		return
	}

	if err := manager.RegisterProvider(fileProvider); err != nil {
		log.Printf("Failed to register file provider: %v", err)
		return
	}

	// Connect to file provider
	ctx := context.Background()
	if err := fileProvider.Connect(ctx); err != nil {
		log.Printf("Failed to connect to file provider: %v", err)
		return
	}

	// Log to different providers
	manager.LogWithProvider(ctx, "console", types.LevelInfo, "This goes to console", map[string]interface{}{})
	manager.LogWithProvider(ctx, "file", types.LevelInfo, "This goes to file", map[string]interface{}{})

	// List providers
	providers := manager.ListProviders()
	fmt.Printf("Registered providers: %v\n", providers)

	// Get provider info
	providerInfo := manager.GetProviderInfo()
	for name, info := range providerInfo {
		fmt.Printf("Provider %s: %+v\n", name, info)
	}

	// Disconnect
	if err := fileProvider.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect from file provider: %v", err)
	}
}

func advancedFeaturesExample(manager *gateway.LoggingManager) {
	// Create a provider with advanced features
	consoleConfig := &types.LoggingConfig{
		Level:   types.LevelDebug,
		Format:  types.FormatJSON,
		Output:  types.OutputStdout,
		Service: "advanced-service",
		Version: "1.0.0",
	}

	consoleProvider := console.NewConsoleProvider(consoleConfig)
	if err := manager.RegisterProvider(consoleProvider); err != nil {
		log.Printf("Failed to register console provider: %v", err)
		return
	}

	ctx := context.Background()

	// Test provider features
	features := consoleProvider.GetSupportedFeatures()
	fmt.Printf("Console provider features: %v\n", features)

	// Test connection info
	connInfo := consoleProvider.GetConnectionInfo()
	fmt.Printf("Connection info: %+v\n", connInfo)

	// Test ping
	if err := consoleProvider.Ping(ctx); err != nil {
		log.Printf("Ping failed: %v", err)
	} else {
		fmt.Println("Ping successful")
	}

	// Test with fields
	providerWithFields := consoleProvider.WithFields(ctx, map[string]interface{}{
		"environment": "development",
		"region":      "us-east-1",
	})

	// Log with additional fields
	if err := providerWithFields.Info(ctx, "Message with additional fields", map[string]interface{}{
		"custom_field": "custom_value",
	}); err != nil {
		log.Printf("Failed to log with fields: %v", err)
	}

	// Test with error
	testErr := fmt.Errorf("test error")
	providerWithError := consoleProvider.WithError(ctx, testErr)
	if err := providerWithError.Error(ctx, "Message with error context"); err != nil {
		log.Printf("Failed to log with error: %v", err)
	}

	// Test with context
	ctxWithValues := context.WithValue(ctx, "request_id", "req-123")
	ctxWithValues = context.WithValue(ctxWithValues, "user_id", "user-456")
	ctxWithValues = context.WithValue(ctxWithValues, "tenant_id", "tenant-789")

	providerWithContext := consoleProvider.WithContext(ctxWithValues)
	if err := providerWithContext.Info(ctxWithValues, "Message with context values"); err != nil {
		log.Printf("Failed to log with context: %v", err)
	}

	// Performance test
	start := time.Now()
	for i := 0; i < 100; i++ {
		manager.Info(ctx, fmt.Sprintf("Performance test message %d", i), map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now(),
		})
	}
	duration := time.Since(start)
	fmt.Printf("Logged 100 messages in %v\n", duration)
}
