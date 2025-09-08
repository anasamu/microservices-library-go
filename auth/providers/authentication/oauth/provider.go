package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/linkedin"
	"golang.org/x/oauth2/microsoft"
)

// OAuthProvider implements OAuth2-based authentication
type OAuthProvider struct {
	configs    map[string]*oauth2.Config
	logger     *logrus.Logger
	configured bool
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	Provider     string   `json:"provider"` // google, microsoft, github, etc.
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
	Provider      string `json:"provider"`
}

// OAuthTokenResponse represents OAuth token response
type OAuthTokenResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time      `json:"expires_at"`
	TokenType    string         `json:"token_type"`
	Scope        string         `json:"scope"`
	UserInfo     *OAuthUserInfo `json:"user_info"`
}

// DefaultOAuthConfig returns default OAuth configuration
func DefaultOAuthConfig() map[string]*OAuthConfig {
	return map[string]*OAuthConfig{
		"google": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/google",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "google",
		},
		"microsoft": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/microsoft",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "microsoft",
		},
		"github": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/github",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "github",
		},
		"linkedin": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/linkedin",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "linkedin",
		},
		"twitter": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/twitter",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "twitter",
		},
		"facebook": {
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback/facebook",
			Scopes:       []string{"openid", "profile", "email"},
			Provider:     "facebook",
		},
	}
}

// NewOAuthProvider creates a new OAuth provider
func NewOAuthProvider(configs map[string]*OAuthConfig, logger *logrus.Logger) *OAuthProvider {
	if logger == nil {
		logger = logrus.New()
	}

	if configs == nil {
		configs = DefaultOAuthConfig()
	}

	provider := &OAuthProvider{
		configs:    make(map[string]*oauth2.Config),
		logger:     logger,
		configured: true,
	}

	// Initialize OAuth2 configs for each provider
	for providerName, config := range configs {
		provider.configs[providerName] = provider.createOAuth2Config(config)
	}

	return provider
}

// createOAuth2Config creates OAuth2 configuration for a specific provider
func (op *OAuthProvider) createOAuth2Config(config *OAuthConfig) *oauth2.Config {
	baseConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
	}

	switch config.Provider {
	case "google":
		baseConfig.Endpoint = google.Endpoint
	case "microsoft":
		baseConfig.Endpoint = microsoft.AzureADEndpoint("common")
	case "github":
		baseConfig.Endpoint = github.Endpoint
	case "linkedin":
		baseConfig.Endpoint = linkedin.Endpoint
	case "twitter":
		baseConfig.Endpoint = twitter.Endpoint
	case "facebook":
		baseConfig.Endpoint = facebook.Endpoint
	default:
		// Custom provider - endpoints should be set manually
		baseConfig.Endpoint = oauth2.Endpoint{}
	}

	return baseConfig
}

// GetName returns the provider name
func (op *OAuthProvider) GetName() string {
	return "oauth"
}

// GetSupportedFeatures returns supported features
func (op *OAuthProvider) GetSupportedFeatures() []string {
	return []string{
		"oauth2",
		"google_oauth",
		"microsoft_oauth",
		"github_oauth",
		"linkedin_oauth",
		"twitter_oauth",
		"facebook_oauth",
		"token_exchange",
		"user_info",
		"refresh_token",
	}
}

// GetConnectionInfo returns connection information
func (op *OAuthProvider) GetConnectionInfo() map[string]interface{} {
	providers := make([]string, 0, len(op.configs))
	for provider := range op.configs {
		providers = append(providers, provider)
	}

	return map[string]interface{}{
		"type":      "oauth2",
		"providers": providers,
	}
}

// Configure configures the OAuth provider
func (op *OAuthProvider) Configure(config map[string]interface{}) error {
	// Parse configuration for each provider
	for providerName, configData := range config {
		if configMap, ok := configData.(map[string]interface{}); ok {
			oauthConfig := &OAuthConfig{
				Provider: providerName,
			}

			if clientID, ok := configMap["client_id"].(string); ok {
				oauthConfig.ClientID = clientID
			}
			if clientSecret, ok := configMap["client_secret"].(string); ok {
				oauthConfig.ClientSecret = clientSecret
			}
			if redirectURL, ok := configMap["redirect_url"].(string); ok {
				oauthConfig.RedirectURL = redirectURL
			}
			if scopes, ok := configMap["scopes"].([]string); ok {
				oauthConfig.Scopes = scopes
			}

			op.configs[providerName] = op.createOAuth2Config(oauthConfig)
		}
	}

	op.configured = true
	op.logger.Info("OAuth provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (op *OAuthProvider) IsConfigured() bool {
	return op.configured
}

// GetAuthURL generates authorization URL for OAuth flow
func (op *OAuthProvider) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	config, exists := op.configs[providerName]
	if !exists {
		return "", fmt.Errorf("OAuth provider not configured: %s", providerName)
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	op.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"state":    state,
	}).Debug("Generated OAuth authorization URL")

	return authURL, nil
}

// ExchangeCode exchanges authorization code for tokens
func (op *OAuthProvider) ExchangeCode(ctx context.Context, providerName, code string) (*OAuthTokenResponse, error) {
	config, exists := op.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OAuth provider not configured: %s", providerName)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info
	userInfo, err := op.getUserInfo(ctx, providerName, token.AccessToken)
	if err != nil {
		op.logger.WithError(err).Warn("Failed to get user info")
		// Continue without user info
	}

	response := &OAuthTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		Scope:        token.Extra("scope").(string),
		UserInfo:     userInfo,
	}

	op.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"user_id":    userInfo.ID,
		"expires_at": token.Expiry,
	}).Info("OAuth token exchange completed")

	return response, nil
}

// RefreshToken refreshes an OAuth token
func (op *OAuthProvider) RefreshToken(ctx context.Context, providerName, refreshToken string) (*OAuthTokenResponse, error) {
	config, exists := op.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OAuth provider not configured: %s", providerName)
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user info
	userInfo, err := op.getUserInfo(ctx, providerName, newToken.AccessToken)
	if err != nil {
		op.logger.WithError(err).Warn("Failed to get user info")
	}

	response := &OAuthTokenResponse{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresAt:    newToken.Expiry,
		TokenType:    newToken.TokenType,
		Scope:        newToken.Extra("scope").(string),
		UserInfo:     userInfo,
	}

	op.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"expires_at": newToken.Expiry,
	}).Info("OAuth token refreshed")

	return response, nil
}

// getUserInfo retrieves user information from OAuth provider
func (op *OAuthProvider) getUserInfo(ctx context.Context, providerName, accessToken string) (*OAuthUserInfo, error) {
	var userInfoURL string

	switch providerName {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	case "github":
		userInfoURL = "https://api.github.com/user"
	case "linkedin":
		userInfoURL = "https://api.linkedin.com/v2/me"
	case "twitter":
		userInfoURL = "https://api.twitter.com/1.1/account/verify_credentials.json"
	case "facebook":
		userInfoURL = "https://graph.facebook.com/me"
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", providerName)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo OAuthUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	userInfo.Provider = providerName
	return &userInfo, nil
}

// ValidateToken validates an OAuth token
func (op *OAuthProvider) ValidateToken(ctx context.Context, providerName, accessToken string) (*OAuthUserInfo, error) {
	userInfo, err := op.getUserInfo(ctx, providerName, accessToken)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	op.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  userInfo.ID,
		"email":    userInfo.Email,
	}).Debug("OAuth token validated")

	return userInfo, nil
}

// HealthCheck performs health check
func (op *OAuthProvider) HealthCheck(ctx context.Context) error {
	if !op.configured {
		return fmt.Errorf("OAuth provider not configured")
	}

	// Check if at least one provider is configured
	if len(op.configs) == 0 {
		return fmt.Errorf("no OAuth providers configured")
	}

	return nil
}

// GetStats returns provider statistics
func (op *OAuthProvider) GetStats(ctx context.Context) map[string]interface{} {
	providers := make([]string, 0, len(op.configs))
	for provider := range op.configs {
		providers = append(providers, provider)
	}

	return map[string]interface{}{
		"provider":   "oauth",
		"configured": op.configured,
		"providers":  providers,
		"count":      len(op.configs),
	}
}

// Close closes the provider
func (op *OAuthProvider) Close() error {
	op.logger.Info("OAuth provider closed")
	return nil
}

// CreateUserFromOAuth creates a user from OAuth user info
func (op *OAuthProvider) CreateUserFromOAuth(ctx context.Context, userInfo *OAuthUserInfo) (string, error) {
	// Generate a unique user ID
	userID := uuid.New().String()

	op.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"email":    userInfo.Email,
		"name":     userInfo.Name,
		"provider": userInfo.Provider,
	}).Info("User created from OAuth")

	return userID, nil
}
