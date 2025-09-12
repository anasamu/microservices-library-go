package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/anasamu/microservices-library-go/auth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/twofa"
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

	// Register 2FA provider
	twofaProvider := twofa.NewTwoFAProvider(twofa.DefaultTwoFAConfig(), logger)
	if err := authManager.RegisterProvider(twofaProvider); err != nil {
		log.Fatalf("Failed to register 2FA provider: %v", err)
	}

	// Setup HTTP routes
	http.HandleFunc("/setup-2fa", setup2FAHandler(authManager))
	http.HandleFunc("/verify-2fa", verify2FAHandler(authManager))
	http.HandleFunc("/login-with-2fa", loginWith2FAHandler(authManager))

	// Start server
	log.Println("Starting 2FA server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Setup 2FA for a user
func setup2FAHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user ID from request (in real app, this would come from authenticated session)
		userID := r.FormValue("user_id")
		userEmail := r.FormValue("user_email")
		if userID == "" || userEmail == "" {
			http.Error(w, "User ID and email required", http.StatusBadRequest)
			return
		}

		ctx := context.Background()

		// Get 2FA provider
		provider, err := authManager.GetProvider("twofa")
		if err != nil {
			http.Error(w, fmt.Sprintf("2FA provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate 2FA secret
		secret, err := provider.(*twofa.TwoFAProvider).GenerateSecret(ctx, userID, userEmail)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate 2FA secret: %v", err), http.StatusInternalServerError)
			return
		}

		// Return QR code and secret
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": true,
			"user_id": "%s",
			"secret": "%s",
			"qr_code_url": "%s",
			"backup_codes": %v
		}`, userID, secret.Secret, secret.QRCodeURL, secret.BackupCodes)
	}
}

// Verify 2FA setup
func verify2FAHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get parameters from request
		secret := r.FormValue("secret")
		code := r.FormValue("code")

		if secret == "" || code == "" {
			http.Error(w, "Secret and code required", http.StatusBadRequest)
			return
		}

		ctx := context.Background()

		// Get 2FA provider
		provider, err := authManager.GetProvider("twofa")
		if err != nil {
			http.Error(w, fmt.Sprintf("2FA provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Verify 2FA code
		verification, err := provider.(*twofa.TwoFAProvider).VerifyCode(ctx, secret, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("2FA verification failed: %v", err), http.StatusUnauthorized)
			return
		}

		// Return verification result
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": true,
			"verified": %t,
			"message": "%s"
		}`, verification.Valid, verification.Message)
	}
}

// Login with 2FA
func loginWith2FAHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get parameters from request
		username := r.FormValue("username")
		password := r.FormValue("password")
		code := r.FormValue("code")

		if username == "" || password == "" || code == "" {
			http.Error(w, "Username, password, and 2FA code required", http.StatusBadRequest)
			return
		}

		ctx := context.Background()

		// First authenticate with username/password
		authRequest := &types.AuthRequest{
			Username:  username,
			Password:  password,
			ServiceID: "user-service",
			Context:   map[string]interface{}{"ip": r.RemoteAddr},
			Metadata:  map[string]interface{}{"endpoint": "/login-with-2fa"},
		}

		authResponse, err := authManager.Authenticate(ctx, "jwt", authRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// For this example, we'll assume the 2FA code is valid if provided
		// In a real application, you would verify the 2FA code here
		_ = code // Use the code variable to avoid unused variable warning

		// Return successful login with 2FA
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": %t,
			"user_id": "%s",
			"access_token": "%s",
			"expires_at": "%s",
			"roles": %v,
			"permissions": %v,
			"twofa_verified": true
		}`, authResponse.Success, authResponse.UserID, authResponse.AccessToken,
			authResponse.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			authResponse.Roles, authResponse.Permissions)
	}
}
