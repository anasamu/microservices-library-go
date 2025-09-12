package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/anasamu/microservices-library-go/auth"
	"github.com/anasamu/microservices-library-go/auth/providers/authentication/oauth"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create auth manager
	authManager := auth.NewAuthManager(auth.DefaultManagerConfig(), logger)

	// Configure OAuth providers
	oauthConfigs := map[string]*oauth.OAuthConfig{
		"google": {
			ClientID:     "your-google-client-id",
			ClientSecret: "your-google-client-secret",
			RedirectURL:  "http://localhost:8080/auth/callback/google",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "google",
		},
		"github": {
			ClientID:     "your-github-client-id",
			ClientSecret: "your-github-client-secret",
			RedirectURL:  "http://localhost:8080/auth/callback/github",
			Scopes:       []string{"user:email"},
			Provider:     "github",
		},
	}

	// Register OAuth provider
	oauthProvider := oauth.NewOAuthProvider(oauthConfigs, logger)
	if err := authManager.RegisterProvider(oauthProvider); err != nil {
		log.Fatalf("Failed to register OAuth provider: %v", err)
	}

	// Setup HTTP routes
	http.HandleFunc("/auth/google", googleAuthHandler(authManager))
	http.HandleFunc("/auth/github", githubAuthHandler(authManager))
	http.HandleFunc("/auth/callback/google", googleCallbackHandler(authManager))
	http.HandleFunc("/auth/callback/github", githubCallbackHandler(authManager))

	// Start server
	log.Println("Starting OAuth server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Google OAuth initiation
func googleAuthHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Get OAuth provider
		provider, err := authManager.GetProvider("oauth")
		if err != nil {
			http.Error(w, fmt.Sprintf("OAuth provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Get authorization URL
		state := "random-state-string" // In production, use a secure random string
		authURL, err := provider.(*oauth.OAuthProvider).GetAuthURL(ctx, "google", state)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get auth URL: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to OAuth provider
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

// GitHub OAuth initiation
func githubAuthHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Get OAuth provider
		provider, err := authManager.GetProvider("oauth")
		if err != nil {
			http.Error(w, fmt.Sprintf("OAuth provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Get authorization URL
		state := "random-state-string" // In production, use a secure random string
		authURL, err := provider.(*oauth.OAuthProvider).GetAuthURL(ctx, "github", state)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get auth URL: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to OAuth provider
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

// Google OAuth callback
func googleCallbackHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Get authorization code from query parameters
		code := r.URL.Query().Get("code")
		
		if code == "" {
			http.Error(w, "Authorization code not provided", http.StatusBadRequest)
			return
		}

		// Get OAuth provider
		provider, err := authManager.GetProvider("oauth")
		if err != nil {
			http.Error(w, fmt.Sprintf("OAuth provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Exchange code for token
		tokenResponse, err := provider.(*oauth.OAuthProvider).ExchangeCode(ctx, "google", code)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
			return
		}

		// Get user info
		userInfo, err := provider.(*oauth.OAuthProvider).ValidateOAuthToken(ctx, "google", tokenResponse.AccessToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": true,
			"access_token": "%s",
			"expires_at": "%s",
			"provider": "google",
			"user_info": {
				"id": "%s",
				"email": "%s",
				"name": "%s",
				"picture": "%s"
			}
		}`, tokenResponse.AccessToken, tokenResponse.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
	}
}

// GitHub OAuth callback
func githubCallbackHandler(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Get authorization code from query parameters
		code := r.URL.Query().Get("code")
		
		if code == "" {
			http.Error(w, "Authorization code not provided", http.StatusBadRequest)
			return
		}

		// Get OAuth provider
		provider, err := authManager.GetProvider("oauth")
		if err != nil {
			http.Error(w, fmt.Sprintf("OAuth provider not found: %v", err), http.StatusInternalServerError)
			return
		}

		// Exchange code for token
		tokenResponse, err := provider.(*oauth.OAuthProvider).ExchangeCode(ctx, "github", code)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
			return
		}

		// Get user info
		userInfo, err := provider.(*oauth.OAuthProvider).ValidateOAuthToken(ctx, "github", tokenResponse.AccessToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"success": true,
			"access_token": "%s",
			"expires_at": "%s",
			"provider": "github",
			"user_info": {
				"id": "%s",
				"email": "%s",
				"name": "%s",
				"picture": "%s"
			}
		}`, tokenResponse.AccessToken, tokenResponse.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
	}
}
