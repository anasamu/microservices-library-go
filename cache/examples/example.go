package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/cache/gateway"
	"github.com/anasamu/microservices-library-go/cache/providers/memory"
	"github.com/anasamu/microservices-library-go/cache/providers/redis"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create cache manager
	manager := gateway.NewCacheManager(&gateway.ManagerConfig{
		DefaultProvider: "memory",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}, logger)

	// Register memory provider
	memoryProvider := memory.NewMemoryProvider(&memory.MemoryConfig{
		MaxSize:         1000,
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		KeyPrefix:       "app",
		Namespace:       "example",
	}, logger)

	if err := manager.RegisterProvider(memoryProvider); err != nil {
		log.Fatalf("Failed to register memory provider: %v", err)
	}

	// Register Redis provider (optional - only if Redis is available)
	redisProvider := redis.NewRedisProvider(&redis.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		KeyPrefix:    "app",
		Namespace:    "example",
	}, logger)

	if err := manager.RegisterProvider(redisProvider); err != nil {
		logger.WithError(err).Warn("Failed to register Redis provider")
	}

	// Connect providers
	ctx := context.Background()

	if err := memoryProvider.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect memory provider: %v", err)
	}

	// Try to connect Redis (optional)
	if err := redisProvider.Connect(ctx); err != nil {
		logger.WithError(err).Warn("Failed to connect Redis provider")
	}

	// Example usage
	fmt.Println("=== Cache Example ===")

	// Basic operations
	fmt.Println("\n1. Basic Cache Operations:")

	// Set a value
	user := map[string]interface{}{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	if err := manager.Set(ctx, "user:1", user, 10*time.Minute); err != nil {
		log.Printf("Failed to set value: %v", err)
	} else {
		fmt.Println("✓ Set user:1")
	}

	// Get a value
	var retrievedUser map[string]interface{}
	if err := manager.Get(ctx, "user:1", &retrievedUser); err != nil {
		log.Printf("Failed to get value: %v", err)
	} else {
		fmt.Printf("✓ Retrieved user: %+v\n", retrievedUser)
	}

	// Check if key exists
	exists, err := manager.Exists(ctx, "user:1")
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
	} else {
		fmt.Printf("✓ Key exists: %v\n", exists)
	}

	// Set with tags
	fmt.Println("\n2. Tagged Cache Operations:")

	products := []map[string]interface{}{
		{"id": 1, "name": "Product 1", "category": "electronics"},
		{"id": 2, "name": "Product 2", "category": "electronics"},
		{"id": 3, "name": "Product 3", "category": "books"},
	}

	for _, product := range products {
		key := fmt.Sprintf("product:%d", product["id"])
		if err := manager.SetWithTags(ctx, key, product, 15*time.Minute, []string{"products", product["category"].(string)}); err != nil {
			log.Printf("Failed to set product %v: %v", product["id"], err)
		} else {
			fmt.Printf("✓ Set %s with tags\n", key)
		}
	}

	// Invalidate by tag
	if err := manager.InvalidateByTag(ctx, "electronics"); err != nil {
		log.Printf("Failed to invalidate by tag: %v", err)
	} else {
		fmt.Println("✓ Invalidated all electronics products")
	}

	// Batch operations
	fmt.Println("\n3. Batch Operations:")

	batchItems := map[string]interface{}{
		"config:theme":    "dark",
		"config:language": "en",
		"config:timezone": "UTC",
		"config:currency": "USD",
	}

	if err := manager.SetMultiple(ctx, batchItems, 30*time.Minute); err != nil {
		log.Printf("Failed to set multiple values: %v", err)
	} else {
		fmt.Println("✓ Set multiple configuration values")
	}

	// Get multiple values
	keys := []string{"config:theme", "config:language", "config:timezone"}
	results, err := manager.GetMultiple(ctx, keys)
	if err != nil {
		log.Printf("Failed to get multiple values: %v", err)
	} else {
		fmt.Printf("✓ Retrieved multiple values: %+v\n", results)
	}

	// Provider information
	fmt.Println("\n4. Provider Information:")

	providers := manager.ListProviders()
	fmt.Printf("Registered providers: %v\n", providers)

	providerInfo := manager.GetProviderInfo()
	for name, info := range providerInfo {
		fmt.Printf("Provider: %s\n", name)
		fmt.Printf("  - Connected: %v\n", info.IsConnected)
		fmt.Printf("  - Features: %v\n", info.SupportedFeatures)
		fmt.Printf("  - Host: %s:%d\n", info.ConnectionInfo.Host, info.ConnectionInfo.Port)
	}

	// Statistics
	fmt.Println("\n5. Cache Statistics:")

	stats, err := manager.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		for provider, stat := range stats {
			fmt.Printf("Provider: %s\n", provider)
			fmt.Printf("  - Hits: %d\n", stat.Hits)
			fmt.Printf("  - Misses: %d\n", stat.Misses)
			fmt.Printf("  - Keys: %d\n", stat.Keys)
			fmt.Printf("  - Memory: %d bytes\n", stat.Memory)
			fmt.Printf("  - Uptime: %v\n", stat.Uptime)
		}
	}

	// TTL operations
	fmt.Println("\n6. TTL Operations:")

	// Set a value with TTL
	if err := manager.Set(ctx, "temp:key", "temporary value", 1*time.Minute); err != nil {
		log.Printf("Failed to set temp key: %v", err)
	} else {
		fmt.Println("✓ Set temporary key with 1 minute TTL")
	}

	// Get TTL
	ttl, err := manager.GetTTL(ctx, "temp:key")
	if err != nil {
		log.Printf("Failed to get TTL: %v", err)
	} else {
		fmt.Printf("✓ TTL remaining: %v\n", ttl)
	}

	// Update TTL
	if err := manager.SetTTL(ctx, "temp:key", 2*time.Minute); err != nil {
		log.Printf("Failed to update TTL: %v", err)
	} else {
		fmt.Println("✓ Updated TTL to 2 minutes")
	}

	// Key enumeration (if supported)
	fmt.Println("\n7. Key Operations:")

	keys, err = manager.GetKeys(ctx, "config:*")
	if err != nil {
		log.Printf("Failed to get keys: %v", err)
	} else {
		fmt.Printf("✓ Found keys matching pattern: %v\n", keys)
	}

	// Cleanup
	fmt.Println("\n8. Cleanup:")

	// Delete specific keys
	if err := manager.Delete(ctx, "user:1"); err != nil {
		log.Printf("Failed to delete user:1: %v", err)
	} else {
		fmt.Println("✓ Deleted user:1")
	}

	// Delete multiple keys
	deleteKeys := []string{"config:theme", "config:language"}
	if err := manager.DeleteMultiple(ctx, deleteKeys); err != nil {
		log.Printf("Failed to delete multiple keys: %v", err)
	} else {
		fmt.Println("✓ Deleted multiple configuration keys")
	}

	// Close connections
	if err := manager.Close(); err != nil {
		log.Printf("Failed to close manager: %v", err)
	} else {
		fmt.Println("✓ Closed all connections")
	}

	fmt.Println("\n=== Example completed ===")
}
