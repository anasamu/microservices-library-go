package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/auth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/sirupsen/logrus"
)

type AuthServer struct {
	authManager *auth.AuthManager
	logger      *logrus.Logger
}

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

	// Create server instance
	server := &AuthServer{
		authManager: authManager,
		logger:      logger,
	}

	// Setup routes
	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/register", server.registerHandler)
	http.HandleFunc("/login", server.loginHandler)
	http.HandleFunc("/validate", server.validateHandler)
	http.HandleFunc("/protected", server.protectedHandler)

	// Start server
	log.Println("Starting auth server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Health check endpoint
func (s *AuthServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "auth-service",
	}

	json.NewEncoder(w).Encode(response)
}

// Register endpoint (simplified - just returns success message)
func (s *AuthServer) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username    string   `json:"username"`
		Email       string   `json:"email"`
		Password    string   `json:"password"`
		FirstName   string   `json:"first_name"`
		LastName    string   `json:"last_name"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// In a real application, you would store user data in a database
	// For this example, we'll just return a success message
	response := map[string]interface{}{
		"success":  true,
		"message":  "User registration successful",
		"username": req.Username,
		"email":    req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login endpoint
func (s *AuthServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		ServiceID string `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create auth request
	authRequest := &types.AuthRequest{
		Username:  req.Username,
		Password:  req.Password,
		ServiceID: req.ServiceID,
		Context: map[string]interface{}{
			"ip_address": r.RemoteAddr,
			"user_agent": r.UserAgent(),
		},
		Metadata: map[string]interface{}{
			"endpoint": "/login",
		},
	}

	ctx := context.Background()
	response, err := s.authManager.Authenticate(ctx, "jwt", authRequest)
	if err != nil {
		s.logger.WithError(err).Error("Authentication failed")
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Token validation endpoint
func (s *AuthServer) validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create token validation request
	tokenRequest := &types.TokenValidationRequest{
		Token:    req.Token,
		Metadata: map[string]interface{}{"service": "auth-service"},
	}

	ctx := context.Background()
	response, err := s.authManager.ValidateToken(ctx, "jwt", tokenRequest)
	if err != nil {
		s.logger.WithError(err).Error("Token validation failed")
		http.Error(w, fmt.Sprintf("Token validation failed: %v", err), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Protected endpoint
func (s *AuthServer) protectedHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token (assuming "Bearer <token>" format)
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]

	// Validate token
	tokenRequest := &types.TokenValidationRequest{
		Token:    token,
		Metadata: map[string]interface{}{"service": "auth-service"},
	}

	ctx := context.Background()
	response, err := s.authManager.ValidateToken(ctx, "jwt", tokenRequest)
	if err != nil {
		s.logger.WithError(err).Error("Token validation failed")
		http.Error(w, fmt.Sprintf("Token validation failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Return protected resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseData := map[string]interface{}{
		"message":   "Access granted to protected resource",
		"user_id":   response.UserID,
		"valid":     response.Valid,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(responseData)
}
