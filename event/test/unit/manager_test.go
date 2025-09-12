package unit

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/event"
	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of EventSourcingProvider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) GetSupportedFeatures() []types.EventSourcingFeature {
	args := m.Called()
	return args.Get(0).([]types.EventSourcingFeature)
}

func (m *MockProvider) GetConnectionInfo() *types.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(*types.ConnectionInfo)
}

func (m *MockProvider) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProvider) AppendEvent(ctx context.Context, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.AppendEventResponse), args.Error(1)
}

func (m *MockProvider) GetEvents(ctx context.Context, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.GetEventsResponse), args.Error(1)
}

func (m *MockProvider) GetEventByID(ctx context.Context, request *types.GetEventByIDRequest) (*types.Event, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.Event), args.Error(1)
}

func (m *MockProvider) GetEventsByStream(ctx context.Context, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.GetEventsResponse), args.Error(1)
}

func (m *MockProvider) CreateStream(ctx context.Context, request *types.CreateStreamRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockProvider) DeleteStream(ctx context.Context, request *types.DeleteStreamRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockProvider) StreamExists(ctx context.Context, request *types.StreamExistsRequest) (bool, error) {
	args := m.Called(ctx, request)
	return args.Bool(0), args.Error(1)
}

func (m *MockProvider) ListStreams(ctx context.Context) ([]types.StreamInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.StreamInfo), args.Error(1)
}

func (m *MockProvider) CreateSnapshot(ctx context.Context, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.CreateSnapshotResponse), args.Error(1)
}

func (m *MockProvider) GetSnapshot(ctx context.Context, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.Snapshot), args.Error(1)
}

func (m *MockProvider) DeleteSnapshot(ctx context.Context, request *types.DeleteSnapshotRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockProvider) AppendEventsBatch(ctx context.Context, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.AppendEventsBatchResponse), args.Error(1)
}

func (m *MockProvider) GetStreamInfo(ctx context.Context, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*types.StreamInfo), args.Error(1)
}

func (m *MockProvider) GetStats(ctx context.Context) (*types.EventSourcingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*types.EventSourcingStats), args.Error(1)
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) Configure(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockProvider) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewEventSourcingManager(t *testing.T) {
	tests := []struct {
		name     string
		config   *gateway.ManagerConfig
		logger   *logrus.Logger
		expected bool
	}{
		{
			name:     "valid config and logger",
			config:   &gateway.ManagerConfig{DefaultProvider: "test"},
			logger:   logrus.New(),
			expected: true,
		},
		{
			name:     "nil config",
			config:   nil,
			logger:   logrus.New(),
			expected: true, // Should use default config
		},
		{
			name:     "nil logger",
			config:   &gateway.ManagerConfig{DefaultProvider: "test"},
			logger:   nil,
			expected: true, // Should create default logger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := gateway.NewEventSourcingManager(tt.config, tt.logger)
			assert.NotNil(t, manager)
			assert.NotNil(t, manager.GetConfig())
			assert.NotNil(t, manager.GetLogger())
		})
	}
}

func TestRegisterProvider(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	t.Run("successful registration", func(t *testing.T) {
		mockProvider.On("GetName").Return("test-provider")
		mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{types.FeatureEventAppend})
		mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{Status: types.StatusDisconnected})
		mockProvider.On("IsConnected").Return(false)

		err := manager.RegisterProvider(mockProvider)
		assert.NoError(t, err)

		providers := manager.ListProviders()
		assert.Contains(t, providers, "test-provider")
	})

	t.Run("nil provider", func(t *testing.T) {
		err := manager.RegisterProvider(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be nil")
	})

	t.Run("empty provider name", func(t *testing.T) {
		mockProvider2 := &MockProvider{}
		mockProvider2.On("GetName").Return("")

		err := manager.RegisterProvider(mockProvider2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot be empty")
	})

	t.Run("duplicate provider", func(t *testing.T) {
		mockProvider3 := &MockProvider{}
		mockProvider3.On("GetName").Return("test-provider")

		err := manager.RegisterProvider(mockProvider3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider test-provider already registered")
	})
}

func TestGetProvider(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{types.FeatureEventAppend})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{Status: types.StatusDisconnected})
	mockProvider.On("IsConnected").Return(false)

	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	t.Run("existing provider", func(t *testing.T) {
		provider, err := manager.GetProvider("test-provider")
		assert.NoError(t, err)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		provider, err := manager.GetProvider("non-existing")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "provider non-existing not found")
	})
}

func TestAppendEvent(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{types.FeatureEventAppend})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{Status: types.StatusConnected})
	mockProvider.On("IsConnected").Return(true)

	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	ctx := context.Background()
	request := &types.AppendEventRequest{
		StreamID:      "test-stream",
		EventType:     "TestEvent",
		EventData:     map[string]interface{}{"test": "data"},
		EventMetadata: map[string]interface{}{"source": "test"},
		AggregateID:   "test-aggregate",
		AggregateType: "Test",
	}

	expectedResponse := &types.AppendEventResponse{
		EventID:   "test-event-id",
		StreamID:  "test-stream",
		Version:   1,
		Timestamp: time.Now(),
	}

	t.Run("successful append", func(t *testing.T) {
		mockProvider.On("AppendEvent", ctx, request).Return(expectedResponse, nil)

		response, err := manager.AppendEvent(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("invalid request", func(t *testing.T) {
		invalidRequest := &types.AppendEventRequest{
			StreamID:  "", // Empty stream ID
			EventType: "TestEvent",
			EventData: map[string]interface{}{"test": "data"},
		}

		response, err := manager.AppendEvent(ctx, invalidRequest)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "stream_id is required")
	})
}

func TestGetEvents(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{types.FeatureEventRetrieval})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{Status: types.StatusConnected})
	mockProvider.On("IsConnected").Return(true)

	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	ctx := context.Background()
	request := &types.GetEventsRequest{
		StreamID: "test-stream",
		Limit:    10,
	}

	expectedEvents := []*types.Event{
		{
			ID:        "event-1",
			StreamID:  "test-stream",
			EventType: "TestEvent",
			EventData: map[string]interface{}{"test": "data1"},
			Version:   1,
			Timestamp: time.Now(),
		},
		{
			ID:        "event-2",
			StreamID:  "test-stream",
			EventType: "TestEvent",
			EventData: map[string]interface{}{"test": "data2"},
			Version:   2,
			Timestamp: time.Now(),
		},
	}

	expectedResponse := &types.GetEventsResponse{
		Events:     expectedEvents,
		TotalCount: 2,
		HasMore:    false,
	}

	t.Run("successful get events", func(t *testing.T) {
		mockProvider.On("GetEvents", ctx, request).Return(expectedResponse, nil)

		response, err := manager.GetEvents(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestCreateStream(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{types.FeatureStreamManagement})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{Status: types.StatusConnected})
	mockProvider.On("IsConnected").Return(true)

	err := manager.RegisterProvider(mockProvider)
	assert.NoError(t, err)

	ctx := context.Background()
	request := &types.CreateStreamRequest{
		StreamID:      "test-stream",
		Name:          "Test Stream",
		AggregateID:   "test-aggregate",
		AggregateType: "Test",
		Metadata:      map[string]interface{}{"description": "Test stream"},
	}

	t.Run("successful create stream", func(t *testing.T) {
		mockProvider.On("CreateStream", ctx, request).Return(nil)

		err := manager.CreateStream(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("invalid request", func(t *testing.T) {
		invalidRequest := &types.CreateStreamRequest{
			StreamID: "", // Empty stream ID
			Name:     "Test Stream",
		}

		err := manager.CreateStream(ctx, invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream_id is required")
	})
}

func TestHealthCheck(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider1 := &MockProvider{}
	mockProvider2 := &MockProvider{}

	mockProvider1.On("GetName").Return("provider-1")
	mockProvider1.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{})
	mockProvider1.On("GetConnectionInfo").Return(&types.ConnectionInfo{})
	mockProvider1.On("IsConnected").Return(true)

	mockProvider2.On("GetName").Return("provider-2")
	mockProvider2.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{})
	mockProvider2.On("GetConnectionInfo").Return(&types.ConnectionInfo{})
	mockProvider2.On("IsConnected").Return(true)

	manager.RegisterProvider(mockProvider1)
	manager.RegisterProvider(mockProvider2)

	ctx := context.Background()

	t.Run("all providers healthy", func(t *testing.T) {
		mockProvider1.On("HealthCheck", ctx).Return(nil)
		mockProvider2.On("HealthCheck", ctx).Return(nil)

		results := manager.HealthCheck(ctx)
		assert.Len(t, results, 2)
		assert.NoError(t, results["provider-1"])
		assert.NoError(t, results["provider-2"])
	})

	t.Run("one provider unhealthy", func(t *testing.T) {
		mockProvider1.On("HealthCheck", ctx).Return(nil)
		mockProvider2.On("HealthCheck", ctx).Return(assert.AnError)

		results := manager.HealthCheck(ctx)
		assert.Len(t, results, 2)
		assert.NoError(t, results["provider-1"])
		assert.Error(t, results["provider-2"])
	})
}

func TestGetStats(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{})
	mockProvider.On("IsConnected").Return(true)

	manager.RegisterProvider(mockProvider)

	ctx := context.Background()
	expectedStats := &types.EventSourcingStats{
		TotalEvents:    100,
		TotalStreams:   10,
		TotalSnapshots: 5,
		EventsByType:   map[string]int64{"TestEvent": 50, "AnotherEvent": 50},
		EventsByStream: map[string]int64{"stream-1": 30, "stream-2": 70},
		StorageSize:    1024 * 1024,
		Uptime:         time.Hour,
		LastUpdate:     time.Now(),
		Provider:       "test-provider",
	}

	t.Run("successful get stats", func(t *testing.T) {
		mockProvider.On("GetStats", ctx).Return(expectedStats, nil)

		stats, err := manager.GetStats(ctx)
		assert.NoError(t, err)
		assert.Len(t, stats, 1)
		assert.Equal(t, expectedStats, stats["test-provider"])
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider.On("GetStats", ctx).Return((*types.EventSourcingStats)(nil), assert.AnError)

		stats, err := manager.GetStats(ctx)
		assert.NoError(t, err) // Manager should handle provider errors gracefully
		assert.Len(t, stats, 0)
	})
}

func TestClose(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return([]types.EventSourcingFeature{})
	mockProvider.On("GetConnectionInfo").Return(&types.ConnectionInfo{})
	mockProvider.On("IsConnected").Return(true)
	mockProvider.On("Close").Return(nil)

	manager.RegisterProvider(mockProvider)

	t.Run("successful close", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)
	})
}

func TestGetProviderInfo(t *testing.T) {
	manager := gateway.NewEventSourcingManager(nil, nil)
	mockProvider := &MockProvider{}

	expectedFeatures := []types.EventSourcingFeature{types.FeatureEventAppend, types.FeatureEventRetrieval}
	expectedConnectionInfo := &types.ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Status:   types.StatusConnected,
	}

	mockProvider.On("GetName").Return("test-provider")
	mockProvider.On("GetSupportedFeatures").Return(expectedFeatures)
	mockProvider.On("GetConnectionInfo").Return(expectedConnectionInfo)
	mockProvider.On("IsConnected").Return(true)

	manager.RegisterProvider(mockProvider)

	info := manager.GetProviderInfo()
	assert.Len(t, info, 1)
	assert.Contains(t, info, "test-provider")

	providerInfo := info["test-provider"]
	assert.Equal(t, "test-provider", providerInfo.Name)
	assert.Equal(t, expectedFeatures, providerInfo.SupportedFeatures)
	assert.Equal(t, expectedConnectionInfo, providerInfo.ConnectionInfo)
	assert.True(t, providerInfo.IsConnected)
}
