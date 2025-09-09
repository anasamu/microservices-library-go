package mocks

import (
	"context"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/stretchr/testify/mock"
)

// MockCircuitBreakerProvider is a mock implementation of CircuitBreakerProvider
type MockCircuitBreakerProvider struct {
	mock.Mock
}

// GetName returns the provider name
func (m *MockCircuitBreakerProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

// GetSupportedFeatures returns the features supported by the provider
func (m *MockCircuitBreakerProvider) GetSupportedFeatures() []types.CircuitBreakerFeature {
	args := m.Called()
	return args.Get(0).([]types.CircuitBreakerFeature)
}

// GetConnectionInfo returns connection information
func (m *MockCircuitBreakerProvider) GetConnectionInfo() *types.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(*types.ConnectionInfo)
}

// Connect establishes connection
func (m *MockCircuitBreakerProvider) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Disconnect closes the connection
func (m *MockCircuitBreakerProvider) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Ping tests the connection
func (m *MockCircuitBreakerProvider) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IsConnected returns the connection status
func (m *MockCircuitBreakerProvider) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// Execute runs a function through the circuit breaker
func (m *MockCircuitBreakerProvider) Execute(ctx context.Context, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error) {
	args := m.Called(ctx, name, fn)
	return args.Get(0).(*types.ExecutionResult), args.Error(1)
}

// ExecuteWithFallback runs a function with fallback through the circuit breaker
func (m *MockCircuitBreakerProvider) ExecuteWithFallback(ctx context.Context, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error) {
	args := m.Called(ctx, name, fn, fallback)
	return args.Get(0).(*types.ExecutionResult), args.Error(1)
}

// GetState returns the state of a circuit breaker
func (m *MockCircuitBreakerProvider) GetState(ctx context.Context, name string) (types.CircuitState, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(types.CircuitState), args.Error(1)
}

// GetStats returns circuit breaker statistics
func (m *MockCircuitBreakerProvider) GetStats(ctx context.Context, name string) (*types.CircuitBreakerStats, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*types.CircuitBreakerStats), args.Error(1)
}

// Reset resets a circuit breaker
func (m *MockCircuitBreakerProvider) Reset(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Configure configures a circuit breaker
func (m *MockCircuitBreakerProvider) Configure(ctx context.Context, name string, config *types.CircuitBreakerConfig) error {
	args := m.Called(ctx, name, config)
	return args.Error(0)
}

// GetConfig returns circuit breaker configuration
func (m *MockCircuitBreakerProvider) GetConfig(ctx context.Context, name string) (*types.CircuitBreakerConfig, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*types.CircuitBreakerConfig), args.Error(1)
}

// GetAllStates returns states of all circuit breakers
func (m *MockCircuitBreakerProvider) GetAllStates(ctx context.Context) (map[string]types.CircuitState, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]types.CircuitState), args.Error(1)
}

// GetAllStats returns statistics of all circuit breakers
func (m *MockCircuitBreakerProvider) GetAllStats(ctx context.Context) (map[string]*types.CircuitBreakerStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]*types.CircuitBreakerStats), args.Error(1)
}

// ResetAll resets all circuit breakers
func (m *MockCircuitBreakerProvider) ResetAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Close closes the provider
func (m *MockCircuitBreakerProvider) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper functions for creating mock data

// NewMockExecutionResult creates a mock execution result
func NewMockExecutionResult(result interface{}, err error, state types.CircuitState) *types.ExecutionResult {
	return &types.ExecutionResult{
		Result:   result,
		Error:    err,
		Duration: 100 * time.Millisecond,
		State:    state,
		Fallback: false,
		Retry:    false,
		Attempts: 1,
	}
}

// NewMockExecutionResultWithFallback creates a mock execution result with fallback
func NewMockExecutionResultWithFallback(result interface{}, err error, state types.CircuitState, fallback bool) *types.ExecutionResult {
	return &types.ExecutionResult{
		Result:   result,
		Error:    err,
		Duration: 100 * time.Millisecond,
		State:    state,
		Fallback: fallback,
		Retry:    false,
		Attempts: 1,
	}
}

// NewMockCircuitBreakerStats creates mock circuit breaker statistics
func NewMockCircuitBreakerStats(provider string, state types.CircuitState) *types.CircuitBreakerStats {
	return &types.CircuitBreakerStats{
		Requests:             10,
		Successes:            8,
		Failures:             2,
		Timeouts:             0,
		Rejects:              0,
		State:                state,
		LastFailureTime:      time.Now().Add(-1 * time.Minute),
		LastSuccessTime:      time.Now().Add(-10 * time.Second),
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 3,
		Uptime:               5 * time.Minute,
		LastUpdate:           time.Now(),
		Provider:             provider,
	}
}

// NewMockConnectionInfo creates mock connection information
func NewMockConnectionInfo(status types.ConnectionStatus) *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     8080,
		Database: "test",
		Username: "test",
		Status:   status,
		Metadata: map[string]string{
			"version": "1.0.0",
			"type":    "mock",
		},
	}
}

// NewMockCircuitBreakerConfig creates mock circuit breaker configuration
func NewMockCircuitBreakerConfig(name string) *types.CircuitBreakerConfig {
	return &types.CircuitBreakerConfig{
		Name:                name,
		MaxRequests:         10,
		Interval:            10 * time.Second,
		Timeout:             60 * time.Second,
		MaxConsecutiveFails: 5,
		FailureThreshold:    0.5,
		SuccessThreshold:    3,
		FallbackEnabled:     true,
		RetryEnabled:        true,
		RetryAttempts:       3,
		RetryDelay:          time.Second,
		ReadyToTrip: func(counts types.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
		Metadata: map[string]string{
			"service": name,
			"version": "1.0.0",
		},
	}
}
