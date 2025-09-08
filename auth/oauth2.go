package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth2Manager handles OAuth2 authentication
type OAuth2Manager struct {
	configs map[string]*oauth2.Config
	logger  *logrus.Logger
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}

// OAuth2Provider represents an OAuth2 provider
type OAuth2Provider struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	ClientID    string   `json:"client_id"`
	Scopes      []string `json:"scopes"`
	AuthURL     string   `json:"auth_url"`
	TokenURL    string   `json:"token_url"`
	UserInfoURL string   `json:"user_info_url"`
	Enabled     bool     `json:"enabled"`
}

// OAuth2UserInfo represents user information from OAuth2 provider
type OAuth2UserInfo struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Picture       string                 `json:"picture"`
	Locale        string                 `json:"locale"`
	VerifiedEmail bool                   `json:"verified_email"`
	Provider      string                 `json:"provider"`
	RawData       map[string]interface{} `json:"raw_data"`
}

// OAuth2Token represents OAuth2 token information
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// NewOAuth2Manager creates a new OAuth2 manager
func NewOAuth2Manager(logger *logrus.Logger) *OAuth2Manager {
	return &OAuth2Manager{
		configs: make(map[string]*oauth2.Config),
		logger:  logger,
	}
}

// RegisterProvider registers an OAuth2 provider
func (om *OAuth2Manager) RegisterProvider(providerName string, config *OAuth2Config) error {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	om.configs[providerName] = oauth2Config

	om.logger.WithFields(logrus.Fields{
		"provider":     providerName,
		"client_id":    config.ClientID,
		"redirect_url": config.RedirectURL,
		"scopes":       config.Scopes,
	}).Info("OAuth2 provider registered successfully")

	return nil
}

// RegisterGoogleProvider registers Google OAuth2 provider
func (om *OAuth2Manager) RegisterGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) error {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	om.configs["google"] = config

	om.logger.WithFields(logrus.Fields{
		"provider":     "google",
		"client_id":    clientID,
		"redirect_url": redirectURL,
		"scopes":       scopes,
	}).Info("Google OAuth2 provider registered successfully")

	return nil
}

// GetAuthURL generates authorization URL for a provider
func (om *OAuth2Manager) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	config, exists := om.configs[providerName]
	if !exists {
		return "", fmt.Errorf("OAuth2 provider not found: %s", providerName)
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	om.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"state":    state,
	}).Debug("Generated OAuth2 authorization URL")

	return authURL, nil
}

// ExchangeCode exchanges authorization code for access token
func (om *OAuth2Manager) ExchangeCode(ctx context.Context, providerName, code string) (*OAuth2Token, error) {
	config, exists := om.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OAuth2 provider not found: %s", providerName)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		om.logger.WithError(err).WithField("provider", providerName).Error("Failed to exchange code for token")
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	oauth2Token := &OAuth2Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        token.Extra("scope").(string),
	}

	om.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"token_type":  oauth2Token.TokenType,
		"expires_at":  oauth2Token.ExpiresAt,
		"has_refresh": oauth2Token.RefreshToken != "",
	}).Info("OAuth2 token obtained successfully")

	return oauth2Token, nil
}

// RefreshToken refreshes an access token
func (om *OAuth2Manager) RefreshToken(ctx context.Context, providerName, refreshToken string) (*OAuth2Token, error) {
	config, exists := om.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OAuth2 provider not found: %s", providerName)
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		om.logger.WithError(err).WithField("provider", providerName).Error("Failed to refresh token")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	oauth2Token := &OAuth2Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresAt:    newToken.Expiry,
		Scope:        newToken.Extra("scope").(string),
	}

	om.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"token_type":  oauth2Token.TokenType,
		"expires_at":  oauth2Token.ExpiresAt,
		"has_refresh": oauth2Token.RefreshToken != "",
	}).Info("OAuth2 token refreshed successfully")

	return oauth2Token, nil
}

// GetUserInfo retrieves user information from OAuth2 provider
func (om *OAuth2Manager) GetUserInfo(ctx context.Context, providerName string, token *OAuth2Token) (*OAuth2UserInfo, error) {
	config, exists := om.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OAuth2 provider not found: %s", providerName)
	}

	client := config.Client(ctx, &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.ExpiresAt,
	})

	userInfoURL := om.getUserInfoURL(providerName)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		om.logger.WithError(err).WithField("provider", providerName).Error("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo OAuth2UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		om.logger.WithError(err).WithField("provider", providerName).Error("Failed to decode user info")
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	userInfo.Provider = providerName

	om.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"user_id":  userInfo.ID,
		"email":    userInfo.Email,
		"name":     userInfo.Name,
	}).Info("OAuth2 user info retrieved successfully")

	return &userInfo, nil
}

// ValidateToken validates an OAuth2 token
func (om *OAuth2Manager) ValidateToken(ctx context.Context, providerName string, token *OAuth2Token) (bool, error) {
	config, exists := om.configs[providerName]
	if !exists {
		return false, fmt.Errorf("OAuth2 provider not found: %s", providerName)
	}

	client := config.Client(ctx, &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.ExpiresAt,
	})

	// Try to make a request to validate the token
	userInfoURL := om.getUserInfoURL(providerName)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		om.logger.WithError(err).WithField("provider", providerName).Debug("Token validation failed")
		return false, nil
	}
	defer resp.Body.Close()

	isValid := resp.StatusCode == http.StatusOK

	om.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"is_valid": isValid,
		"status":   resp.StatusCode,
	}).Debug("OAuth2 token validation completed")

	return isValid, nil
}

// getUserInfoURL returns the user info URL for a provider
func (om *OAuth2Manager) getUserInfoURL(providerName string) string {
	switch providerName {
	case "google":
		return "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		return "https://api.github.com/user"
	case "facebook":
		return "https://graph.facebook.com/me?fields=id,name,email,picture"
	case "microsoft":
		return "https://graph.microsoft.com/v1.0/me"
	default:
		return ""
	}
}

// GetSupportedProviders returns list of supported OAuth2 providers
func (om *OAuth2Manager) GetSupportedProviders() []string {
	var providers []string
	for provider := range om.configs {
		providers = append(providers, provider)
	}
	return providers
}

// IsProviderRegistered checks if a provider is registered
func (om *OAuth2Manager) IsProviderRegistered(providerName string) bool {
	_, exists := om.configs[providerName]
	return exists
}

// CreateUserFromOAuth2 creates a user from OAuth2 user info
func (om *OAuth2Manager) CreateUserFromOAuth2(ctx context.Context, userInfo *OAuth2UserInfo) (*User, error) {
	user := &User{
		ID:            uuid.New(),
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		FirstName:     userInfo.FirstName,
		LastName:      userInfo.LastName,
		Picture:       userInfo.Picture,
		Locale:        userInfo.Locale,
		VerifiedEmail: userInfo.VerifiedEmail,
		Provider:      userInfo.Provider,
		ProviderID:    userInfo.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      userInfo.RawData,
	}

	om.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"email":       user.Email,
		"provider":    user.Provider,
		"provider_id": user.ProviderID,
	}).Info("User created from OAuth2 info")

	return user, nil
}

// User represents a user created from OAuth2
type User struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Picture       string                 `json:"picture"`
	Locale        string                 `json:"locale"`
	VerifiedEmail bool                   `json:"verified_email"`
	Provider      string                 `json:"provider"`
	ProviderID    string                 `json:"provider_id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}
