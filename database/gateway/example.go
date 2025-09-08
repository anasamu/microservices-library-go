package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the database gateway system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create database manager
	config := DefaultManagerConfig()
	config.DefaultProvider = "postgresql"
	config.MaxConnections = 100
	config.RetryAttempts = 3
	config.RetryDelay = 5 * time.Second

	databaseManager := NewDatabaseManager(config, logger)

	// Example: Basic gateway operations
	ctx := context.Background()

	// Get supported providers (will be empty until providers are registered)
	providers := databaseManager.GetSupportedProviders()
	fmt.Printf("Supported providers: %v\n", providers)

	// Health check (will be empty until providers are registered)
	results := databaseManager.HealthCheck(ctx)
	fmt.Printf("Health check results: %v\n", results)

	// Get connected providers (will be empty until providers are connected)
	connected := databaseManager.GetConnectedProviders()
	fmt.Printf("Connected providers: %v\n", connected)

	fmt.Println("Database gateway example completed. See examples/ directory for full usage examples.")
}
