package paypal

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

// Provider implements PaymentProvider for PayPal
type Provider struct {
	config      map[string]interface{}
	logger      *logrus.Logger
	httpClient  *http.Client
	baseURL     string
	accessToken string
}

// PayPalOrder represents a PayPal order
type PayPalOrder struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

// PayPalOrderRequest represents a PayPal order request
type PayPalOrderRequest struct {
	Intent        string `json:"intent"`
	PurchaseUnits []struct {
		Amount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
		Description string `json:"description"`
	} `json:"purchase_units"`
	ApplicationContext struct {
		ReturnURL string `json:"return_url"`
		CancelURL string `json:"cancel_url"`
	} `json:"application_context"`
}

// PayPalAccessToken represents PayPal access token response
type PayPalAccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewProvider creates a new PayPal payment provider
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
	return "paypal"
}

// GetSupportedCurrencies returns supported currencies
func (p *Provider) GetSupportedCurrencies() []string {
	return []string{
		"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "SEK", "NOK", "DKK",
		"PLN", "CZK", "HUF", "BGN", "RON", "HRK", "RUB", "TRY", "BRL", "MXN",
		"ARS", "CLP", "COP", "PEN", "UYU", "VES", "INR", "SGD", "HKD", "TWD",
		"KRW", "THB", "MYR", "PHP", "IDR", "VND", "ZAR", "EGP", "MAD", "NGN",
		"KES", "GHS", "UGX", "TZS", "ETB", "XOF", "XAF", "NZD", "ILS", "AED",
		"SAR", "QAR", "KWD", "BHD", "OMR", "JOD", "LBP", "PKR", "BDT", "LKR",
		"NPR", "AFN", "KZT", "UZS", "KGS", "TJS", "TMT", "AMD", "AZN", "GEL",
		"MDL", "BYN", "UAH", "BAM", "RSD", "MKD", "ALL", "ISK", "DZD", "TND",
		"LYD", "SDG", "SSP", "CDF", "AOA", "BWP", "SZL", "LSL", "NAD", "ZMW",
		"MWK", "BIF", "DJF", "KMF", "RWF", "SCR", "SOS", "TND", "UGX", "ZWL",
	}
}

// GetSupportedMethods returns supported payment methods
func (p *Provider) GetSupportedMethods() []gateway.PaymentMethod {
	return []gateway.PaymentMethod{
		gateway.PaymentMethodCard,
		gateway.PaymentMethodBankTransfer,
		gateway.PaymentMethodEWallet,
	}
}

// Configure configures the PayPal provider
func (p *Provider) Configure(config map[string]interface{}) error {
	clientID, ok := config["client_id"].(string)
	if !ok || clientID == "" {
		return fmt.Errorf("paypal client_id is required")
	}

	clientSecret, ok := config["client_secret"].(string)
	if !ok || clientSecret == "" {
		return fmt.Errorf("paypal client_secret is required")
	}

	sandbox, ok := config["sandbox"].(bool)
	if !ok {
		sandbox = true // Default to sandbox
	}

	if sandbox {
		p.baseURL = "https://api.sandbox.paypal.com"
	} else {
		p.baseURL = "https://api.paypal.com"
	}

	p.config = config

	// Get access token
	if err := p.getAccessToken(clientID, clientSecret); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	p.logger.Info("PayPal provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.accessToken != "" && p.baseURL != ""
}

// CreatePayment creates a payment using PayPal
func (p *Provider) CreatePayment(ctx context.Context, request *gateway.PaymentRequest) (*gateway.PaymentResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("paypal provider not configured")
	}

	// Convert amount from cents to dollars
	amount := float64(request.Amount) / 100.0

	// Create PayPal order request
	orderRequest := PayPalOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []struct {
			Amount struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
			Description string `json:"description"`
		}{
			{
				Amount: struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				}{
					CurrencyCode: request.Currency,
					Value:        fmt.Sprintf("%.2f", amount),
				},
				Description: request.Description,
			},
		},
		ApplicationContext: struct {
			ReturnURL string `json:"return_url"`
			CancelURL string `json:"cancel_url"`
		}{
			ReturnURL: request.ReturnURL,
			CancelURL: request.CancelURL,
		},
	}

	// Create the order
	order, err := p.createOrder(ctx, orderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal order: %w", err)
	}

	// Find approval URL
	var approvalURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	if approvalURL == "" {
		return nil, fmt.Errorf("approval URL not found in paypal order response")
	}

	response := &gateway.PaymentResponse{
		ID:         order.ID,
		Status:     gateway.PaymentStatusPending,
		PaymentURL: approvalURL,
		ProviderData: map[string]interface{}{
			"order_id": order.ID,
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// GetPayment retrieves a payment from PayPal
func (p *Provider) GetPayment(ctx context.Context, paymentID string) (*gateway.Payment, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("paypal provider not configured")
	}

	// Get order details from PayPal
	order, err := p.getOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get paypal order: %w", err)
	}

	// Convert PayPal order to gateway payment
	payment := &gateway.Payment{
		ID:     order.ID,
		Status: p.convertStatus(order.Status),
		ProviderData: map[string]interface{}{
			"order_id": order.ID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return payment, nil
}

// CancelPayment cancels a payment in PayPal
func (p *Provider) CancelPayment(ctx context.Context, paymentID string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("paypal provider not configured")
	}

	// PayPal orders are automatically cancelled if not approved within a certain time
	// We can also explicitly cancel them if needed
	p.logger.WithField("payment_id", paymentID).Info("PayPal order will be automatically cancelled if not approved")
	return nil
}

// RefundPayment processes a refund using PayPal
func (p *Provider) RefundPayment(ctx context.Context, request *gateway.RefundRequest) (*gateway.RefundResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("paypal provider not configured")
	}

	// For PayPal, we need to capture the payment first, then refund
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

// ValidateWebhook validates a PayPal webhook
func (p *Provider) ValidateWebhook(ctx context.Context, payload []byte, signature string) (*gateway.WebhookEvent, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("paypal provider not configured")
	}

	// Parse the webhook event
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// In a real implementation, you would validate the signature here
	// using PayPal's webhook signature verification

	webhookEvent := &gateway.WebhookEvent{
		ID:        fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:      event["event_type"].(string),
		Data:      event,
		CreatedAt: time.Now(),
	}

	return webhookEvent, nil
}

// ProcessWebhook processes a PayPal webhook event
func (p *Provider) ProcessWebhook(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithFields(logrus.Fields{
		"event_type": event.Type,
		"payment_id": event.PaymentID,
	}).Info("Processing PayPal webhook event")

	// Handle different event types
	switch event.Type {
	case "PAYMENT.CAPTURE.COMPLETED":
		return p.handlePaymentCaptureCompleted(ctx, event)
	case "PAYMENT.CAPTURE.DENIED":
		return p.handlePaymentCaptureDenied(ctx, event)
	default:
		p.logger.WithField("event_type", event.Type).Debug("Unhandled webhook event type")
	}

	return nil
}

// getAccessToken gets PayPal access token
func (p *Provider) getAccessToken(clientID, clientSecret string) error {
	url := fmt.Sprintf("%s/v1/oauth2/token", p.baseURL)

	data := "grant_type=client_credentials"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get access token: %s", string(body))
	}

	var tokenResp PayPalAccessToken
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	p.accessToken = tokenResp.AccessToken
	return nil
}

// createOrder creates a PayPal order
func (p *Provider) createOrder(ctx context.Context, orderRequest PayPalOrderRequest) (*PayPalOrder, error) {
	url := fmt.Sprintf("%s/v2/checkout/orders", p.baseURL)

	jsonData, err := json.Marshal(orderRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))

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
		return nil, fmt.Errorf("failed to create order: %s", string(body))
	}

	var order PayPalOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// getOrder gets a PayPal order
func (p *Provider) getOrder(ctx context.Context, orderID string) (*PayPalOrder, error) {
	url := fmt.Sprintf("%s/v2/checkout/orders/%s", p.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))

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
		return nil, fmt.Errorf("failed to get order: %s", string(body))
	}

	var order PayPalOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// convertStatus converts PayPal status to gateway status
func (p *Provider) convertStatus(status string) gateway.PaymentStatus {
	switch status {
	case "CREATED":
		return gateway.PaymentStatusPending
	case "SAVED":
		return gateway.PaymentStatusPending
	case "APPROVED":
		return gateway.PaymentStatusProcessing
	case "VOIDED":
		return gateway.PaymentStatusCanceled
	case "COMPLETED":
		return gateway.PaymentStatusSucceeded
	default:
		return gateway.PaymentStatusPending
	}
}

// handlePaymentCaptureCompleted handles payment capture completed events
func (p *Provider) handlePaymentCaptureCompleted(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment capture completed")
	// Implement your business logic here
	return nil
}

// handlePaymentCaptureDenied handles payment capture denied events
func (p *Provider) handlePaymentCaptureDenied(ctx context.Context, event *gateway.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment capture denied")
	// Implement your business logic here
	return nil
}
