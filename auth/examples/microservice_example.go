package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/auth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/oauth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/twofa"
	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create auth manager
	authManager := gateway.NewAuthManager(gateway.DefaultManagerConfig(), logger)

	// Register JWT provider
	jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
	if err := authManager.RegisterProvider(jwtProvider); err != nil {
		log.Fatalf("Failed to register JWT provider: %v", err)
	}

	// Register OAuth provider
	oauthConfigs := map[string]*oauth.OAuthConfig{
		"google": {
			ClientID:     "your-google-client-id",
			ClientSecret: "your-google-client-secret",
			RedirectURL:  "http://localhost:8080/auth/callback/google",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "google",
		},
	}
	oauthProvider := oauth.NewOAuthProvider(oauthConfigs, logger)
	if err := authManager.RegisterProvider(oauthProvider); err != nil {
		log.Fatalf("Failed to register OAuth provider: %v", err)
	}

	// Register 2FA provider
	twofaProvider := twofa.NewTwoFAProvider(twofa.DefaultTwoFAConfig(), logger)
	if err := authManager.RegisterProvider(twofaProvider); err != nil {
		log.Fatalf("Failed to register 2FA provider: %v", err)
	}

	// Setup routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/login", loginHandler(authManager))
	http.HandleFunc("/protected", protectedHandler)
	http.HandleFunc("/admin", adminHandler)

	// Start server
	log.Println("Starting microservice on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

// Login endpoint
func loginHandler(authManager *gateway.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse login request
		username := r.FormValue("username")
		password := r.FormValue("password")
		serviceID := r.FormValue("service_id")

		if username == "" || password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		// Create auth request
		authRequest := &types.AuthRequest{
			Username:  username,
			Password:  password,
			ServiceID: serviceID,
			Context: map[string]interface{}{
				"ip_address": r.RemoteAddr,
				"user_agent": r.UserAgent(),
			},
			Metadata: map[string]interface{}{
				"endpoint": "/login",
			},
		}

		// Authenticate user
		ctx := context.Background()
		response, err := authManager.Authenticate(ctx, "jwt", authRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": %t,
			"user_id": "%s",
			"access_token": "%s",
			"expires_at": "%s",
			"roles": %v,
			"permissions": %v
		}`, response.Success, response.UserID, response.AccessToken,
			response.ExpiresAt.Format(time.RFC3339), response.Roles, response.Permissions)
	}
}

// Protected endpoint
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
		"message": "Access granted to protected resource",
		"timestamp": "%s"
	}`, time.Now().Format(time.RFC3339))
}

// Admin endpoint
func adminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
		"message": "Access granted to admin resource",
		"timestamp": "%s"
	}`, time.Now().Format(time.RFC3339))
}

// Example of using the auth library in a microservice
func exampleUsage() {
	logger := logrus.New()

	// Create auth manager
	authManager := gateway.NewAuthManager(gateway.DefaultManagerConfig(), logger)

	// Register providers
	jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
	authManager.RegisterProvider(jwtProvider)

	// Example: Create a user
	ctx := context.Background()
	createUserRequest := &types.CreateUserRequest{
		Username:    "john_doe",
		Email:       "john@example.com",
		Password:    "secure_password",
		FirstName:   "John",
		LastName:    "Doe",
		Roles:       []string{"user"},
		Permissions: []string{"read", "write"},
		Metadata:    map[string]interface{}{"source": "registration"},
	}

	userResponse, err := authManager.CreateUser(ctx, "jwt", createUserRequest)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	log.Printf("User created: %+v", userResponse)

	// Example: Authenticate user
	authRequest := &types.AuthRequest{
		Username:  "john_doe",
		Password:  "secure_password",
		ServiceID: "user-service",
		Context:   map[string]interface{}{"ip": "192.168.1.1"},
		Metadata:  map[string]interface{}{"endpoint": "/login"},
	}

	authResponse, err := authManager.Authenticate(ctx, "jwt", authRequest)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return
	}

	log.Printf("Authentication successful: %+v", authResponse)

	// Example: Validate token
	tokenRequest := &types.TokenValidationRequest{
		Token:    authResponse.AccessToken,
		Metadata: map[string]interface{}{"service": "user-service"},
	}

	tokenResponse, err := authManager.ValidateToken(ctx, "jwt", tokenRequest)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return
	}

	log.Printf("Token validation result: %+v", tokenResponse)

	// Example: Check permission
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
		return
	}

	log.Printf("Permission check result: %+v", permissionResponse)

	// Example: Health check
	healthResults := authManager.HealthCheck(ctx)
	log.Printf("Health check results: %+v", healthResults)

	// Example: Get statistics
	stats := authManager.GetStats(ctx)
	log.Printf("Auth statistics: %+v", stats)
}
