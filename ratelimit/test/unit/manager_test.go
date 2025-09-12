package unit

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/ratelimit"
	"github.com/anasamu/microservices-library-go/ratelimit/test/mocks"
	"github.com/anasamu/microservices-library-go/ratelimit/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimitManager(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	assert.NotNil(t, manager)
	assert.Equal(t, 3, manager.GetConfig().RetryAttempts)
	assert.Equal(t, time.Second, manager.GetConfig().RetryDelay)
	assert.Equal(t, 30*time.Second, manager.GetConfig().Timeout)
	assert.True(t, manager.GetConfig().FallbackEnabled)
}

func TestRegisterProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()

	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	// Try to register the same provider again
	err = manager.RegisterProvider(mockProvider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Try to register nil provider
	err = manager.RegisterProvider(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestGetProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Get existing provider
	provider, err := manager.GetProvider("test-provider")
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", provider.GetName())

	// Get non-existing provider
	provider, err = manager.GetProvider("non-existing")
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetDefaultProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	// No providers registered
	provider, err := manager.GetDefaultProvider()
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "no providers registered")

	// Register a provider
	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err = manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	// Get default provider (should return the first one)
	provider, err = manager.GetDefaultProvider()
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", provider.GetName())
}

func TestAllow(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	// Set custom Allow function
	mockProvider.SetAllowFunc(func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
		return &types.RateLimitResult{
			Allowed:   true,
			Limit:     limit.Limit,
			Remaining: limit.Limit - 1,
			ResetTime: time.Now().Add(limit.Window),
			Key:       key,
		}, nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	limit := &types.RateLimit{
		Limit:     10,
		Window:    time.Minute,
		Algorithm: types.AlgorithmTokenBucket,
	}

	result, err := manager.Allow(ctx, "test-key", limit)
	assert.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(9), result.Remaining)
	assert.Equal(t, "test-key", result.Key)
}

func TestAllowWithProvider(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider1 := mocks.NewMockProvider()
	mockProvider1.SetName("provider1")

	mockProvider2 := mocks.NewMockProvider()
	mockProvider2.SetName("provider2")

	// Set different behaviors for each provider
	mockProvider1.SetAllowFunc(func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
		return &types.RateLimitResult{
			Allowed:   true,
			Limit:     limit.Limit,
			Remaining: limit.Limit - 1,
			ResetTime: time.Now().Add(limit.Window),
			Key:       key,
		}, nil
	})

	mockProvider2.SetAllowFunc(func(ctx context.Context, key string, limit *types.RateLimit) (*types.RateLimitResult, error) {
		return &types.RateLimitResult{
			Allowed:   false,
			Limit:     limit.Limit,
			Remaining: 0,
			ResetTime: time.Now().Add(limit.Window),
			Key:       key,
		}, nil
	})

	err := manager.RegisterProvider(mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider(mockProvider2)
	require.NoError(t, err)

	ctx := context.Background()
	limit := &types.RateLimit{
		Limit:     10,
		Window:    time.Minute,
		Algorithm: types.AlgorithmTokenBucket,
	}

	// Test with provider1
	result, err := manager.AllowWithProvider(ctx, "provider1", "test-key", limit)
	assert.NoError(t, err)
	assert.True(t, result.Allowed)

	// Test with provider2
	result, err = manager.AllowWithProvider(ctx, "provider2", "test-key", limit)
	assert.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestReset(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	resetCalled := false
	mockProvider.SetResetFunc(func(ctx context.Context, key string) error {
		resetCalled = true
		assert.Equal(t, "test-key", key)
		return nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.Reset(ctx, "test-key")
	assert.NoError(t, err)
	assert.True(t, resetCalled)
}

func TestGetRemaining(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	mockProvider.SetGetRemainingFunc(func(ctx context.Context, key string, limit *types.RateLimit) (int64, error) {
		return limit.Limit - 5, nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	limit := &types.RateLimit{
		Limit:     10,
		Window:    time.Minute,
		Algorithm: types.AlgorithmTokenBucket,
	}

	remaining, err := manager.GetRemaining(ctx, "test-key", limit)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), remaining)
}

func TestGetResetTime(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	expectedTime := time.Now().Add(time.Minute)
	mockProvider.SetGetResetTimeFunc(func(ctx context.Context, key string, limit *types.RateLimit) (time.Time, error) {
		return expectedTime, nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	limit := &types.RateLimit{
		Limit:     10,
		Window:    time.Minute,
		Algorithm: types.AlgorithmTokenBucket,
	}

	resetTime, err := manager.GetResetTime(ctx, "test-key", limit)
	assert.NoError(t, err)
	assert.Equal(t, expectedTime, resetTime)
}

func TestAllowMultiple(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	requests := []*types.RateLimitRequest{
		{
			Key: "key1",
			Limit: &types.RateLimit{
				Limit:     10,
				Window:    time.Minute,
				Algorithm: types.AlgorithmTokenBucket,
			},
		},
		{
			Key: "key2",
			Limit: &types.RateLimit{
				Limit:     5,
				Window:    time.Minute,
				Algorithm: types.AlgorithmTokenBucket,
			},
		},
	}

	results, err := manager.AllowMultiple(ctx, requests)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "key1", results[0].Key)
	assert.Equal(t, "key2", results[1].Key)
}

func TestResetMultiple(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	resetKeys := make([]string, 0)
	mockProvider.SetResetMultipleFunc(func(ctx context.Context, keys []string) error {
		resetKeys = append(resetKeys, keys...)
		return nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	keys := []string{"key1", "key2", "key3"}

	err = manager.ResetMultiple(ctx, keys)
	assert.NoError(t, err)
	assert.Equal(t, keys, resetKeys)
}

func TestGetStats(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	stats, err := manager.GetStats(ctx)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Contains(t, stats, "test-provider")
	assert.Equal(t, int64(100), stats["test-provider"].TotalRequests)
}

func TestGetKeys(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	keys, err := manager.GetKeys(ctx, "test*")
	assert.NoError(t, err)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestListProviders(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	// No providers
	providers := manager.ListProviders()
	assert.Len(t, providers, 0)

	// Add providers
	mockProvider1 := mocks.NewMockProvider()
	mockProvider1.SetName("provider1")

	mockProvider2 := mocks.NewMockProvider()
	mockProvider2.SetName("provider2")

	err := manager.RegisterProvider(mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider(mockProvider2)
	require.NoError(t, err)

	providers = manager.ListProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "provider1")
	assert.Contains(t, providers, "provider2")
}

func TestGetProviderInfo(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	info := manager.GetProviderInfo()
	assert.Len(t, info, 1)
	assert.Contains(t, info, "test-provider")

	providerInfo := info["test-provider"]
	assert.Equal(t, "test-provider", providerInfo.Name)
	assert.True(t, providerInfo.IsConnected)
	assert.Len(t, providerInfo.SupportedFeatures, 6)
}

func TestClose(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewRateLimitManager(nil, logger)

	mockProvider := mocks.NewMockProvider()
	mockProvider.SetName("test-provider")

	disconnectCalled := false
	mockProvider.SetDisconnectFunc(func(ctx context.Context) error {
		disconnectCalled = true
		return nil
	})

	err := manager.RegisterProvider(mockProvider)
	require.NoError(t, err)

	err = manager.Close()
	assert.NoError(t, err)
	assert.True(t, disconnectCalled)
}

// Helper method to access config for testing
func (rlm *gateway.RateLimitManager) GetConfig() *gateway.ManagerConfig {
	return rlm.config
}
