package gateway

import (
	"context"
	"fmt"
	"log"

	"github.com/anasamu/microservices-library-go/libs/payment/providers/midtrans"
	"github.com/anasamu/microservices-library-go/libs/payment/providers/paypal"
	"github.com/anasamu/microservices-library-go/libs/payment/providers/stripe"
	"github.com/anasamu/microservices-library-go/libs/payment/providers/xendit"
	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the payment gateway system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create payment manager
	config := DefaultManagerConfig()
	config.DefaultProvider = "stripe"

	paymentManager := NewPaymentManager(config, logger)

	// Register payment providers
	registerProviders(paymentManager, logger)

	// Example 1: Create payment with Stripe
	exampleStripePayment(paymentManager)

	// Example 2: Create payment with PayPal
	examplePayPalPayment(paymentManager)

	// Example 3: Create payment with Xendit
	exampleXenditPayment(paymentManager)

	// Example 4: Create payment with Midtrans
	exampleMidtransPayment(paymentManager)

	// Example 5: Webhook processing
	exampleWebhookProcessing(paymentManager)
}

// registerProviders registers all payment providers
func registerProviders(paymentManager *PaymentManager, logger *logrus.Logger) {
	// Register Stripe provider
	stripeProvider := stripe.NewProvider(logger)
	stripeConfig := map[string]interface{}{
		"api_key": "sk_test_your_stripe_secret_key",
	}
	if err := stripeProvider.Configure(stripeConfig); err != nil {
		log.Printf("Failed to configure Stripe: %v", err)
	} else {
		paymentManager.RegisterProvider(stripeProvider)
	}

	// Register PayPal provider
	paypalProvider := paypal.NewProvider(logger)
	paypalConfig := map[string]interface{}{
		"client_id":     "your_paypal_client_id",
		"client_secret": "your_paypal_client_secret",
		"sandbox":       true,
	}
	if err := paypalProvider.Configure(paypalConfig); err != nil {
		log.Printf("Failed to configure PayPal: %v", err)
	} else {
		paymentManager.RegisterProvider(paypalProvider)
	}

	// Register Xendit provider
	xenditProvider := xendit.NewProvider(logger)
	xenditConfig := map[string]interface{}{
		"api_key": "xnd_public_development_your_xendit_api_key",
	}
	if err := xenditProvider.Configure(xenditConfig); err != nil {
		log.Printf("Failed to configure Xendit: %v", err)
	} else {
		paymentManager.RegisterProvider(xenditProvider)
	}

	// Register Midtrans provider
	midtransProvider := midtrans.NewProvider(logger)
	midtransConfig := map[string]interface{}{
		"server_key":    "SB-Mid-server-your_midtrans_server_key",
		"client_key":    "SB-Mid-client-your_midtrans_client_key",
		"is_production": false,
	}
	if err := midtransProvider.Configure(midtransConfig); err != nil {
		log.Printf("Failed to configure Midtrans: %v", err)
	} else {
		paymentManager.RegisterProvider(midtransProvider)
	}
}

// exampleStripePayment demonstrates Stripe payment creation
func exampleStripePayment(paymentManager *PaymentManager) {
	fmt.Println("=== Stripe Payment Example ===")

	ctx := context.Background()

	// Create payment request
	request := &PaymentRequest{
		Amount:      2000, // $20.00 in cents
		Currency:    "USD",
		Description: "Test payment with Stripe",
		Customer: &Customer{
			Email: "customer@example.com",
			Name:  "John Doe",
		},
		PaymentMethod: PaymentMethodCard,
		ReturnURL:     "https://example.com/success",
		CancelURL:     "https://example.com/cancel",
		Metadata: map[string]interface{}{
			"order_id": "order_123",
			"source":   "web",
		},
	}

	// Create payment
	response, err := paymentManager.CreatePayment(ctx, "stripe", request)
	if err != nil {
		log.Printf("Failed to create Stripe payment: %v", err)
		return
	}

	fmt.Printf("Stripe Payment Created:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Payment URL: %s\n", response.PaymentURL)
	fmt.Printf("  Client Secret: %s\n", response.ClientSecret)

	// Get payment status
	payment, err := paymentManager.GetPayment(ctx, "stripe", response.ID)
	if err != nil {
		log.Printf("Failed to get Stripe payment: %v", err)
		return
	}

	fmt.Printf("Payment Status: %s\n", payment.Status)
}

// examplePayPalPayment demonstrates PayPal payment creation
func examplePayPalPayment(paymentManager *PaymentManager) {
	fmt.Println("\n=== PayPal Payment Example ===")

	ctx := context.Background()

	// Create payment request
	request := &PaymentRequest{
		Amount:      2500, // $25.00 in cents
		Currency:    "USD",
		Description: "Test payment with PayPal",
		Customer: &Customer{
			Email: "customer@example.com",
			Name:  "Jane Doe",
		},
		PaymentMethod: PaymentMethodCard,
		ReturnURL:     "https://example.com/success",
		CancelURL:     "https://example.com/cancel",
		Metadata: map[string]interface{}{
			"order_id": "order_456",
			"source":   "mobile",
		},
	}

	// Create payment
	response, err := paymentManager.CreatePayment(ctx, "paypal", request)
	if err != nil {
		log.Printf("Failed to create PayPal payment: %v", err)
		return
	}

	fmt.Printf("PayPal Payment Created:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Payment URL: %s\n", response.PaymentURL)
}

// exampleXenditPayment demonstrates Xendit payment creation
func exampleXenditPayment(paymentManager *PaymentManager) {
	fmt.Println("\n=== Xendit Payment Example ===")

	ctx := context.Background()

	// Create payment request
	request := &PaymentRequest{
		Amount:      100000, // 100,000 IDR in cents
		Currency:    "IDR",
		Description: "Test payment with Xendit",
		Customer: &Customer{
			Email: "customer@example.com",
			Name:  "Ahmad Wijaya",
		},
		PaymentMethod: PaymentMethodCard,
		ReturnURL:     "https://example.com/success",
		CancelURL:     "https://example.com/cancel",
		Metadata: map[string]interface{}{
			"order_id": "order_789",
			"source":   "web",
		},
	}

	// Create payment
	response, err := paymentManager.CreatePayment(ctx, "xendit", request)
	if err != nil {
		log.Printf("Failed to create Xendit payment: %v", err)
		return
	}

	fmt.Printf("Xendit Payment Created:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Payment URL: %s\n", response.PaymentURL)
}

// exampleMidtransPayment demonstrates Midtrans payment creation
func exampleMidtransPayment(paymentManager *PaymentManager) {
	fmt.Println("\n=== Midtrans Payment Example ===")

	ctx := context.Background()

	// Create payment request
	request := &PaymentRequest{
		Amount:      50000, // 50,000 IDR in cents
		Currency:    "IDR",
		Description: "Test payment with Midtrans",
		Customer: &Customer{
			Email: "customer@example.com",
			Name:  "Siti Nurhaliza",
			Phone: "+6281234567890",
		},
		PaymentMethod: PaymentMethodCard,
		ReturnURL:     "https://example.com/success",
		CancelURL:     "https://example.com/cancel",
		Metadata: map[string]interface{}{
			"order_id": "order_101",
			"source":   "mobile",
		},
	}

	// Create payment
	response, err := paymentManager.CreatePayment(ctx, "midtrans", request)
	if err != nil {
		log.Printf("Failed to create Midtrans payment: %v", err)
		return
	}

	fmt.Printf("Midtrans Payment Created:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Payment URL: %s\n", response.PaymentURL)
}

// exampleWebhookProcessing demonstrates webhook processing
func exampleWebhookProcessing(paymentManager *PaymentManager) {
	fmt.Println("\n=== Webhook Processing Example ===")

	ctx := context.Background()

	// Simulate webhook payloads for different providers
	webhookExamples := map[string]struct {
		payload   []byte
		signature string
	}{
		"stripe": {
			payload:   []byte(`{"id":"evt_123","type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`),
			signature: "t=1234567890,v1=signature_hash",
		},
		"paypal": {
			payload:   []byte(`{"id":"WH-123","event_type":"PAYMENT.CAPTURE.COMPLETED","resource":{"id":"capture_123"}}`),
			signature: "signature_hash",
		},
		"xendit": {
			payload:   []byte(`{"id":"evt_123","event":"invoice.paid","data":{"id":"inv_123","status":"PAID"}}`),
			signature: "signature_hash",
		},
		"midtrans": {
			payload:   []byte(`{"transaction_time":"2023-01-01 12:00:00","transaction_status":"settlement","order_id":"order_123"}`),
			signature: "signature_hash",
		},
	}

	// Process webhooks for each provider
	for provider, webhook := range webhookExamples {
		fmt.Printf("Processing %s webhook...\n", provider)

		err := paymentManager.ProcessWebhook(ctx, provider, webhook.payload, webhook.signature)
		if err != nil {
			log.Printf("Failed to process %s webhook: %v", provider, err)
		} else {
			fmt.Printf("  %s webhook processed successfully\n", provider)
		}
	}
}

// ExampleRefund demonstrates refund processing
func ExampleRefund() {
	fmt.Println("\n=== Refund Example ===")

	// Create logger and payment manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	paymentManager := NewPaymentManager(config, logger)

	// Register providers
	registerProviders(paymentManager, logger)

	ctx := context.Background()

	// Create refund request
	refundRequest := &RefundRequest{
		PaymentID: "payment_123",
		Amount:    int64Ptr(1000), // Refund $10.00
		Reason:    "Customer requested refund",
		Metadata: map[string]interface{}{
			"refund_reason": "customer_request",
		},
	}

	// Process refund with different providers
	providers := []string{"stripe", "paypal", "xendit", "midtrans"}

	for _, provider := range providers {
		fmt.Printf("Processing refund with %s...\n", provider)

		response, err := paymentManager.RefundPayment(ctx, provider, refundRequest)
		if err != nil {
			log.Printf("Failed to process %s refund: %v", provider, err)
		} else {
			fmt.Printf("  Refund ID: %s\n", response.ID)
			fmt.Printf("  Amount: %d\n", response.Amount)
			fmt.Printf("  Status: %s\n", response.Status)
		}
	}
}

// ExampleProviderCapabilities demonstrates checking provider capabilities
func ExampleProviderCapabilities() {
	fmt.Println("\n=== Provider Capabilities Example ===")

	// Create logger and payment manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	paymentManager := NewPaymentManager(config, logger)

	// Register providers
	registerProviders(paymentManager, logger)

	// Get supported providers
	providers := paymentManager.GetSupportedProviders()
	fmt.Printf("Supported providers: %v\n", providers)

	// Check capabilities for each provider
	for _, providerName := range providers {
		fmt.Printf("\n%s capabilities:\n", providerName)

		methods, currencies, err := paymentManager.GetProviderCapabilities(providerName)
		if err != nil {
			log.Printf("Failed to get %s capabilities: %v", providerName, err)
			continue
		}

		fmt.Printf("  Payment methods: %v\n", methods)
		fmt.Printf("  Currencies: %v\n", currencies)
	}
}

// int64Ptr returns a pointer to an int64
func int64Ptr(i int64) *int64 {
	return &i
}
