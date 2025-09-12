package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anasamu/microservices-library-go/auth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create auth manager
	authManager := auth.NewAuthManager(auth.DefaultManagerConfig(), logger)

	// Register JWT provider
	jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
	if err := authManager.RegisterProvider(jwtProvider); err != nil {
		log.Fatalf("Failed to register JWT provider: %v", err)
	}

	ctx := context.Background()

	// Example 1: Authenticate user
	fmt.Println("=== Authenticating User ===")
	authRequest := &types.AuthRequest{
		Username:  "john_doe",
		Password:  "secure_password",
		Email:     "john@example.com",
		ServiceID: "user-service",
		Context:   map[string]interface{}{"ip": "192.168.1.1"},
		Metadata:  map[string]interface{}{"endpoint": "/login"},
	}

	authResponse, err := authManager.Authenticate(ctx, "jwt", authRequest)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
	} else {
		fmt.Printf("Authentication successful: %+v\n", authResponse)
	}

	// Example 2: Validate token
	if authResponse != nil {
		fmt.Println("\n=== Validating Token ===")
		tokenRequest := &types.TokenValidationRequest{
			Token:    authResponse.AccessToken,
			Metadata: map[string]interface{}{"service": "user-service"},
		}

		tokenResponse, err := authManager.ValidateToken(ctx, "jwt", tokenRequest)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
		} else {
			fmt.Printf("Token validation result: %+v\n", tokenResponse)
		}

		// Example 3: Check permission
		if tokenResponse != nil {
			fmt.Println("\n=== Checking Permission ===")
			permissionRequest := &types.PermissionRequest{
				UserID:     tokenResponse.UserID,
				Permission: "write",
				Resource:   "user-data",
				Context:    map[string]interface{}{"service": "user-service"},
				Metadata:   map[string]interface{}{"endpoint": "/protected"},
			}

			permissionResponse, err := authManager.CheckPermission(ctx, "jwt", permissionRequest)
			if err != nil {
				log.Printf("Permission check failed: %v", err)
			} else {
				fmt.Printf("Permission check result: %+v\n", permissionResponse)
			}
		}
	}

	// Example 4: Health check
	fmt.Println("\n=== Health Check ===")
	healthResults := authManager.HealthCheck(ctx)
	fmt.Printf("Health check results: %+v\n", healthResults)

	// Example 5: Get statistics
	fmt.Println("\n=== Statistics ===")
	stats := authManager.GetStats(ctx)
	fmt.Printf("Auth statistics: %+v\n", stats)
}
