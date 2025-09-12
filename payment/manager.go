package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PaymentManager manages multiple payment gateways
type PaymentManager struct {
	providers map[string]PaymentProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds payment manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	WebhookSecret   string            `json:"webhook_secret"`
	Metadata        map[string]string `json:"metadata"`
}

// PaymentProvider interface for payment gateways
type PaymentProvider interface {
	// Provider information
	GetName() string
	GetSupportedCurrencies() []string
	GetSupportedMethods() []PaymentMethod

	// Payment operations
	CreatePayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	CancelPayment(ctx context.Context, paymentID string) error
	RefundPayment(ctx context.Context, request *RefundRequest) (*RefundResponse, error)

	// Webhook handling
	ValidateWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
	ProcessWebhook(ctx context.Context, event *WebhookEvent) error

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	ID            string                 `json:"id"`
	Amount        int64                  `json:"amount"` // Amount in cents
	Currency      string                 `json:"currency"`
	Description   string                 `json:"description"`
	Customer      *Customer              `json:"customer"`
	PaymentMethod PaymentMethod          `json:"payment_method"`
	Metadata      map[string]interface{} `json:"metadata"`
	ReturnURL     string                 `json:"return_url,omitempty"`
	CancelURL     string                 `json:"cancel_url,omitempty"`
	WebhookURL    string                 `json:"webhook_url,omitempty"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID           string                 `json:"id"`
	Status       PaymentStatus          `json:"status"`
	PaymentURL   string                 `json:"payment_url,omitempty"`
	ClientSecret string                 `json:"client_secret,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

// Payment represents a payment
type Payment struct {
	ID            string                 `json:"id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        PaymentStatus          `json:"status"`
	Description   string                 `json:"description"`
	Customer      *Customer              `json:"customer"`
	PaymentMethod PaymentMethod          `json:"payment_method"`
	Metadata      map[string]interface{} `json:"metadata"`
	ProviderData  map[string]interface{} `json:"provider_data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	PaidAt        *time.Time             `json:"paid_at,omitempty"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID string                 `json:"payment_id"`
	Amount    *int64                 `json:"amount,omitempty"` // If nil, refund full amount
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID           string                 `json:"id"`
	PaymentID    string                 `json:"payment_id"`
	Amount       int64                  `json:"amount"`
	Status       RefundStatus           `json:"status"`
	Reason       string                 `json:"reason,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Customer represents a customer
type Customer struct {
	ID       string                 `json:"id,omitempty"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name,omitempty"`
	Phone    string                 `json:"phone,omitempty"`
	Address  *Address               `json:"address,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Address represents a customer address
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	PaymentID string                 `json:"payment_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// PaymentMethod represents a payment method
type PaymentMethod string

const (
	PaymentMethodCard           PaymentMethod = "card"
	PaymentMethodBankTransfer   PaymentMethod = "bank_transfer"
	PaymentMethodEWallet        PaymentMethod = "ewallet"
	PaymentMethodQRCode         PaymentMethod = "qr_code"
	PaymentMethodVirtualAccount PaymentMethod = "virtual_account"
	PaymentMethodRetail         PaymentMethod = "retail"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSucceeded  PaymentStatus = "succeeded"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCanceled   PaymentStatus = "canceled"
	PaymentStatusExpired    PaymentStatus = "expired"
)

// RefundStatus represents refund status
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCanceled  RefundStatus = "canceled"
)

// DefaultManagerConfig returns default payment manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "stripe",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		WebhookSecret:   "",
		Metadata:        make(map[string]string),
	}
}

// NewPaymentManager creates a new payment manager
func NewPaymentManager(config *ManagerConfig, logger *logrus.Logger) *PaymentManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &PaymentManager{
		providers: make(map[string]PaymentProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a payment provider
func (pm *PaymentManager) RegisterProvider(provider PaymentProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	pm.providers[name] = provider
	pm.logger.WithField("provider", name).Info("Payment provider registered")

	return nil
}

// GetProvider returns a payment provider by name
func (pm *PaymentManager) GetProvider(name string) (PaymentProvider, error) {
	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("payment provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default payment provider
func (pm *PaymentManager) GetDefaultProvider() (PaymentProvider, error) {
	return pm.GetProvider(pm.config.DefaultProvider)
}

// CreatePayment creates a payment using the specified provider
func (pm *PaymentManager) CreatePayment(ctx context.Context, providerName string, request *PaymentRequest) (*PaymentResponse, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Set default values
	if request.ID == "" {
		request.ID = uuid.New().String()
	}

	// Validate request
	if err := pm.validatePaymentRequest(request); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// Check if provider supports the payment method
	if !pm.supportsPaymentMethod(provider, request.PaymentMethod) {
		return nil, fmt.Errorf("provider %s does not support payment method %s", providerName, request.PaymentMethod)
	}

	// Check if provider supports the currency
	if !pm.supportsCurrency(provider, request.Currency) {
		return nil, fmt.Errorf("provider %s does not support currency %s", providerName, request.Currency)
	}

	// Create payment with retry logic
	var response *PaymentResponse
	for attempt := 1; attempt <= pm.config.RetryAttempts; attempt++ {
		response, err = provider.CreatePayment(ctx, request)
		if err == nil {
			break
		}

		pm.logger.WithError(err).WithFields(logrus.Fields{
			"provider":   providerName,
			"attempt":    attempt,
			"payment_id": request.ID,
		}).Warn("Payment creation failed, retrying")

		if attempt < pm.config.RetryAttempts {
			time.Sleep(pm.config.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create payment after %d attempts: %w", pm.config.RetryAttempts, err)
	}

	pm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"payment_id": response.ID,
		"amount":     request.Amount,
		"currency":   request.Currency,
	}).Info("Payment created successfully")

	return response, nil
}

// GetPayment retrieves a payment from the specified provider
func (pm *PaymentManager) GetPayment(ctx context.Context, providerName, paymentID string) (*Payment, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	payment, err := provider.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// CancelPayment cancels a payment using the specified provider
func (pm *PaymentManager) CancelPayment(ctx context.Context, providerName, paymentID string) error {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.CancelPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	pm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"payment_id": paymentID,
	}).Info("Payment canceled successfully")

	return nil
}

// RefundPayment processes a refund using the specified provider
func (pm *PaymentManager) RefundPayment(ctx context.Context, providerName string, request *RefundRequest) (*RefundResponse, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.RefundPayment(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	pm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"payment_id": request.PaymentID,
		"refund_id":  response.ID,
		"amount":     response.Amount,
	}).Info("Refund processed successfully")

	return response, nil
}

// ProcessWebhook processes a webhook event
func (pm *PaymentManager) ProcessWebhook(ctx context.Context, providerName string, payload []byte, signature string) error {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return err
	}

	event, err := provider.ValidateWebhook(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to validate webhook: %w", err)
	}

	err = provider.ProcessWebhook(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	pm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"event_type": event.Type,
		"payment_id": event.PaymentID,
	}).Info("Webhook processed successfully")

	return nil
}

// GetSupportedProviders returns a list of registered providers
func (pm *PaymentManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(pm.providers))
	for name := range pm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (pm *PaymentManager) GetProviderCapabilities(providerName string) ([]PaymentMethod, []string, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	return provider.GetSupportedMethods(), provider.GetSupportedCurrencies(), nil
}

// validatePaymentRequest validates a payment request
func (pm *PaymentManager) validatePaymentRequest(request *PaymentRequest) error {
	if request.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if request.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if request.Description == "" {
		return fmt.Errorf("description is required")
	}

	if request.Customer == nil {
		return fmt.Errorf("customer is required")
	}

	if request.Customer.Email == "" {
		return fmt.Errorf("customer email is required")
	}

	return nil
}

// supportsPaymentMethod checks if provider supports the payment method
func (pm *PaymentManager) supportsPaymentMethod(provider PaymentProvider, method PaymentMethod) bool {
	supportedMethods := provider.GetSupportedMethods()
	for _, supported := range supportedMethods {
		if supported == method {
			return true
		}
	}
	return false
}

// supportsCurrency checks if provider supports the currency
func (pm *PaymentManager) supportsCurrency(provider PaymentProvider, currency string) bool {
	supportedCurrencies := provider.GetSupportedCurrencies()
	for _, supported := range supportedCurrencies {
		if supported == currency {
			return true
		}
	}
	return false
}
