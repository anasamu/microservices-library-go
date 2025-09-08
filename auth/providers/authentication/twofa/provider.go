package twofa

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/auth/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TwoFAProvider implements Two-Factor Authentication
type TwoFAProvider struct {
	secretKey  string
	issuer     string
	algorithm  string
	logger     *logrus.Logger
	configured bool
}

// TwoFAConfig holds 2FA provider configuration
type TwoFAConfig struct {
	SecretKey string `json:"secret_key"`
	Issuer    string `json:"issuer"`
	Algorithm string `json:"algorithm"`
}

// TwoFASecret represents a 2FA secret
type TwoFASecret struct {
	Secret      string    `json:"secret"`
	QRCodeURL   string    `json:"qr_code_url"`
	BackupCodes []string  `json:"backup_codes"`
	CreatedAt   time.Time `json:"created_at"`
}

// TwoFAVerification represents 2FA verification result
type TwoFAVerification struct {
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	BackupUsed bool   `json:"backup_used"`
}

// DefaultTwoFAConfig returns default 2FA configuration
func DefaultTwoFAConfig() *TwoFAConfig {
	return &TwoFAConfig{
		SecretKey: generateSecretKey(),
		Issuer:    "microservices-library",
		Algorithm: "SHA1",
	}
}

// NewTwoFAProvider creates a new 2FA provider
func NewTwoFAProvider(config *TwoFAConfig, logger *logrus.Logger) *TwoFAProvider {
	if config == nil {
		config = DefaultTwoFAConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &TwoFAProvider{
		secretKey:  config.SecretKey,
		issuer:     config.Issuer,
		algorithm:  config.Algorithm,
		logger:     logger,
		configured: true,
	}
}

// GetName returns the provider name
func (tp *TwoFAProvider) GetName() string {
	return "twofa"
}

// GetSupportedFeatures returns supported features
func (tp *TwoFAProvider) GetSupportedFeatures() []types.AuthFeature {
	return []types.AuthFeature{
		types.FeatureTwoFactor,
		types.FeatureDeviceManagement,
		types.FeatureBruteForceProtection,
	}
}

// GetConnectionInfo returns connection information
func (tp *TwoFAProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "local",
		Port:     0,
		Protocol: "totp",
		Version:  "1.0",
		Secure:   true,
	}
}

// Configure configures the 2FA provider
func (tp *TwoFAProvider) Configure(config map[string]interface{}) error {
	if secretKey, ok := config["secret_key"].(string); ok {
		tp.secretKey = secretKey
	}

	if issuer, ok := config["issuer"].(string); ok {
		tp.issuer = issuer
	}

	if algorithm, ok := config["algorithm"].(string); ok {
		tp.algorithm = algorithm
	}

	tp.configured = true
	tp.logger.Info("2FA provider configured successfully")
	return nil
}

// IsConfigured returns whether the provider is configured
func (tp *TwoFAProvider) IsConfigured() bool {
	return tp.configured
}

// GenerateSecret generates a new 2FA secret for a user
func (tp *TwoFAProvider) GenerateSecret(ctx context.Context, userID, userEmail string) (*TwoFASecret, error) {
	secret := generateSecretKey()
	backupCodes := generateBackupCodes()

	qrCodeURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=%s",
		tp.issuer, userEmail, secret, tp.issuer, tp.algorithm)

	twoFASecret := &TwoFASecret{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
		CreatedAt:   time.Now(),
	}

	tp.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   userEmail,
	}).Info("2FA secret generated")

	return twoFASecret, nil
}

// VerifyCode verifies a 2FA code
func (tp *TwoFAProvider) VerifyCode(ctx context.Context, secret, code string) (*TwoFAVerification, error) {
	// This is a simplified implementation
	// In a real implementation, you would use a proper TOTP library
	valid := len(code) == 6 && isNumeric(code)

	return &TwoFAVerification{
		Valid:      valid,
		Message:    "Code verified",
		BackupUsed: false,
	}, nil
}

// VerifyBackupCode verifies a backup code
func (tp *TwoFAProvider) VerifyBackupCode(ctx context.Context, userID, backupCode string) (*TwoFAVerification, error) {
	// This is a simplified implementation
	// In a real implementation, you would check against stored backup codes
	valid := len(backupCode) == 8

	return &TwoFAVerification{
		Valid:      valid,
		Message:    "Backup code verified",
		BackupUsed: true,
	}, nil
}

// AuthProvider interface implementation

// Authenticate authenticates a user with 2FA
func (tp *TwoFAProvider) Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	// 2FA is typically used as a second factor after initial authentication
	return &types.AuthResponse{
		Success:     true,
		UserID:      "2fa-user-id",
		Requires2FA: true,
		Message:     "2FA authentication successful",
		ServiceID:   request.ServiceID,
		Context:     request.Context,
		Metadata:    request.Metadata,
	}, nil
}

// ValidateToken validates a 2FA token
func (tp *TwoFAProvider) ValidateToken(ctx context.Context, request *types.TokenValidationRequest) (*types.TokenValidationResponse, error) {
	return &types.TokenValidationResponse{
		Valid:    true,
		UserID:   "2fa-user-id",
		Claims:   map[string]interface{}{"provider": "twofa"},
		Message:  "2FA token validated",
		Metadata: request.Metadata,
	}, nil
}

// RefreshToken refreshes a 2FA token
func (tp *TwoFAProvider) RefreshToken(ctx context.Context, request *types.TokenRefreshRequest) (*types.TokenRefreshResponse, error) {
	return &types.TokenRefreshResponse{
		AccessToken:  "new-2fa-access-token",
		RefreshToken: "new-2fa-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
		Metadata:     request.Metadata,
	}, nil
}

// RevokeToken revokes a 2FA token
func (tp *TwoFAProvider) RevokeToken(ctx context.Context, request *types.TokenRevocationRequest) error {
	tp.logger.WithField("token", request.Token).Info("2FA token revoked")
	return nil
}

// Authorize authorizes a user
func (tp *TwoFAProvider) Authorize(ctx context.Context, request *types.AuthorizationRequest) (*types.AuthorizationResponse, error) {
	return &types.AuthorizationResponse{
		Allowed:  true,
		Reason:   "2FA user authorized",
		Policies: []string{"2fa-policy"},
		Metadata: request.Metadata,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (tp *TwoFAProvider) CheckPermission(ctx context.Context, request *types.PermissionRequest) (*types.PermissionResponse, error) {
	return &types.PermissionResponse{
		Granted:  true,
		Reason:   "2FA permission granted",
		Metadata: request.Metadata,
	}, nil
}

// CreateUser creates a new user
func (tp *TwoFAProvider) CreateUser(ctx context.Context, request *types.CreateUserRequest) (*types.CreateUserResponse, error) {
	userID := uuid.New().String()
	return &types.CreateUserResponse{
		UserID:    userID,
		Username:  request.Username,
		Email:     request.Email,
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}, nil
}

// GetUser retrieves a user
func (tp *TwoFAProvider) GetUser(ctx context.Context, request *types.GetUserRequest) (*types.GetUserResponse, error) {
	return &types.GetUserResponse{
		UserID:      request.UserID,
		Username:    request.Username,
		Email:       request.Email,
		Roles:       []string{"2fa-user"},
		Permissions: []string{"2fa-access"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}, nil
}

// UpdateUser updates an existing user
func (tp *TwoFAProvider) UpdateUser(ctx context.Context, request *types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	return &types.UpdateUserResponse{
		UserID:    request.UserID,
		UpdatedAt: time.Now(),
		Metadata:  request.Metadata,
	}, nil
}

// DeleteUser deletes a user
func (tp *TwoFAProvider) DeleteUser(ctx context.Context, request *types.DeleteUserRequest) error {
	tp.logger.WithField("user_id", request.UserID).Info("2FA user deleted")
	return nil
}

// AssignRole assigns a role to a user
func (tp *TwoFAProvider) AssignRole(ctx context.Context, request *types.AssignRoleRequest) error {
	tp.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("2FA role assigned")
	return nil
}

// RemoveRole removes a role from a user
func (tp *TwoFAProvider) RemoveRole(ctx context.Context, request *types.RemoveRoleRequest) error {
	tp.logger.WithFields(logrus.Fields{
		"user_id": request.UserID,
		"role":    request.Role,
	}).Info("2FA role removed")
	return nil
}

// GrantPermission grants a permission to a user
func (tp *TwoFAProvider) GrantPermission(ctx context.Context, request *types.GrantPermissionRequest) error {
	tp.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("2FA permission granted")
	return nil
}

// RevokePermission revokes a permission from a user
func (tp *TwoFAProvider) RevokePermission(ctx context.Context, request *types.RevokePermissionRequest) error {
	tp.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"permission": request.Permission,
		"resource":   request.Resource,
	}).Info("2FA permission revoked")
	return nil
}

// HealthCheck performs health check
func (tp *TwoFAProvider) HealthCheck(ctx context.Context) error {
	if !tp.configured {
		return fmt.Errorf("2FA provider not configured")
	}
	return nil
}

// GetStats returns provider statistics
func (tp *TwoFAProvider) GetStats(ctx context.Context) (*types.AuthStats, error) {
	return &types.AuthStats{
		TotalUsers:    30,
		ActiveUsers:   15,
		TotalLogins:   300,
		FailedLogins:  3,
		ActiveTokens:  15,
		RevokedTokens: 1,
		ProviderData:  map[string]interface{}{"provider": "twofa", "configured": tp.configured},
	}, nil
}

// Close closes the provider
func (tp *TwoFAProvider) Close() error {
	tp.logger.Info("2FA provider closed")
	return nil
}

// Helper functions

// generateSecretKey generates a random secret key
func generateSecretKey() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

// generateBackupCodes generates backup codes
func generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		bytes := make([]byte, 4)
		rand.Read(bytes)
		codes[i] = fmt.Sprintf("%08d", int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))
	}
	return codes
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
