package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit/gateway"
	"github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory"
	"github.com/anasamu/microservices-library-go/ratelimit/providers/redis"
	"github.com/anasamu/microservices-library-go/ratelimit/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create rate limit manager
	config := &gateway.ManagerConfig{
		DefaultProvider: "inmemory",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewRateLimitManager(config, logger)

	// Register in-memory provider
	inmemoryProvider, err := inmemory.NewProvider(&inmemory.Config{
		CleanupInterval: time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to create in-memory provider: %v", err)
	}

	if err := manager.RegisterProvider(inmemoryProvider); err != nil {
		log.Fatalf("Failed to register in-memory provider: %v", err)
	}

	// Register Redis provider (optional)
	redisProvider, err := redis.NewProvider(&redis.Config{
		Host:     "localhost",
		Port:     6379,
		Database: 0,
		Password: "",
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to create Redis provider, continuing with in-memory only")
	} else {
		if err := manager.RegisterProvider(redisProvider); err != nil {
			logger.WithError(err).Warn("Failed to register Redis provider, continuing with in-memory only")
		}
	}

	ctx := context.Background()

	// Example 1: Basic rate limiting
	fmt.Println("=== Basic Rate Limiting Example ===")
	basicRateLimit := &types.RateLimit{
		Limit:     10,
		Window:    time.Minute,
		Algorithm: types.AlgorithmTokenBucket,
	}

	key := "user:123"
	for i := 0; i < 12; i++ {
		result, err := manager.Allow(ctx, key, basicRateLimit)
		if err != nil {
			log.Printf("Error checking rate limit: %v", err)
			continue
		}

		fmt.Printf("Request %d: Allowed=%v, Remaining=%d, ResetTime=%v\n",
			i+1, result.Allowed, result.Remaining, result.ResetTime.Format(time.RFC3339))

		if !result.Allowed {
			fmt.Printf("Rate limit exceeded! Retry after: %v\n", result.RetryAfter)
			break
		}
	}

	// Example 2: Different algorithms
	fmt.Println("\n=== Different Algorithms Example ===")
	algorithms := []types.Algorithm{
		types.AlgorithmTokenBucket,
		types.AlgorithmSlidingWindow,
		types.AlgorithmFixedWindow,
	}

	for _, algorithm := range algorithms {
		fmt.Printf("\nTesting %s algorithm:\n", algorithm)
		limit := &types.RateLimit{
			Limit:     5,
			Window:    10 * time.Second,
			Algorithm: algorithm,
		}

		testKey := fmt.Sprintf("test:%s", algorithm)
		for i := 0; i < 7; i++ {
			result, err := manager.Allow(ctx, testKey, limit)
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}

			fmt.Printf("  Request %d: Allowed=%v, Remaining=%d\n",
				i+1, result.Allowed, result.Remaining)

			if !result.Allowed {
				break
			}
		}
	}

	// Example 3: Batch operations
	fmt.Println("\n=== Batch Operations Example ===")
	requests := []*types.RateLimitRequest{
		{
			Key: "batch:user1",
			Limit: &types.RateLimit{
				Limit:     3,
				Window:    time.Minute,
				Algorithm: types.AlgorithmTokenBucket,
			},
		},
		{
			Key: "batch:user2",
			Limit: &types.RateLimit{
				Limit:     3,
				Window:    time.Minute,
				Algorithm: types.AlgorithmTokenBucket,
			},
		},
		{
			Key: "batch:user3",
			Limit: &types.RateLimit{
				Limit:     3,
				Window:    time.Minute,
				Algorithm: types.AlgorithmTokenBucket,
			},
		},
	}

	results, err := manager.AllowMultiple(ctx, requests)
	if err != nil {
		log.Printf("Error in batch operation: %v", err)
	} else {
		for i, result := range results {
			fmt.Printf("Batch request %d (%s): Allowed=%v, Remaining=%d\n",
				i+1, result.Key, result.Allowed, result.Remaining)
		}
	}

	// Example 4: Get remaining requests
	fmt.Println("\n=== Get Remaining Requests Example ===")
	remaining, err := manager.GetRemaining(ctx, key, basicRateLimit)
	if err != nil {
		log.Printf("Error getting remaining requests: %v", err)
	} else {
		fmt.Printf("Remaining requests for %s: %d\n", key, remaining)
	}

	// Example 5: Reset rate limit
	fmt.Println("\n=== Reset Rate Limit Example ===")
	if err := manager.Reset(ctx, key); err != nil {
		log.Printf("Error resetting rate limit: %v", err)
	} else {
		fmt.Printf("Rate limit reset for %s\n", key)
	}

	// Check remaining after reset
	remaining, err = manager.GetRemaining(ctx, key, basicRateLimit)
	if err != nil {
		log.Printf("Error getting remaining requests after reset: %v", err)
	} else {
		fmt.Printf("Remaining requests after reset: %d\n", remaining)
	}

	// Example 6: Get statistics
	fmt.Println("\n=== Statistics Example ===")
	stats, err := manager.GetStats(ctx)
	if err != nil {
		log.Printf("Error getting statistics: %v", err)
	} else {
		for provider, stat := range stats {
			fmt.Printf("Provider %s: Total=%d, Allowed=%d, Blocked=%d, ActiveKeys=%d\n",
				provider, stat.TotalRequests, stat.AllowedRequests, stat.BlockedRequests, stat.ActiveKeys)
		}
	}

	// Example 7: Provider information
	fmt.Println("\n=== Provider Information ===")
	providerInfo := manager.GetProviderInfo()
	for name, info := range providerInfo {
		fmt.Printf("Provider: %s\n", name)
		fmt.Printf("  Connected: %v\n", info.IsConnected)
		fmt.Printf("  Features: %v\n", info.SupportedFeatures)
		if info.ConnectionInfo != nil {
			fmt.Printf("  Host: %s:%d\n", info.ConnectionInfo.Host, info.ConnectionInfo.Port)
		}
	}

	// Clean up
	if err := manager.Close(); err != nil {
		log.Printf("Error closing manager: %v", err)
	}

	fmt.Println("\nExample completed successfully!")
}
