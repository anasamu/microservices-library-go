package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/payment"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/refund"
)

// Provider implements PaymentProvider for Stripe
type Provider struct {
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new Stripe payment provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "stripe"
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
func (p *Provider) GetSupportedMethods() []payment.PaymentMethod {
	return []payment.PaymentMethod{
		payment.PaymentMethodCard,
		payment.PaymentMethodBankTransfer,
		payment.PaymentMethodEWallet,
	}
}

// Configure configures the Stripe provider
func (p *Provider) Configure(config map[string]interface{}) error {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("stripe api_key is required")
	}

	stripe.Key = apiKey
	p.config = config

	p.logger.Info("Stripe provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return stripe.Key != ""
}

// CreatePayment creates a payment using Stripe
func (p *Provider) CreatePayment(ctx context.Context, request *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("stripe provider not configured")
	}

	// Create Stripe checkout session
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(request.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(request.Description),
					},
					UnitAmount: stripe.Int64(request.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(request.ReturnURL),
		CancelURL:  stripe.String(request.CancelURL),
		Metadata:   p.convertMetadata(request.Metadata),
	}

	// Add customer information
	if request.Customer != nil {
		sessionParams.CustomerEmail = stripe.String(request.Customer.Email)
		// Note: CustomerName is not supported in CheckoutSessionParams
		// Customer name can be added to metadata if needed
		if request.Customer.Name != "" {
			if sessionParams.Metadata == nil {
				sessionParams.Metadata = make(map[string]string)
			}
			sessionParams.Metadata["customer_name"] = request.Customer.Name
		}
	}

	// Create the session
	session, err := session.New(sessionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe session: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:         session.ID,
		Status:     payment.PaymentStatusPending,
		PaymentURL: session.URL,
		ProviderData: map[string]interface{}{
			"session_id":    session.ID,
			"client_secret": session.ClientSecret,
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// GetPayment retrieves a payment from Stripe
func (p *Provider) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("stripe provider not configured")
	}

	// Retrieve the session
	session, err := session.Get(paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe session: %w", err)
	}

	// Get payment intent if available
	var paymentIntent *stripe.PaymentIntent
	if session.PaymentIntent != nil {
		paymentIntent, err = paymentintent.Get(session.PaymentIntent.ID, nil)
		if err != nil {
			p.logger.WithError(err).Warn("Failed to get payment intent")
		}
	}

	// Extract description from line items
	description := ""
	if len(session.LineItems.Data) > 0 && session.LineItems.Data[0].Description != "" {
		description = session.LineItems.Data[0].Description
	}

	// Convert metadata from map[string]string to map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range session.Metadata {
		metadata[k] = v
	}

	paymentObj := &payment.Payment{
		ID:          session.ID,
		Amount:      session.AmountTotal,
		Currency:    string(session.Currency),
		Status:      p.convertStatus(session.PaymentStatus),
		Description: description,
		Metadata:    metadata,
		ProviderData: map[string]interface{}{
			"session_id":     session.ID,
			"payment_intent": session.PaymentIntent,
		},
		CreatedAt: time.Unix(session.Created, 0),
		UpdatedAt: time.Unix(session.Created, 0),
	}

	// Add customer information
	if session.CustomerEmail != "" {
		paymentObj.Customer = &payment.Customer{
			Email: session.CustomerEmail,
		}
	}

	// Add payment method
	if paymentIntent != nil {
		paymentObj.PaymentMethod = payment.PaymentMethodCard
		if paymentIntent.Status == stripe.PaymentIntentStatusSucceeded {
			paidAt := time.Unix(paymentIntent.Created, 0)
			paymentObj.PaidAt = &paidAt
		}
	}

	return paymentObj, nil
}

// CancelPayment cancels a payment in Stripe
func (p *Provider) CancelPayment(ctx context.Context, paymentID string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("stripe provider not configured")
	}

	// Expire the checkout session
	_, err := session.Expire(paymentID, nil)
	if err != nil {
		return fmt.Errorf("failed to expire stripe session: %w", err)
	}

	return nil
}

// RefundPayment processes a refund using Stripe
func (p *Provider) RefundPayment(ctx context.Context, request *payment.RefundRequest) (*payment.RefundResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("stripe provider not configured")
	}

	// Get the payment intent from the session
	session, err := session.Get(request.PaymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe session: %w", err)
	}

	if session.PaymentIntent == nil {
		return nil, fmt.Errorf("no payment intent found for session")
	}

	// Create refund parameters
	refundParams := &stripe.RefundParams{
		PaymentIntent: stripe.String(session.PaymentIntent.ID),
		Reason:        stripe.String("requested_by_customer"),
		Metadata:      p.convertMetadata(request.Metadata),
	}

	// Set amount if specified
	if request.Amount != nil {
		refundParams.Amount = stripe.Int64(*request.Amount)
	}

	// Create the refund
	refund, err := refund.New(refundParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe refund: %w", err)
	}

	response := &payment.RefundResponse{
		ID:        refund.ID,
		PaymentID: request.PaymentID,
		Amount:    refund.Amount,
		Status:    p.convertRefundStatus(refund.Status),
		Reason:    request.Reason,
		ProviderData: map[string]interface{}{
			"refund_id": refund.ID,
		},
		CreatedAt: time.Unix(refund.Created, 0),
	}

	return response, nil
}

// ValidateWebhook validates a Stripe webhook
func (p *Provider) ValidateWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("stripe provider not configured")
	}

	// Parse the webhook event
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// In a real implementation, you would validate the signature here
	// using stripe's webhook signature verification

	webhookEvent := &payment.WebhookEvent{
		ID:        event.ID,
		Type:      string(event.Type),
		Data:      make(map[string]interface{}),
		CreatedAt: time.Unix(event.Created, 0),
	}

	// Extract payment ID from event data
	switch event.Type {
	case "checkout.session.completed":
		if session, ok := event.Data.Object["session"].(map[string]interface{}); ok {
			if id, ok := session["id"].(string); ok {
				webhookEvent.PaymentID = id
			}
		}
	case "payment_intent.succeeded":
		if paymentIntent, ok := event.Data.Object["payment_intent"].(map[string]interface{}); ok {
			if id, ok := paymentIntent["id"].(string); ok {
				webhookEvent.PaymentID = id
			}
		}
	}

	// Store the raw event data
	webhookEvent.Data["raw_event"] = event.Data.Object

	return webhookEvent, nil
}

// ProcessWebhook processes a Stripe webhook event
func (p *Provider) ProcessWebhook(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithFields(logrus.Fields{
		"event_type": event.Type,
		"payment_id": event.PaymentID,
	}).Info("Processing Stripe webhook event")

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		return p.handleCheckoutSessionCompleted(ctx, event)
	case "payment_intent.succeeded":
		return p.handlePaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return p.handlePaymentIntentFailed(ctx, event)
	default:
		p.logger.WithField("event_type", event.Type).Debug("Unhandled webhook event type")
	}

	return nil
}

// handleCheckoutSessionCompleted handles checkout session completed events
func (p *Provider) handleCheckoutSessionCompleted(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Checkout session completed")
	// Implement your business logic here
	return nil
}

// handlePaymentIntentSucceeded handles payment intent succeeded events
func (p *Provider) handlePaymentIntentSucceeded(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment intent succeeded")
	// Implement your business logic here
	return nil
}

// handlePaymentIntentFailed handles payment intent failed events
func (p *Provider) handlePaymentIntentFailed(ctx context.Context, event *payment.WebhookEvent) error {
	p.logger.WithField("payment_id", event.PaymentID).Info("Payment intent failed")
	// Implement your business logic here
	return nil
}

// convertStatus converts Stripe status to payment status
func (p *Provider) convertStatus(status stripe.CheckoutSessionPaymentStatus) payment.PaymentStatus {
	switch status {
	case stripe.CheckoutSessionPaymentStatusPaid:
		return payment.PaymentStatusSucceeded
	case stripe.CheckoutSessionPaymentStatusUnpaid:
		return payment.PaymentStatusPending
	case stripe.CheckoutSessionPaymentStatusNoPaymentRequired:
		return payment.PaymentStatusSucceeded
	default:
		return payment.PaymentStatusPending
	}
}

// convertRefundStatus converts Stripe refund status to payment status
func (p *Provider) convertRefundStatus(status stripe.RefundStatus) payment.RefundStatus {
	switch status {
	case stripe.RefundStatusSucceeded:
		return payment.RefundStatusSucceeded
	case stripe.RefundStatusPending:
		return payment.RefundStatusPending
	case stripe.RefundStatusFailed:
		return payment.RefundStatusFailed
	case stripe.RefundStatusCanceled:
		return payment.RefundStatusCanceled
	default:
		return payment.RefundStatusPending
	}
}

// convertMetadata converts payment metadata to Stripe metadata
func (p *Provider) convertMetadata(metadata map[string]interface{}) map[string]string {
	stripeMetadata := make(map[string]string)
	for key, value := range metadata {
		stripeMetadata[key] = fmt.Sprintf("%v", value)
	}
	return stripeMetadata
}
