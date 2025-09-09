package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// MiddlewareManager manages multiple middleware providers
type MiddlewareManager struct {
	providers map[string]MiddlewareProvider
	logger    *logrus.Logger
	config    *ManagerConfig
	chains    map[string]*types.MiddlewareChain
}

// ManagerConfig holds middleware manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	Metadata        map[string]string `json:"metadata"`
}

// MiddlewareProvider interface for middleware backends
type MiddlewareProvider interface {
	// Provider information
	GetName() string
	GetType() string
	GetSupportedFeatures() []types.MiddlewareFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Middleware operations
	ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error)
	ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error)
	CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error)
	ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error)

	// HTTP middleware
	CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler
	WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler

	// Configuration and management
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*types.MiddlewareStats, error)
	Close() error
}

// DefaultManagerConfig returns default middleware manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "default",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		Metadata:        make(map[string]string),
	}
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager(config *ManagerConfig, logger *logrus.Logger) *MiddlewareManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &MiddlewareManager{
		providers: make(map[string]MiddlewareProvider),
		logger:    logger,
		config:    config,
		chains:    make(map[string]*types.MiddlewareChain),
	}
}

// RegisterProvider registers a middleware provider
func (mm *MiddlewareManager) RegisterProvider(provider MiddlewareProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	mm.providers[name] = provider
	mm.logger.WithField("provider", name).Info("Middleware provider registered")

	return nil
}

// GetProvider returns a middleware provider by name
func (mm *MiddlewareManager) GetProvider(name string) (MiddlewareProvider, error) {
	provider, exists := mm.providers[name]
	if !exists {
		return nil, fmt.Errorf("middleware provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default middleware provider
func (mm *MiddlewareManager) GetDefaultProvider() (MiddlewareProvider, error) {
	return mm.GetProvider(mm.config.DefaultProvider)
}

// ProcessRequest processes a request through the specified provider
func (mm *MiddlewareManager) ProcessRequest(ctx context.Context, providerName string, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := mm.validateMiddlewareRequest(request); err != nil {
		return nil, fmt.Errorf("invalid middleware request: %w", err)
	}

	response, err := provider.ProcessRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to process request: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"request_id": request.ID,
		"success":    response.Success,
	}).Debug("Request processing completed")

	return response, nil
}

// ProcessResponse processes a response through the specified provider
func (mm *MiddlewareManager) ProcessResponse(ctx context.Context, providerName string, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	processedResponse, err := provider.ProcessResponse(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("failed to process response: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"response_id": response.ID,
		"success":     processedResponse.Success,
	}).Debug("Response processing completed")

	return processedResponse, nil
}

// CreateChain creates a middleware chain using the specified provider
func (mm *MiddlewareManager) CreateChain(ctx context.Context, providerName string, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	chain, err := provider.CreateChain(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware chain: %w", err)
	}

	// Store chain for later use
	chainName := config.Name
	if chainName == "" {
		chainName = fmt.Sprintf("chain_%d", time.Now().UnixNano())
	}
	mm.chains[chainName] = chain

	mm.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"chain_name":  chainName,
		"config_type": config.Type,
	}).Info("Middleware chain created")

	return chain, nil
}

// ExecuteChain executes a middleware chain
func (mm *MiddlewareManager) ExecuteChain(ctx context.Context, providerName string, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.ExecuteChain(ctx, chain, request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute middleware chain: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":   providerName,
		"request_id": request.ID,
		"success":    response.Success,
	}).Debug("Middleware chain execution completed")

	return response, nil
}

// GetChain returns a middleware chain by name
func (mm *MiddlewareManager) GetChain(name string) (*types.MiddlewareChain, error) {
	chain, exists := mm.chains[name]
	if !exists {
		return nil, fmt.Errorf("middleware chain not found: %s", name)
	}
	return chain, nil
}

// CreateHTTPMiddleware creates HTTP middleware using the specified provider
func (mm *MiddlewareManager) CreateHTTPMiddleware(providerName string, config *types.MiddlewareConfig) (types.HTTPMiddlewareHandler, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	middleware := provider.CreateHTTPMiddleware(config)
	if middleware == nil {
		return nil, fmt.Errorf("failed to create HTTP middleware")
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"config_type": config.Type,
	}).Info("HTTP middleware created")

	return middleware, nil
}

// WrapHTTPHandler wraps an HTTP handler with middleware
func (mm *MiddlewareManager) WrapHTTPHandler(providerName string, handler http.Handler, config *types.MiddlewareConfig) (http.Handler, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	wrappedHandler := provider.WrapHTTPHandler(handler, config)
	if wrappedHandler == nil {
		return nil, fmt.Errorf("failed to wrap HTTP handler")
	}

	mm.logger.WithFields(logrus.Fields{
		"provider":    providerName,
		"config_type": config.Type,
	}).Info("HTTP handler wrapped with middleware")

	return wrappedHandler, nil
}

// validateMiddlewareRequest validates a middleware request
func (mm *MiddlewareManager) validateMiddlewareRequest(request *types.MiddlewareRequest) error {
	if request == nil {
		return fmt.Errorf("middleware request cannot be nil")
	}

	if request.ID == "" {
		return fmt.Errorf("request ID is required")
	}

	if request.Type == "" {
		return fmt.Errorf("request type is required")
	}

	return nil
}

// HealthCheck performs health check on all providers
func (mm *MiddlewareManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range mm.providers {
		results[name] = provider.HealthCheck(ctx)
	}

	return results
}

// GetStats returns statistics for all providers
func (mm *MiddlewareManager) GetStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	for name, provider := range mm.providers {
		if providerStats, err := provider.GetStats(ctx); err == nil {
			stats[name] = providerStats
		}
	}

	return stats
}

// Close closes all providers
func (mm *MiddlewareManager) Close() error {
	var errors []error

	for name, provider := range mm.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	mm.logger.Info("All middleware providers closed")
	return nil
}
