package unit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/middleware"
	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareManager(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewMiddlewareManager(gateway.DefaultManagerConfig(), logger)

	t.Run("RegisterProvider", func(t *testing.T) {
		provider := &MockProvider{
			name: "test-provider",
		}

		err := manager.RegisterProvider(provider)
		assert.NoError(t, err)

		// Try to register same provider again
		err = manager.RegisterProvider(provider)
		assert.NoError(t, err)
	})

	t.Run("GetProvider", func(t *testing.T) {
		provider, err := manager.GetProvider("test-provider")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "test-provider", provider.GetName())

		// Try to get non-existent provider
		_, err = manager.GetProvider("non-existent")
		assert.Error(t, err)
	})

	t.Run("ProcessRequest", func(t *testing.T) {
		request := &types.MiddlewareRequest{
			ID:        "test-request",
			Type:      "http",
			Method:    "GET",
			Path:      "/test",
			Headers:   make(map[string]string),
			Query:     make(map[string]string),
			Context:   make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		response, err := manager.ProcessRequest(context.Background(), "test-provider", request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
	})

	t.Run("CreateChain", func(t *testing.T) {
		config := &types.MiddlewareConfig{
			Name:     "test-chain",
			Type:     "test",
			Enabled:  true,
			Priority: 1,
			Timeout:  30 * time.Second,
			Metadata: make(map[string]interface{}),
		}

		chain, err := manager.CreateChain(context.Background(), "test-provider", config)
		assert.NoError(t, err)
		assert.NotNil(t, chain)

		// Test getting the chain
		retrievedChain, err := manager.GetChain("test-chain")
		assert.NoError(t, err)
		assert.Equal(t, chain, retrievedChain)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		results := manager.HealthCheck(context.Background())
		assert.NotNil(t, results)
		assert.Contains(t, results, "test-provider")
		assert.NoError(t, results["test-provider"])
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := manager.GetStats(context.Background())
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "test-provider")
	})

	t.Run("Close", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)
	})
}

func TestDefaultManagerConfig(t *testing.T) {
	config := gateway.DefaultManagerConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "default", config.DefaultProvider)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 5*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

// MockProvider is a mock implementation of MiddlewareProvider for testing
type MockProvider struct {
	name string
}

func (mp *MockProvider) GetName() string {
	return mp.name
}

func (mp *MockProvider) GetType() string {
	return "mock"
}

func (mp *MockProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureJWT,
		types.FeatureStructuredLogging,
	}
}

func (mp *MockProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     8080,
		Protocol: "http",
		Version:  "1.0.0",
		Secure:   false,
	}
}

func (mp *MockProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Timestamp: time.Now(),
	}, nil
}

func (mp *MockProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	return response, nil
}

func (mp *MockProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	return types.NewMiddlewareChain(), nil
}

func (mp *MockProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	return chain.Execute(ctx, request)
}

func (mp *MockProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

func (mp *MockProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	return handler
}

func (mp *MockProvider) Configure(config map[string]interface{}) error {
	return nil
}

func (mp *MockProvider) IsConfigured() bool {
	return true
}

func (mp *MockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (mp *MockProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
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
}

func (mp *MockProvider) Close() error {
	return nil
}
