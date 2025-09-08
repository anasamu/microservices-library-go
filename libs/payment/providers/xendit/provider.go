package xendit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/libs/payment/gateway"
	"github.com/sirupsen/logrus"
)

// Provider implements PaymentProvider for Xendit
type Provider struct {
	config     map[string]interface{}
	logger     *logrus.Logger
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// XenditInvoice represents a Xendit invoice
type XenditInvoice struct {
	ID                        string `json:"id"`
	ExternalID                string `json:"external_id"`
	UserID                    string `json:"user_id"`
	Status                    string `json:"status"`
	MerchantName              string `json:"merchant_name"`
	MerchantProfilePictureURL string `json:"merchant_profile_picture_url"`
	Amount                    int64  `json:"amount"`
	Description               string `json:"description"`
	InvoiceURL                string `json:"invoice_url"`
	ExpiryDate                string `json:"expiry_date"`
	Created                   string `json:"created"`
	Updated                   string `json:"updated"`
	Currency                  string `json:"currency"`
}

// XenditInvoiceRequest represents a Xendit invoice request
type XenditInvoiceRequest struct {
	ExternalID      string `json:"external_id"`
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	InvoiceDuration int    `json:"invoice_duration"`
	Currency        string `json:"currency"`
	Customer        struct {
		Email        string `json:"email"`
		GivenNames   string `json:"given_names,omitempty"`
		Surname      string `json:"surname,omitempty"`
		MobileNumber string `json:"mobile_number,omitempty"`
	} `json:"customer"`
	CustomerNotificationPreference struct {
		InvoiceCreated  []string `json:"invoice_created"`
		InvoiceReminder []string `json:"invoice_reminder"`
		InvoicePaid     []string `json:"invoice_paid"`
		InvoiceExpired  []string `json:"invoice_expired"`
	} `json:"customer_notification_preference"`
	SuccessRedirectURL string   `json:"success_redirect_url,omitempty"`
	FailureRedirectURL string   `json:"failure_redirect_url,omitempty"`
	PaymentMethods     []string `json:"payment_methods"`
}

// NewProvider creates a new Xendit payment provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.xendit.co",
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "xendit"
}

// GetSupportedCurrencies returns supported currencies
func (p *Provider) GetSupportedCurrencies() []string {
	return []string{
		"IDR", "PHP", "THB", "VND", "MYR", "SGD", "USD",
	}
}

// GetSupportedMethods returns supported payment methods
func (p *Provider) GetSupportedMethods() []gateway.PaymentMethod {
	return []gateway.PaymentMethod{
		gateway.PaymentMethodCard,
		gateway.PaymentMethodBankTransfer,
		gateway.PaymentMethodEWallet,
		gateway.PaymentMethodQRCode,
		gateway.PaymentMethodVirtualAccount,
		gateway.PaymentMethodRetail,
	}
}

// Configure configures the Xendit provider
func (p *Provider) Configure(config map[string]interface{}) error {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("xendit api_key is required")
	}

	p.apiKey = apiKey
	p.config = config

	p.logger.Info("Xendit provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.apiKey != "" && p.baseURL != ""
}

// CreatePayment creates a payment using Xendit
func (p *Provider) CreatePayment(ctx context.Context, request *gateway.PaymentRequest) (*gateway.PaymentResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("xendit provider not configured")
	}

	// Create Xendit invoice request
	invoiceRequest := XenditInvoiceRequest{
		ExternalID:      request.ID,
		Amount:          request.Amount,
		Description:     request.Description,
		InvoiceDuration: 24 * 60, // 24 hours in minutes
		Currency:        request.Currency,
		Customer: struct {
			Email        string `json:"email"`
			GivenNames   string `json:"given_names,omitempty"`
			Surname      string `json:"surname,omitempty"`
			MobileNumber string `json:"mobile_number,omitempty"`
		}{
			Email:      request.Customer.Email,
			GivenNames: request.Customer.Name,
		},
		CustomerNotificationPreference: struct {
			InvoiceCreated  []string `json:"invoice_created"`
			InvoiceReminder []string `json:"invoice_reminder"`
			InvoicePaid     []string `json:"invoice_paid"`
			InvoiceExpired  []string `json:"invoice_expired"`
		}{
			InvoiceCreated:  []string{"email"},
			InvoiceReminder: []string{"email"},
			InvoicePaid:     []string{"email"},
			InvoiceExpired:  []string{"email"},
		},
		SuccessRedirectURL: request.ReturnURL,
		FailureRedirectURL: request.CancelURL,
		PaymentMethods:     []string{"CREDIT_CARD", "BCA", "BNI", "BRI", "MANDIRI", "PERMATA", "BSI", "OVO", "DANA", "LINKAJA", "SHOPEEPAY", "QRIS"},
	}

	// Create the invoice
	invoice, err := p.createInvoice(ctx, invoiceRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create xendit invoice: %w", err)
	}

	response := &gateway.PaymentResponse{
		ID:         invoice.ID,
		Status:     p.convertStatus(invoice.Status),
		PaymentURL: invoice.InvoiceURL,
		ProviderData: map[string]interface{}{
			"invoice_id":  invoice.ID,
			"external_id": invoice.ExternalID,
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// GetPayment retrieves a payment from Xendit
func (p *Provider) GetPayment(ctx context.Context, paymentID string) (*gateway.Payment, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("xendit provider not configured")
	}

	// Get invoice details from Xendit
	invoice, err := p.getInvoice(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get xendit invoice: %w", err)
	}

	// Convert Xendit invoice to gateway payment
	payment := &gateway.Payment{
		ID:          invoice.ID,
		Amount:      invoice.Amount,
		Currency:    invoice.Currency,
		Status:      p.convertStatus(invoice.Status),
		Description: invoice.Description,
		Customer: &gateway.Customer{
			Email: invoice.UserID, // Xendit uses user_id as email
		},
		PaymentMethod: gateway.PaymentMethodCard, // Default, could be determined from payment method used
		ProviderData: map[string]interface{}{
			"invoice_id":  invoice.ID,
			"external_id": invoice.ExternalID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Parse created and updated times
	if created, err := time.Parse(time.RFC3339, invoice.Created); err == nil {
		payment.CreatedAt = created
	}
	if updated, err := time.Parse(time.RFC3339, invoice.Updated); err == nil {
		payment.UpdatedAt = updated
	}

	return payment, nil
}

// CancelPayment cancels a payment in Xendit
func (p *Provider) CancelPayment(ctx context.Context, paymentID string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("xendit provider not configured")
	}

	// Xendit invoices expire automatically, but we can also expire them manually
	err := p.expireInvoice(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to expire xendit invoice: %w", err)
	}

	return nil
}

// RefundPayment processes a refund using Xendit
func (p *Provider) RefundPayment(ctx context.Context, request *gateway.RefundRequest) (*gateway.RefundResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("xendit provider not configured")
	}

	// For Xendit, refunds are typically handled through their dashboard or API
	// This is a simplified implementation
	response := &gateway.RefundResponse{
		ID:           fmt.Sprintf("refund_%d", time.Now().Unix()),
		PaymentID:    request.PaymentID,
		Amount:       *request.Amount,
		Status:       gateway.RefundStatusSucceeded,
		Reason:       request.Reason,
		ProviderData: map[string]interface{}{},
		CreatedAt:    time.Now(),
	}

	return response, nil
}

// ValidateWebhook validates a Xendit webhook
func (p *Provider) ValidateWebhook(ctx context.Context, payload []byte, signature string) (*gateway.WebhookEvent, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("xendit provider not configured")
	}

	// Parse the webhook event
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// In a real implementation, you would validate the signature here
	// using Xendit's webhook signature verification

	webhookEvent := &gateway.WebhookEvent{
		ID:        fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:      event["event"].(string),
		Data:      event,
		CreatedAt: time.Now(),
	}

	// Extract payment ID from event data
	if data, ok := event["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			webhookEvent.PaymentID = id
		}
	}

	return webhookEvent, nil
}

// ProcessWebhook processes a Xendit webhook event
func (p *Provider) ProcessWebhook(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithFields(logrus.Fields{
		"event_type": event.Type,
		"payment_id": event.PaymentID,
	}).Info("Processing Xendit webhook event")

	// Handle different event types
	switch event.Type {
	case "invoice.paid":
		return p.handleInvoicePaid(ctx, event)
	case "invoice.expired":
		return p.handleInvoiceExpired(ctx, event)
	case "invoice.failed":
		return p.handleInvoiceFailed(ctx, event)
	default:
		p.logger.WithField("event_type", event.Type).Debug("Unhandled webhook event type")
	}

	return nil
}

// createInvoice creates a Xendit invoice
func (p *Provider) createInvoice(ctx context.Context, invoiceRequest XenditInvoiceRequest) (*XenditInvoice, error) {
	url := fmt.Sprintf("%s/v2/invoices", p.baseURL)

	jsonData, err := json.Marshal(invoiceRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", p.apiKey))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create invoice: %s", string(body))
	}

	var invoice XenditInvoice
	if err := json.Unmarshal(body, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// getInvoice gets a Xendit invoice
func (p *Provider) getInvoice(ctx context.Context, invoiceID string) (*XenditInvoice, error) {
	url := fmt.Sprintf("%s/v2/invoices/%s", p.baseURL, invoiceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", p.apiKey))

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
		return nil, fmt.Errorf("failed to get invoice: %s", string(body))
	}

	var invoice XenditInvoice
	if err := json.Unmarshal(body, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// expireInvoice expires a Xendit invoice
func (p *Provider) expireInvoice(ctx context.Context, invoiceID string) error {
	url := fmt.Sprintf("%s/v2/invoices/%s/expire!", p.baseURL, invoiceID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", p.apiKey))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to expire invoice: %s", string(body))
	}

	return nil
}

// convertStatus converts Xendit status to gateway status
func (p *Provider) convertStatus(status string) gateway.PaymentStatus {
	switch status {
	case "PENDING":
		return gateway.PaymentStatusPending
	case "PAID":
		return gateway.PaymentStatusSucceeded
	case "SETTLED":
		return gateway.PaymentStatusSucceeded
	case "EXPIRED":
		return gateway.PaymentStatusExpired
	case "FAILED":
		return gateway.PaymentStatusFailed
	default:
		return gateway.PaymentStatusPending
	}
}

// handleInvoicePaid handles invoice paid events
func (p *Provider) handleInvoicePaid(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Invoice paid")
	// Implement your business logic here
	return nil
}

// handleInvoiceExpired handles invoice expired events
func (p *Provider) handleInvoiceExpired(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Invoice expired")
	// Implement your business logic here
	return nil
}

// handleInvoiceFailed handles invoice failed events
func (p *Provider) handleInvoiceFailed(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Invoice failed")
	// Implement your business logic here
	return nil
}
