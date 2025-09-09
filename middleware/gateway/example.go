package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// ExampleMiddlewareManager demonstrates how to use the middleware manager
func ExampleMiddlewareManager() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create middleware manager
	manager := NewMiddlewareManager(DefaultManagerConfig(), logger)

	// Example: Register a custom middleware provider
	// Note: In real usage, you would register actual providers like auth, logging, etc.
	customProvider := &ExampleProvider{
		name:   "example",
		logger: logger,
	}

	if err := manager.RegisterProvider(customProvider); err != nil {
		log.Fatalf("Failed to register provider: %v", err)
	}

	// Example: Create middleware configuration
	config := &types.MiddlewareConfig{
		Name:       "example-chain",
		Type:       "authentication",
		Enabled:    true,
		Priority:   1,
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: 5 * time.Second,
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"service": "example-service",
		},
		Rules: []types.MiddlewareRule{
			{
				ID:          "rule-1",
				Name:        "auth-rule",
				Description: "Authentication rule",
				Enabled:     true,
				Priority:    1,
				Conditions: []types.MiddlewareCondition{
					{
						ID:       "condition-1",
						Type:     "header",
						Field:    "Authorization",
						Operator: "exists",
						Value:    nil,
					},
				},
				Actions: []types.MiddlewareAction{
					{
						ID:   "action-1",
						Type: "authenticate",
						Parameters: map[string]interface{}{
							"provider": "jwt",
						},
					},
				},
			},
		},
	}

	// Example: Create middleware chain
	ctx := context.Background()
	chain, err := manager.CreateChain(ctx, "example", config)
	if err != nil {
		log.Fatalf("Failed to create chain: %v", err)
	}

	// Example: Create middleware request
	request := &types.MiddlewareRequest{
		ID:     "req-123",
		Type:   "http",
		Method: "GET",
		Path:   "/api/users",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
		Body:      []byte(`{"user_id": "123"}`),
		Query:     map[string]string{"page": "1"},
		UserID:    "user123",
		ServiceID: "user-service",
		Context: map[string]interface{}{
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		},
		Metadata: map[string]interface{}{
			"request_source": "web",
		},
		Timestamp: time.Now(),
	}

	// Example: Execute middleware chain
	response, err := manager.ExecuteChain(ctx, "example", chain, request)
	if err != nil {
		log.Fatalf("Failed to execute chain: %v", err)
	}

	fmt.Printf("Middleware response: %+v\n", response)

	// Example: Create HTTP middleware
	httpMiddleware, err := manager.CreateHTTPMiddleware("example", config)
	if err != nil {
		log.Fatalf("Failed to create HTTP middleware: %v", err)
	}

	// Example: Wrap HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	wrappedHandler := httpMiddleware(handler)

	// Example: Setup HTTP server
	http.Handle("/api/", wrappedHandler)

	fmt.Println("HTTP server started on :8080")
	fmt.Println("Visit http://localhost:8080/api/ to test the middleware")

	// Uncomment to start the server
	// log.Fatal(http.ListenAndServe(":8080", nil))
}

// ExampleProvider is a simple example middleware provider
type ExampleProvider struct {
	name   string
	logger *logrus.Logger
}

func (ep *ExampleProvider) GetName() string {
	return ep.name
}

func (ep *ExampleProvider) GetType() string {
	return "example"
}

func (ep *ExampleProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureJWT,
		types.FeatureStructuredLogging,
		types.FeatureMetrics,
	}
}

func (ep *ExampleProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     8080,
		Protocol: "http",
		Version:  "1.0.0",
		Secure:   false,
	}
}

func (ep *ExampleProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	ep.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
	}).Info("Processing request")

	// Simple authentication check
	if request.Headers["Authorization"] == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "unauthorized"}`),
			Error:     "missing authorization header",
			Timestamp: time.Now(),
		}, nil
	}

	return &types.MiddlewareResponse{
		ID:         request.ID,
		Success:    true,
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:      request.Body,
		Message:   "request processed successfully",
		Timestamp: time.Now(),
	}, nil
}

func (ep *ExampleProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	ep.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Info("Processing response")

	// Add custom headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Processed-By"] = "example-middleware"
	response.Headers["X-Processed-At"] = time.Now().Format(time.RFC3339)

	return response, nil
}

func (ep *ExampleProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	ep.logger.WithField("chain_name", config.Name).Info("Creating middleware chain")

	// Create a simple chain with request and response processing
	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return ep.ProcessRequest(ctx, req)
		},
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			// This would typically process the response
			return nil, nil
		},
	)

	return chain, nil
}

func (ep *ExampleProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	ep.logger.WithField("request_id", request.ID).Info("Executing middleware chain")
	return chain.Execute(ctx, request)
}

func (ep *ExampleProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	ep.logger.WithField("config_type", config.Type).Info("Creating HTTP middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log the request
			ep.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
			}).Info("HTTP request received")

			// Simple authentication check
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized"}`))
				return
			}

			// Add custom headers
			w.Header().Set("X-Processed-By", "example-middleware")
			w.Header().Set("X-Processed-At", time.Now().Format(time.RFC3339))

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

func (ep *ExampleProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := ep.CreateHTTPMiddleware(config)
	return middleware(handler)
}

func (ep *ExampleProvider) Configure(config map[string]interface{}) error {
	ep.logger.Info("Configuring example provider")
	return nil
}

func (ep *ExampleProvider) IsConfigured() bool {
	return true
}

func (ep *ExampleProvider) HealthCheck(ctx context.Context) error {
	ep.logger.Debug("Health check for example provider")
	return nil
}

func (ep *ExampleProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	return &types.MiddlewareStats{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		AverageLatency:     10 * time.Millisecond,
		MaxLatency:         50 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         100.0,
		ActiveConnections:  10,
		ProviderData: map[string]interface{}{
			"provider_type": "example",
			"version":       "1.0.0",
		},
	}, nil
}

func (ep *ExampleProvider) Close() error {
	ep.logger.Info("Closing example provider")
	return nil
}
