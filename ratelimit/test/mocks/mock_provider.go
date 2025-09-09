package mocks

import (
	"context"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit/types"
)

// MockProvider is a mock implementation of RateLimitProvider for testing
type MockProvider struct {
	Name              string
	SupportedFeatures []types.RateLimitFeature
	ConnectionInfo    *types.ConnectionInfo
	Connected         bool
	AllowFunc         func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error)
	ResetFunc         func(ctx context.Context, key string) error
	GetRemainingFunc  func(ctx context.Context, key string, limit *types.RateLimit) (int64, error)
	GetResetTimeFunc  func(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error)
	AllowMultipleFunc func(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error)
	ResetMultipleFunc func(ctx context.Context, keys []string) error
	GetStatsFunc      func(ctx context.Context) (*types.RateLimitStats, error)
	GetKeysFunc       func(ctx context.Context, pattern string) ([]string, error)
	ConnectFunc       func(ctx context.Context) error
	DisconnectFunc    func(ctx context.Context) error
	PingFunc          func(ctx context.Context) error
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		Name: "mock",
		SupportedFeatures: []types.RateLimitFeature{
			types.FeatureAllow,
			types.FeatureReset,
			types.FeatureRemaining,
			types.FeatureResetTime,
			types.FeatureStats,
			types.FeatureBatch,
		},
		ConnectionInfo: &types.ConnectionInfo{
			Host:   "mock",
			Port:   0,
			Status: types.StatusConnected,
		},
		Connected: true,
	}
}

// GetName returns the provider name
func (m *MockProvider) GetName() string {
	return m.Name
}

// GetSupportedFeatures returns supported features
func (m *MockProvider) GetSupportedFeatures() []types.RateLimitFeature {
	return m.SupportedFeatures
}

// GetConnectionInfo returns connection information
func (m *MockProvider) GetConnectionInfo() *types.ConnectionInfo {
	return m.ConnectionInfo
}

// Connect establishes connection
func (m *MockProvider) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx)
	}
	m.Connected = true
	return nil
}

// Disconnect closes the connection
func (m *MockProvider) Disconnect(ctx context.Context) error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc(ctx)
	}
	m.Connected = false
	return nil
}

// Ping checks if provider is reachable
func (m *MockProvider) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	if !m.Connected {
		return &types.RateLimitError{
			Code:    types.ErrCodeConnection,
			Message: "provider not connected",
		}
	}
	return nil
}

// IsConnected returns connection status
func (m *MockProvider) IsConnected() bool {
	return m.Connected
}

// Allow checks if a request is allowed
func (m *MockProvider) Allow(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
	if m.AllowFunc != nil {
		return m.AllowFunc(ctx, key, limit)
	}

	// Default behavior: allow all requests
	return &types.RateLimitResult{
		Allowed:    true,
		Limit:      limit.Limit,
		Remaining:  limit.Limit - 1,
		ResetTime:  time.Now().Add(limit.Window),
		RetryAfter: 0,
		Key:        key,
	}, nil
}

// Reset resets the rate limit for a key
func (m *MockProvider) Reset(ctx context.Context, key string) error {
	if m.ResetFunc != nil {
		return m.ResetFunc(ctx, key)
	}
	return nil
}

// GetRemaining returns the remaining requests for a key
func (m *MockProvider) GetRemaining(ctx context.Context, key string, limit *types.RateLimit) (int64, error) {
	if m.GetRemainingFunc != nil {
		return m.GetRemainingFunc(ctx, key, limit)
	}
	return limit.Limit, nil
}

// GetResetTime returns the reset time for a key
func (m *MockProvider) GetResetTime(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error) {
	if m.GetResetTimeFunc != nil {
		return m.GetResetTimeFunc(ctx, key, limit)
	}
	return time.Now().Add(limit.Window), nil
}

// AllowMultiple checks multiple rate limit requests
func (m *MockProvider) AllowMultiple(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error) {
	if m.AllowMultipleFunc != nil {
		return m.AllowMultipleFunc(ctx, requests)
	}

	results := make([]*types.RateLimitResult, len(requests))
	for i, req := range requests {
		result, err := m.Allow(ctx, req.Key, req.Limit)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}

	return results, nil
}

// ResetMultiple resets multiple rate limits
func (m *MockProvider) ResetMultiple(ctx context.Context, keys []string) error {
	if m.ResetMultipleFunc != nil {
		return m.ResetMultipleFunc(ctx, keys)
	}
	return nil
}

// GetStats returns rate limit statistics
func (m *MockProvider) GetStats(ctx context.Context) (*types.RateLimitStats, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}

	return &types.RateLimitStats{
		TotalRequests:   100,
		AllowedRequests: 90,
		BlockedRequests: 10,
		ActiveKeys:      5,
		Memory:          1024,
		Uptime:          time.Hour,
		LastUpdate:      time.Now(),
		Provider:        m.Name,
	}, nil
}

// GetKeys returns keys matching a pattern
func (m *MockProvider) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	if m.GetKeysFunc != nil {
		return m.GetKeysFunc(ctx, pattern)
	}

	return []string{"key1", "key2", "key3"}, nil
}

// SetAllowFunc sets a custom Allow function
func (m *MockProvider) SetAllowFunc(fn func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error)) {
	m.AllowFunc = fn
}

// SetResetFunc sets a custom Reset function
func (m *MockProvider) SetResetFunc(fn func(ctx context.Context, key string) error) {
	m.ResetFunc = fn
}

// SetGetRemainingFunc sets a custom GetRemaining function
func (m *MockProvider) SetGetRemainingFunc(fn func(ctx context.Context, key string, limit *types.RateLimit) (int64, error)) {
	m.GetRemainingFunc = fn
}

// SetGetResetTimeFunc sets a custom GetResetTime function
func (m *MockProvider) SetGetResetTimeFunc(fn func(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error)) {
	m.GetResetTimeFunc = fn
}

// SetAllowMultipleFunc sets a custom AllowMultiple function
func (m *MockProvider) SetAllowMultipleFunc(fn func(ctx context.Context, requests []*types.RateLimitRequest) ([]*types.RateLimitResult, error)) {
	m.AllowMultipleFunc = fn
}

// SetResetMultipleFunc sets a custom ResetMultiple function
func (m *MockProvider) SetResetMultipleFunc(fn func(ctx context.Context, keys []string) error) {
	m.ResetMultipleFunc = fn
}

// SetGetStatsFunc sets a custom GetStats function
func (m *MockProvider) SetGetStatsFunc(fn func(ctx context.Context) (*types.RateLimitStats, error)) {
	m.GetStatsFunc = fn
}

// SetGetKeysFunc sets a custom GetKeys function
func (m *MockProvider) SetGetKeysFunc(fn func(ctx context.Context, pattern string) ([]string, error)) {
	m.GetKeysFunc = fn
}

// SetConnectFunc sets a custom Connect function
func (m *MockProvider) SetConnectFunc(fn func(ctx context.Context) error) {
	m.ConnectFunc = fn
}

// SetDisconnectFunc sets a custom Disconnect function
func (m *MockProvider) SetDisconnectFunc(fn func(ctx context.Context) error) {
	m.DisconnectFunc = fn
}

// SetPingFunc sets a custom Ping function
func (m *MockProvider) SetPingFunc(fn func(ctx context.Context) error) {
	m.PingFunc = fn
}

// SetConnected sets the connection status
func (m *MockProvider) SetConnected(connected bool) {
	m.Connected = connected
}

// SetName sets the provider name
func (m *MockProvider) SetName(name string) {
	m.Name = name
}

// SetSupportedFeatures sets the supported features
func (m *MockProvider) SetSupportedFeatures(features []types.RateLimitFeature) {
	m.SupportedFeatures = features
}

// SetConnectionInfo sets the connection info
func (m *MockProvider) SetConnectionInfo(info *types.ConnectionInfo) {
	m.ConnectionInfo = info
}
