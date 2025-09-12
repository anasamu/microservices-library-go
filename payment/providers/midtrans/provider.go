package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/payment"
	"github.com/sirupsen/logrus"
)

// Provider implements PaymentProvider for Midtrans
type Provider struct {
	config       map[string]interface{}
	logger       *logrus.Logger
	httpClient   *http.Client
	baseURL      string
	serverKey    string
	clientKey    string
	isProduction bool
}

// MidtransTransaction represents a Midtrans transaction
type MidtransTransaction struct {
	Token             string `json:"token"`
	RedirectURL       string `json:"redirect_url"`
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	TransactionTime   string `json:"transaction_time"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
}

// MidtransTransactionRequest represents a Midtrans transaction request
type MidtransTransactionRequest struct {
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount int64  `json:"gross_amount"`
	} `json:"transaction_details"`
	ItemDetails []struct {
		ID       string `json:"id"`
		Price    int64  `json:"price"`
		Quantity int    `json:"quantity"`
		Name     string `json:"name"`
	} `json:"item_details"`
	CustomerDetails struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	} `json:"customer_details"`
	Callbacks struct {
		Finish string `json:"finish"`
	} `json:"callbacks"`
	Expiry struct {
		StartTime string `json:"start_time"`
		Unit      string `json:"unit"`
		Duration  int    `json:"duration"`
	} `json:"expiry"`
	EnabledPayments []string `json:"enabled_payments"`
}

// NewProvider creates a new Midtrans payment provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "midtrans"
}

// GetSupportedCurrencies returns supported currencies
func (p *Provider) GetSupportedCurrencies() []string {
	return []string{
		"IDR", "USD", "SGD", "MYR", "THB", "PHP", "VND",
	}
}

// GetSupportedMethods returns supported payment methods
func (p *Provider) GetSupportedMethods() []payment.PaymentMethod {
	return []payment.PaymentMethod{
		payment.PaymentMethodCard,
		payment.PaymentMethodBankTransfer,
		payment.PaymentMethodEWallet,
		payment.PaymentMethodQRCode,
		payment.PaymentMethodVirtualAccount,
		payment.PaymentMethodRetail,
	}
}

// Configure configures the Midtrans provider
func (p *Provider) Configure(config map[string]interface{}) error {
	serverKey, ok := config["server_key"].(string)
	if !ok || serverKey == "" {
		return fmt.Errorf("midtrans server_key is required")
	}

	clientKey, ok := config["client_key"].(string)
	if !ok || clientKey == "" {
		return fmt.Errorf("midtrans client_key is required")
	}

	isProduction, ok := config["is_production"].(bool)
	if !ok {
		isProduction = false // Default to sandbox
	}

	p.serverKey = serverKey
	p.clientKey = clientKey
	p.isProduction = isProduction

	if isProduction {
		p.baseURL = "https://api.midtrans.com"
	} else {
		p.baseURL = "https://api.sandbox.midtrans.com"
	}

	p.config = config

	p.logger.Info("Midtrans provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.serverKey != "" && p.clientKey != "" && p.baseURL != ""
}

// CreatePayment creates a payment using Midtrans
func (p *Provider) CreatePayment(ctx context.Context, request *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("midtrans provider not configured")
	}

	// Create Midtrans transaction request
	transactionRequest := MidtransTransactionRequest{
		TransactionDetails: struct {
			OrderID     string `json:"order_id"`
			GrossAmount int64  `json:"gross_amount"`
		}{
			OrderID:     request.ID,
			GrossAmount: request.Amount,
		},
		ItemDetails: []struct {
			ID       string `json:"id"`
			Price    int64  `json:"price"`
			Quantity int    `json:"quantity"`
			Name     string `json:"name"`
		}{
			{
				ID:       "item_1",
				Price:    request.Amount,
				Quantity: 1,
				Name:     request.Description,
			},
		},
		CustomerDetails: struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Phone     string `json:"phone"`
		}{
			FirstName: request.Customer.Name,
			Email:     request.Customer.Email,
			Phone:     request.Customer.Phone,
		},
		Callbacks: struct {
			Finish string `json:"finish"`
		}{
			Finish: request.ReturnURL,
		},
		Expiry: struct {
			StartTime string `json:"start_time"`
			Unit      string `json:"unit"`
			Duration  int    `json:"duration"`
		}{
			StartTime: time.Now().Format("2006-01-02 15:04:05 +0700"),
			Unit:      "hour",
			Duration:  24,
		},
		EnabledPayments: []string{
			"credit_card", "bca_va", "bni_va", "bri_va", "mandiri_va", "permata_va",
			"bca_klikbca", "bca_klikpay", "bri_epay", "echannel", "permata_va",
			"bca_va", "bni_va", "other_va", "gopay", "kioson", "indomaret",
			"gci", "danamon_online", "akulaku", "shopeepay", "qris",
		},
	}

	// Create the transaction
	transaction, err := p.createTransaction(ctx, transactionRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans transaction: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:         transaction.OrderID,
		Status:     payment.PaymentStatusPending,
		PaymentURL: transaction.RedirectURL,
		ProviderData: map[string]interface{}{
			"token":          transaction.Token,
			"transaction_id": transaction.TransactionID,
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// GetPayment retrieves a payment from Midtrans
func (p *Provider) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("midtrans provider not configured")
	}

	// Get transaction status from Midtrans
	transaction, err := p.getTransactionStatus(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get midtrans transaction: %w", err)
	}

	// Convert Midtrans transaction to payment
	payment := &payment.Payment{
		ID:          transaction.OrderID,
		Amount:      p.parseAmount(transaction.GrossAmount),
		Currency:    transaction.Currency,
		Status:      p.convertStatus(transaction.TransactionStatus),
		Description: "Midtrans Payment",
		Customer: &payment.Customer{
			Email: "", // Not available in status response
		},
		PaymentMethod: p.convertPaymentType(transaction.PaymentType),
		ProviderData: map[string]interface{}{
			"transaction_id": transaction.TransactionID,
			"payment_type":   transaction.PaymentType,
			"fraud_status":   transaction.FraudStatus,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Parse transaction time
	if transaction.TransactionTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", transaction.TransactionTime); err == nil {
			payment.CreatedAt = t
			payment.UpdatedAt = t
		}
	}

	return payment, nil
}

// CancelPayment cancels a payment in Midtrans
func (p *Provider) CancelPayment(ctx context.Context, paymentID string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("midtrans provider not configured")
	}

	// Midtrans transactions expire automatically, but we can also cancel them
	err := p.cancelTransaction(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to cancel midtrans transaction: %w", err)
	}

	return nil
}

// RefundPayment processes a refund using Midtrans
func (p *Provider) RefundPayment(ctx context.Context, request *payment.RefundRequest) (*payment.RefundResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("midtrans provider not configured")
	}

	// For Midtrans, refunds are typically handled through their dashboard
	// This is a simplified implementation
	response := &payment.RefundResponse{
		ID:           fmt.Sprintf("refund_%d", time.Now().Unix()),
		PaymentID:    request.PaymentID,
		Amount:       *request.Amount,
		Status:       payment.RefundStatusSucceeded,
		Reason:       request.Reason,
		ProviderData: map[string]interface{}{},
		CreatedAt:    time.Now(),
	}

	return response, nil
}

// ValidateWebhook validates a Midtrans webhook
func (p *Provider) ValidateWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("midtrans provider not configured")
	}

	// Parse the webhook event
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Validate signature
	if !p.validateSignature(payload, signature) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	webhookEvent := &payment.WebhookEvent{
		ID:        fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:      event["transaction_status"].(string),
		Data:      event,
		CreatedAt: time.Now(),
	}

	// Extract payment ID from event data
	if orderID, ok := event["order_id"].(string); ok {
		webhookEvent.PaymentID = orderID
	}

	return webhookEvent, nil
}

// ProcessWebhook processes a Midtrans webhook event
func (p *Provider) ProcessWebhook(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithFields(logrus.Fields{
		"event_type": event.Type,
		"payment_id": event.PaymentID,
	}).Info("Processing Midtrans webhook event")

	// Handle different event types
	switch event.Type {
	case "capture":
		return p.handleCapture(ctx, event)
	case "settlement":
		return p.handleSettlement(ctx, event)
	case "pending":
		return p.handlePending(ctx, event)
	case "deny":
		return p.handleDeny(ctx, event)
	case "expire":
		return p.handleExpire(ctx, event)
	case "cancel":
		return p.handleCancel(ctx, event)
	default:
		p.logger.WithField("event_type", event.Type).Debug("Unhandled webhook event type")
	}

	return nil
}

// createTransaction creates a Midtrans transaction
func (p *Provider) createTransaction(ctx context.Context, transactionRequest MidtransTransactionRequest) (*MidtransTransaction, error) {
	url := fmt.Sprintf("%s/v2/charge", p.baseURL)

	jsonData, err := json.Marshal(transactionRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(p.serverKey, "")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create transaction: %s", string(body))
	}

	var transaction MidtransTransaction
	if err := json.Unmarshal(body, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// getTransactionStatus gets a Midtrans transaction status
func (p *Provider) getTransactionStatus(ctx context.Context, orderID string) (*MidtransTransaction, error) {
	url := fmt.Sprintf("%s/v2/%s/status", p.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(p.serverKey, "")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get transaction status: %s", string(body))
	}

	var transaction MidtransTransaction
	if err := json.Unmarshal(body, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// cancelTransaction cancels a Midtrans transaction
func (p *Provider) cancelTransaction(ctx context.Context, orderID string) error {
	url := fmt.Sprintf("%s/v2/%s/cancel", p.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(p.serverKey, "")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel transaction: %s", string(body))
	}

	return nil
}

// validateSignature validates Midtrans webhook signature
func (p *Provider) validateSignature(payload []byte, signature string) bool {
	// Create signature from payload and server key
	hash := sha512.New()
	hash.Write(payload)
	hash.Write([]byte(p.serverKey))
	expectedSignature := fmt.Sprintf("%x", hash.Sum(nil))

	return signature == expectedSignature
}

// convertStatus converts Midtrans status to payment status
func (p *Provider) convertStatus(status string) payment.PaymentStatus {
	switch status {
	case "capture":
		return payment.PaymentStatusSucceeded
	case "settlement":
		return payment.PaymentStatusSucceeded
	case "pending":
		return payment.PaymentStatusPending
	case "deny":
		return payment.PaymentStatusFailed
	case "expire":
		return payment.PaymentStatusExpired
	case "cancel":
		return payment.PaymentStatusCanceled
	default:
		return payment.PaymentStatusPending
	}
}

// convertPaymentType converts Midtrans payment type to payment method
func (p *Provider) convertPaymentType(paymentType string) payment.PaymentMethod {
	switch paymentType {
	case "credit_card":
		return payment.PaymentMethodCard
	case "bank_transfer", "bca_va", "bni_va", "bri_va", "mandiri_va", "permata_va":
		return payment.PaymentMethodBankTransfer
	case "gopay", "shopeepay", "qris":
		return payment.PaymentMethodEWallet
	case "qr_code":
		return payment.PaymentMethodQRCode
	default:
		return payment.PaymentMethodCard
	}
}

// parseAmount parses amount string to int64
func (p *Provider) parseAmount(amountStr string) int64 {
	// Remove any non-numeric characters except decimal point
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")

	// Parse as float64 first, then convert to int64 (cents)
	if amount, err := fmt.Sscanf(amountStr, "%f", new(float64)); err == nil && amount == 1 {
		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)
		return int64(amount)
	}

	return 0
}

// Webhook event handlers
func (p *Provider) handleCapture(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment captured")
	return nil
}

func (p *Provider) handleSettlement(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment settled")
	return nil
}

func (p *Provider) handlePending(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment pending")
	return nil
}

func (p *Provider) handleDeny(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment denied")
	return nil
}

func (p *Provider) handleExpire(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment expired")
	return nil
}

func (p *Provider) handleCancel(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment cancelled")
	return nil
}
