package mocks

import (
	"context"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
)

// MockProvider is a mock implementation of MiddlewareProvider for testing
type MockProvider struct {
	NameValue                string
	TypeValue                string
	SupportedFeatures        []types.MiddlewareFeature
	ConnectionInfoValue      *types.ConnectionInfo
	ProcessRequestFunc       func(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error)
	ProcessResponseFunc      func(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error)
	CreateChainFunc          func(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error)
	ExecuteChainFunc         func(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error)
	CreateHTTPMiddlewareFunc func(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler
	WrapHTTPHandlerFunc      func(handler http.Handler, config *types.MiddlewareConfig) http.Handler
	ConfigureFunc            func(config map[string]interface{}) error
	IsConfiguredValue        bool
	HealthCheckFunc          func(ctx context.Context) error
	GetStatsFunc             func(ctx context.Context) (*types.MiddlewareStats, error)
	CloseFunc                func() error
}

// NewMockProvider creates a new mock provider with default values
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		NameValue: name,
		TypeValue: "mock",
		SupportedFeatures: []types.MiddlewareFeature{
			types.FeatureJWT,
			types.FeatureStructuredLogging,
		},
		ConnectionInfoValue: &types.ConnectionInfo{
			Host:     "localhost",
			Port:     8080,
			Protocol: "http",
			Version:  "1.0.0",
			Secure:   false,
		},
		ProcessRequestFunc: func(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return &types.MiddlewareResponse{
				ID:        request.ID,
				Success:   true,
				Timestamp: time.Now(),
			}, nil
		},
		ProcessResponseFunc: func(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
			return response, nil
		},
		CreateChainFunc: func(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
			return types.NewMiddlewareChain(), nil
		},
		ExecuteChainFunc: func(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return chain.Execute(ctx, request)
		},
		CreateHTTPMiddlewareFunc: func(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
			return func(next http.Handler) http.Handler {
				return next
			}
		},
		WrapHTTPHandlerFunc: func(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
			return handler
		},
		ConfigureFunc: func(config map[string]interface{}) error {
			return nil
		},
		IsConfiguredValue: true,
		HealthCheckFunc: func(ctx context.Context) error {
			return nil
		},
		GetStatsFunc: func(ctx context.Context) (*types.MiddlewareStats, error) {
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
					"provider_type": "mock",
				},
			}, nil
		},
		CloseFunc: func() error {
			return nil
		},
	}
}

// GetName returns the provider name
func (mp *MockProvider) GetName() string {
	return mp.NameValue
}

// GetType returns the provider type
func (mp *MockProvider) GetType() string {
	return mp.TypeValue
}

// GetSupportedFeatures returns supported features
func (mp *MockProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return mp.SupportedFeatures
}

// GetConnectionInfo returns connection information
func (mp *MockProvider) GetConnectionInfo() *types.ConnectionInfo {
	return mp.ConnectionInfoValue
}

// ProcessRequest processes a middleware request
func (mp *MockProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	return mp.ProcessRequestFunc(ctx, request)
}

// ProcessResponse processes a middleware response
func (mp *MockProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	return mp.ProcessResponseFunc(ctx, response)
}

// CreateChain creates a middleware chain
func (mp *MockProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	return mp.CreateChainFunc(ctx, config)
}

// ExecuteChain executes a middleware chain
func (mp *MockProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	return mp.ExecuteChainFunc(ctx, chain, request)
}

// CreateHTTPMiddleware creates HTTP middleware
func (mp *MockProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	return mp.CreateHTTPMiddlewareFunc(config)
}

// WrapHTTPHandler wraps an HTTP handler
func (mp *MockProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	return mp.WrapHTTPHandlerFunc(handler, config)
}

// Configure configures the provider
func (mp *MockProvider) Configure(config map[string]interface{}) error {
	return mp.ConfigureFunc(config)
}

// IsConfigured returns whether the provider is configured
func (mp *MockProvider) IsConfigured() bool {
	return mp.IsConfiguredValue
}

// HealthCheck performs health check
func (mp *MockProvider) HealthCheck(ctx context.Context) error {
	return mp.HealthCheckFunc(ctx)
}

// GetStats returns provider statistics
func (mp *MockProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	return mp.GetStatsFunc(ctx)
}

// Close closes the provider
func (mp *MockProvider) Close() error {
	return mp.CloseFunc()
}
