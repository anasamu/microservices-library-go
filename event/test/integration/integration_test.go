package integration

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/event/gateway"
	"github.com/anasamu/microservices-library-go/event/providers/postgresql"
	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require actual database connections
// These tests are skipped by default and should be run manually with proper setup

func TestPostgreSQLIntegration(t *testing.T) {
	// Skip if no database is available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if PostgreSQL is available
	config := &postgresql.PostgreSQLConfig{
		Host:            "localhost",
		Port:            5432,
		Database:        "eventsourcing_test",
		Username:        "postgres",
		Password:        "password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	provider, err := postgresql.NewPostgreSQLProvider(config, logger)
	require.NoError(t, err)

	// Try to connect
	ctx := context.Background()
	err = provider.Connect(ctx)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	defer provider.Disconnect(ctx)

	// Test basic operations
	t.Run("CreateStream", func(t *testing.T) {
		request := &types.CreateStreamRequest{
			StreamID:      "test-stream-integration",
			Name:          "Test Stream",
			AggregateID:   "test-aggregate",
			AggregateType: "Test",
			Metadata:      map[string]interface{}{"test": true},
		}

		err := provider.CreateStream(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("AppendEvent", func(t *testing.T) {
		request := &types.AppendEventRequest{
			StreamID:      "test-stream-integration",
			EventType:     "TestEvent",
			EventData:     map[string]interface{}{"message": "Hello World"},
			EventMetadata: map[string]interface{}{"source": "integration_test"},
			AggregateID:   "test-aggregate",
			AggregateType: "Test",
		}

		response, err := provider.AppendEvent(ctx, request)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.EventID)
		assert.Equal(t, "test-stream-integration", response.StreamID)
		assert.Equal(t, int64(1), response.Version)
	})

	t.Run("GetEvents", func(t *testing.T) {
		request := &types.GetEventsRequest{
			StreamID: "test-stream-integration",
			Limit:    10,
		}

		response, err := provider.GetEvents(ctx, request)
		assert.NoError(t, err)
		assert.Len(t, response.Events, 1)
		assert.Equal(t, "TestEvent", response.Events[0].EventType)
		assert.Equal(t, "Hello World", response.Events[0].EventData["message"])
	})

	t.Run("GetEventsByStream", func(t *testing.T) {
		request := &types.GetEventsByStreamRequest{
			StreamID: "test-stream-integration",
			Limit:    10,
		}

		response, err := provider.GetEventsByStream(ctx, request)
		assert.NoError(t, err)
		assert.Len(t, response.Events, 1)
	})

	t.Run("StreamExists", func(t *testing.T) {
		request := &types.StreamExistsRequest{
			StreamID: "test-stream-integration",
		}

		exists, err := provider.StreamExists(ctx, request)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("GetStreamInfo", func(t *testing.T) {
		request := &types.GetStreamInfoRequest{
			StreamID: "test-stream-integration",
		}

		info, err := provider.GetStreamInfo(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, "test-stream-integration", info.ID)
		assert.Equal(t, int64(1), info.EventCount)
	})

	t.Run("CreateSnapshot", func(t *testing.T) {
		request := &types.CreateSnapshotRequest{
			StreamID:      "test-stream-integration",
			AggregateID:   "test-aggregate",
			AggregateType: "Test",
			Version:       1,
			Data: map[string]interface{}{
				"state": "active",
				"count": 1,
			},
			Metadata: map[string]interface{}{"snapshot_type": "test"},
		}

		response, err := provider.CreateSnapshot(ctx, request)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.SnapshotID)
		assert.Equal(t, int64(1), response.Version)
	})

	t.Run("GetSnapshot", func(t *testing.T) {
		request := &types.GetSnapshotRequest{
			StreamID: "test-stream-integration",
		}

		snapshot, err := provider.GetSnapshot(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, "test-stream-integration", snapshot.StreamID)
		assert.Equal(t, "active", snapshot.Data["state"])
		assert.Equal(t, float64(1), snapshot.Data["count"]) // JSON numbers are float64
	})

	t.Run("AppendEventsBatch", func(t *testing.T) {
		events := []*types.Event{
			{
				EventType:     "BatchEvent1",
				EventData:     map[string]interface{}{"message": "Batch 1"},
				EventMetadata: map[string]interface{}{"batch": true},
				AggregateID:   "test-aggregate",
				AggregateType: "Test",
			},
			{
				EventType:     "BatchEvent2",
				EventData:     map[string]interface{}{"message": "Batch 2"},
				EventMetadata: map[string]interface{}{"batch": true},
				AggregateID:   "test-aggregate",
				AggregateType: "Test",
			},
		}

		request := &types.AppendEventsBatchRequest{
			StreamID: "test-stream-integration",
			Events:   events,
		}

		response, err := provider.AppendEventsBatch(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.AppendedCount)
		assert.Equal(t, 0, response.FailedCount)
		assert.Equal(t, int64(2), response.StartVersion)
		assert.Equal(t, int64(4), response.EndVersion)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := provider.GetStats(ctx)
		assert.NoError(t, err)
		assert.Greater(t, stats.TotalEvents, int64(0))
		assert.Greater(t, stats.TotalStreams, int64(0))
		assert.Greater(t, stats.TotalSnapshots, int64(0))
		assert.Equal(t, "postgresql", stats.Provider)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		err := provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	// Cleanup
	t.Run("DeleteStream", func(t *testing.T) {
		request := &types.DeleteStreamRequest{
			StreamID: "test-stream-integration",
		}

		err := provider.DeleteStream(ctx, request)
		assert.NoError(t, err)
	})
}

func TestManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create manager
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &gateway.ManagerConfig{
		DefaultProvider:   "postgresql",
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		Timeout:           30 * time.Second,
		MaxEventSize:      1024 * 1024,
		MaxBatchSize:      100,
		SnapshotThreshold: 1000,
		RetentionPeriod:   365 * 24 * time.Hour,
	}

	manager := gateway.NewEventSourcingManager(config, logger)

	// Create and register PostgreSQL provider
	pgConfig := &postgresql.PostgreSQLConfig{
		Host:            "localhost",
		Port:            5432,
		Database:        "eventsourcing_test",
		Username:        "postgres",
		Password:        "password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	pgProvider, err := postgresql.NewPostgreSQLProvider(pgConfig, logger)
	require.NoError(t, err)

	err = manager.RegisterProvider(pgProvider)
	require.NoError(t, err)

	// Try to connect
	ctx := context.Background()
	err = pgProvider.Connect(ctx)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	defer manager.Close()

	t.Run("ManagerOperations", func(t *testing.T) {
		// Create stream
		streamRequest := &types.CreateStreamRequest{
			StreamID:      "manager-test-stream",
			Name:          "Manager Test Stream",
			AggregateID:   "manager-test-aggregate",
			AggregateType: "ManagerTest",
		}

		err := manager.CreateStream(ctx, streamRequest)
		assert.NoError(t, err)

		// Append event
		eventRequest := &types.AppendEventRequest{
			StreamID:      "manager-test-stream",
			EventType:     "ManagerTestEvent",
			EventData:     map[string]interface{}{"test": "manager"},
			EventMetadata: map[string]interface{}{"source": "manager_test"},
			AggregateID:   "manager-test-aggregate",
			AggregateType: "ManagerTest",
		}

		response, err := manager.AppendEvent(ctx, eventRequest)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.EventID)

		// Get events
		getEventsRequest := &types.GetEventsRequest{
			StreamID: "manager-test-stream",
			Limit:    10,
		}

		eventsResponse, err := manager.GetEvents(ctx, getEventsRequest)
		assert.NoError(t, err)
		assert.Len(t, eventsResponse.Events, 1)
		assert.Equal(t, "ManagerTestEvent", eventsResponse.Events[0].EventType)

		// Get stream info
		streamInfoRequest := &types.GetStreamInfoRequest{
			StreamID: "manager-test-stream",
		}

		streamInfo, err := manager.GetStreamInfo(ctx, streamInfoRequest)
		assert.NoError(t, err)
		assert.Equal(t, "manager-test-stream", streamInfo.ID)
		assert.Equal(t, int64(1), streamInfo.EventCount)

		// Health check
		healthResults := manager.HealthCheck(ctx)
		assert.Len(t, healthResults, 1)
		assert.NoError(t, healthResults["postgresql"])

		// Get stats
		stats, err := manager.GetStats(ctx)
		assert.NoError(t, err)
		assert.Len(t, stats, 1)
		assert.Contains(t, stats, "postgresql")

		// Cleanup
		deleteRequest := &types.DeleteStreamRequest{
			StreamID: "manager-test-stream",
		}

		err = manager.DeleteStream(ctx, deleteRequest)
		assert.NoError(t, err)
	})
}

func TestEventSourcingAggregateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test demonstrates a complete event sourcing aggregate workflow
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &gateway.ManagerConfig{
		DefaultProvider: "postgresql",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
	}

	manager := gateway.NewEventSourcingManager(config, logger)

	pgConfig := &postgresql.PostgreSQLConfig{
		Host:            "localhost",
		Port:            5432,
		Database:        "eventsourcing_test",
		Username:        "postgres",
		Password:        "password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	pgProvider, err := postgresql.NewPostgreSQLProvider(pgConfig, logger)
	require.NoError(t, err)

	err = manager.RegisterProvider(pgProvider)
	require.NoError(t, err)

	ctx := context.Background()
	err = pgProvider.Connect(ctx)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	defer manager.Close()

	// Create a simple aggregate
	aggregate := &TestAggregate{
		ID:      "aggregate-test-123",
		Version: 0,
		manager: manager,
	}

	t.Run("AggregateWorkflow", func(t *testing.T) {
		// Create aggregate
		err := aggregate.Create(ctx, "Test User", "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "Test User", aggregate.Name)
		assert.Equal(t, "test@example.com", aggregate.Email)
		assert.Equal(t, int64(1), aggregate.Version)

		// Update aggregate
		err = aggregate.Update(ctx, "Updated User", "updated@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "Updated User", aggregate.Name)
		assert.Equal(t, "updated@example.com", aggregate.Email)
		assert.Equal(t, int64(2), aggregate.Version)

		// Load from events (simulate aggregate reconstruction)
		newAggregate := &TestAggregate{
			ID:      "aggregate-test-123",
			Version: 0,
			manager: manager,
		}

		err = newAggregate.LoadFromEvents(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Updated User", newAggregate.Name)
		assert.Equal(t, "updated@example.com", newAggregate.Email)
		assert.Equal(t, int64(2), newAggregate.Version)

		// Cleanup
		deleteRequest := &types.DeleteStreamRequest{
			StreamID: "aggregate-test-123",
		}

		err = manager.DeleteStream(ctx, deleteRequest)
		assert.NoError(t, err)
	})
}

// TestAggregate is a simple aggregate for testing
type TestAggregate struct {
	ID      string
	Name    string
	Email   string
	Version int64
	manager *gateway.EventSourcingManager
}

func (a *TestAggregate) Create(ctx context.Context, name, email string) error {
	eventRequest := &types.AppendEventRequest{
		StreamID:        a.ID,
		EventType:       "UserCreated",
		EventData:       map[string]interface{}{"name": name, "email": email},
		EventMetadata:   map[string]interface{}{"source": "aggregate"},
		AggregateID:     a.ID,
		AggregateType:   "User",
		ExpectedVersion: &a.Version,
	}

	response, err := a.manager.AppendEvent(ctx, eventRequest)
	if err != nil {
		return err
	}

	// Update aggregate state
	a.Name = name
	a.Email = email
	a.Version = response.Version

	return nil
}

func (a *TestAggregate) Update(ctx context.Context, name, email string) error {
	eventRequest := &types.AppendEventRequest{
		StreamID:        a.ID,
		EventType:       "UserUpdated",
		EventData:       map[string]interface{}{"name": name, "email": email},
		EventMetadata:   map[string]interface{}{"source": "aggregate"},
		AggregateID:     a.ID,
		AggregateType:   "User",
		ExpectedVersion: &a.Version,
	}

	response, err := a.manager.AppendEvent(ctx, eventRequest)
	if err != nil {
		return err
	}

	// Update aggregate state
	a.Name = name
	a.Email = email
	a.Version = response.Version

	return nil
}

func (a *TestAggregate) LoadFromEvents(ctx context.Context) error {
	request := &types.GetEventsByStreamRequest{
		StreamID: a.ID,
		Limit:    1000,
	}

	response, err := a.manager.GetEventsByStream(ctx, request)
	if err != nil {
		return err
	}

	// Replay events to rebuild aggregate state
	for _, event := range response.Events {
		a.Version = event.Version

		switch event.EventType {
		case "UserCreated":
			if name, ok := event.EventData["name"].(string); ok {
				a.Name = name
			}
			if email, ok := event.EventData["email"].(string); ok {
				a.Email = email
			}
		case "UserUpdated":
			if name, ok := event.EventData["name"].(string); ok {
				a.Name = name
			}
			if email, ok := event.EventData["email"].(string); ok {
				a.Email = email
			}
		}
	}

	return nil
}
