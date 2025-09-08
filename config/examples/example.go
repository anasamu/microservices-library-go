package main

import (
	"fmt"
	"log"

	"github.com/anasamu/microservices-library-go/config/gateway"
	"github.com/anasamu/microservices-library-go/config/providers/env"
	"github.com/anasamu/microservices-library-go/config/providers/file"
	"github.com/anasamu/microservices-library-go/config/providers/vault"
	"github.com/anasamu/microservices-library-go/config/types"
)

func main() {
	// Create configuration manager
	manager := gateway.NewManager()

	// Example 1: File-based configuration
	fmt.Println("=== File-based Configuration ===")
	fileProvider := file.NewProvider("config.yaml", "yaml")
	manager.RegisterProvider("file", fileProvider)
	manager.SetCurrentProvider("file")

	config, err := manager.Load()
	if err != nil {
		log.Printf("Error loading file config: %v", err)
	} else {
		fmt.Printf("Server Port: %s\n", config.Server.Port)
		fmt.Printf("Database Host: %s\n", config.Database.PostgreSQL.Host)
		fmt.Printf("Environment: %s\n", config.Server.Environment)
	}

	// Example 2: Environment-based configuration
	fmt.Println("\n=== Environment-based Configuration ===")
	envProvider := env.NewProvider("")
	manager.RegisterProvider("env", envProvider)
	manager.SetCurrentProvider("env")

	config, err = manager.Load()
	if err != nil {
		log.Printf("Error loading env config: %v", err)
	} else {
		fmt.Printf("Server Port: %s\n", config.Server.Port)
		fmt.Printf("Database Host: %s\n", config.Database.PostgreSQL.Host)
		fmt.Printf("Environment: %s\n", config.Server.Environment)
	}

	// Example 3: Vault-based configuration
	fmt.Println("\n=== Vault-based Configuration ===")
	vaultProvider, err := vault.NewProvider("http://localhost:8200", "dev-token", "secret/config")
	if err != nil {
		log.Printf("Error creating vault provider: %v", err)
	} else {
		manager.RegisterProvider("vault", vaultProvider)
		manager.SetCurrentProvider("vault")

		config, err = manager.Load()
		if err != nil {
			log.Printf("Error loading vault config: %v", err)
		} else {
			fmt.Printf("Server Port: %s\n", config.Server.Port)
			fmt.Printf("Database Host: %s\n", config.Database.PostgreSQL.Host)
			fmt.Printf("Environment: %s\n", config.Server.Environment)
		}
	}

	// Example 4: Microservices Configuration
	fmt.Println("\n=== Microservices Configuration ===")
	manager.SetCurrentProvider("file")
	config, err = manager.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
	} else {
		// Get user service configuration
		userService, exists := config.GetServiceConfig("user_service")
		if exists {
			fmt.Printf("User Service: %s:%s\n", userService.Host, userService.Port)
			fmt.Printf("Health Check: %s\n", userService.HealthCheck.Path)
			fmt.Printf("Retry Enabled: %t\n", userService.Retry.Enabled)
		}

		// Get service URL
		userServiceURL, err := config.GetServiceURL("user_service")
		if err == nil {
			fmt.Printf("User Service URL: %s\n", userServiceURL)
		}

		// Get health check URL
		healthURL, err := config.GetServiceHealthCheckURL("user_service")
		if err == nil {
			fmt.Printf("Health Check URL: %s\n", healthURL)
		}

		// List all services
		allServices := config.GetAllServices()
		fmt.Printf("Total Services Configured: %d\n", len(allServices))
		for name, service := range allServices {
			fmt.Printf("- %s: %s:%s\n", name, service.Host, service.Port)
		}
	}

	// Example 5: Dynamic Custom Configuration
	fmt.Println("\n=== Dynamic Custom Configuration ===")
	manager.SetCurrentProvider("file")
	config, err = manager.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
	} else {
		// Set custom configuration values
		config.SetCustomValue("feature_flags.new_ui", true)
		config.SetCustomValue("feature_flags.beta_features", false)
		config.SetCustomValue("api_rate_limit", 1000)
		config.SetCustomValue("cache_ttl", 3600)

		// Get custom configuration values
		newUI := config.GetCustomBool("feature_flags.new_ui", false)
		betaFeatures := config.GetCustomBool("feature_flags.beta_features", false)
		rateLimit := config.GetCustomInt("api_rate_limit", 100)
		cacheTTL := config.GetCustomInt("cache_ttl", 300)

		fmt.Printf("New UI Enabled: %t\n", newUI)
		fmt.Printf("Beta Features Enabled: %t\n", betaFeatures)
		fmt.Printf("API Rate Limit: %d\n", rateLimit)
		fmt.Printf("Cache TTL: %d seconds\n", cacheTTL)

		// Add custom service
		customService := types.ServiceConfig{
			Name:     "analytics-service",
			Host:     "analytics.example.com",
			Port:     "8080",
			Protocol: "https",
			HealthCheck: types.HealthCheckConfig{
				Enabled: true,
				Path:    "/health",
			},
			Custom: map[string]interface{}{
				"api_key":    "secret-key",
				"batch_size": 100,
				"enabled":    true,
			},
		}
		config.SetServiceConfig("analytics_service", customService)

		// Get custom service
		analyticsService, exists := config.GetServiceConfig("analytics_service")
		if exists {
			fmt.Printf("Custom Analytics Service: %s\n", analyticsService.Name)
			if apiKey, ok := analyticsService.Custom["api_key"]; ok {
				fmt.Printf("API Key: %s\n", apiKey)
			}
		}
	}

	// Example 6: Configuration watching
	fmt.Println("\n=== Configuration Watching ===")
	manager.SetCurrentProvider("file")

	err = manager.Watch(func(newConfig *types.Config) {
		fmt.Printf("Configuration changed! New port: %s\n", newConfig.Server.Port)

		// Check if custom values changed
		if newUI := newConfig.GetCustomBool("feature_flags.new_ui", false); newUI {
			fmt.Println("New UI feature flag is now enabled!")
		}
	})
	if err != nil {
		log.Printf("Error setting up watcher: %v", err)
	}

	// Example 7: Configuration updates
	fmt.Println("\n=== Configuration Updates ===")
	err = manager.UpdateConfig(func(config *types.Config) {
		config.Server.Port = "9090"
		config.Server.Environment = "production"

		// Update custom values
		config.SetCustomValue("deployment_mode", "production")
		config.SetCustomValue("debug_mode", false)

		// Update service configuration
		if userService, exists := config.GetServiceConfig("user_service"); exists {
			userService.Host = "user-service.prod.example.com"
			userService.Environment = "production"
			config.SetServiceConfig("user_service", *userService)
		}
	})
	if err != nil {
		log.Printf("Error updating config: %v", err)
	} else {
		fmt.Println("Configuration updated successfully")
	}

	// Example 8: List all providers
	fmt.Println("\n=== Available Providers ===")
	providers := manager.ListProviders()
	for _, provider := range providers {
		fmt.Printf("- %s\n", provider)
	}

	// Clean up
	err = manager.Close()
	if err != nil {
		log.Printf("Error closing manager: %v", err)
	}
}
