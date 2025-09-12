package unit

import (
	"context"
	"testing"

	"github.com/anasamu/microservices-library-go/scheduling"
	"github.com/anasamu/microservices-library-go/scheduling/test/mocks"
	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchedulingManager(t *testing.T) {
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
			expected: true, // Should create default config
		},
		{
			name:     "nil logger",
			config:   &gateway.ManagerConfig{DefaultProvider: "test"},
			logger:   nil,
			expected: true, // Should create default logger
		},
		{
			name:     "both nil",
			config:   nil,
			logger:   nil,
			expected: true, // Should create defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := gateway.NewSchedulingManager(tt.config, tt.logger)
			assert.NotNil(t, manager)
		})
	}
}

func TestRegisterProvider(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	tests := []struct {
		name         string
		providerName string
		provider     *mocks.MockSchedulingProvider
		expectErr    bool
	}{
		{
			name:         "valid provider",
			providerName: "test",
			provider:     mockProvider,
			expectErr:    false,
		},
		{
			name:         "empty provider name",
			providerName: "",
			provider:     mockProvider,
			expectErr:    true,
		},
		{
			name:         "nil provider",
			providerName: "test",
			provider:     nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RegisterProvider(tt.providerName, tt.provider)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	tests := []struct {
		name         string
		providerName string
		expectErr    bool
	}{
		{
			name:         "existing provider",
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing provider",
			providerName: "nonexistent",
			expectErr:    true,
		},
		{
			name:         "empty provider name (should use default)",
			providerName: "",
			expectErr:    true, // No default provider set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := manager.GetProvider(tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, "test", provider.GetName())
			}
		})
	}
}

func TestListProviders(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)

	// Initially should be empty
	providers := manager.ListProviders()
	assert.Empty(t, providers)

	// Add providers
	mockProvider1 := mocks.NewMockSchedulingProvider("provider1")
	mockProvider2 := mocks.NewMockSchedulingProvider("provider2")

	err := manager.RegisterProvider("provider1", mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider("provider2", mockProvider2)
	require.NoError(t, err)

	providers = manager.ListProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "provider1")
	assert.Contains(t, providers, "provider2")
}

func TestScheduleTask(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	task := &types.Task{
		ID:   "test-task",
		Name: "Test Task",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 0 * * *",
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeHTTP,
			HTTP: &types.HTTPHandler{
				URL:    "http://example.com",
				Method: "GET",
			},
		},
	}

	tests := []struct {
		name         string
		task         *types.Task
		providerName string
		expectErr    bool
	}{
		{
			name:         "valid task",
			task:         task,
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing provider",
			task:         task,
			providerName: "nonexistent",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.ScheduleTask(context.Background(), tt.task, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.task.ID, result.TaskID)
			}
		})
	}
}

func TestCancelTask(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	// Schedule a task first
	task := &types.Task{
		ID:   "test-task",
		Name: "Test Task",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 0 * * *",
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeHTTP,
			HTTP: &types.HTTPHandler{
				URL:    "http://example.com",
				Method: "GET",
			},
		},
	}

	_, err = manager.ScheduleTask(context.Background(), task, "test")
	require.NoError(t, err)

	tests := []struct {
		name         string
		taskID       string
		providerName string
		expectErr    bool
	}{
		{
			name:         "valid task ID",
			taskID:       "test-task",
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing task",
			taskID:       "nonexistent",
			providerName: "test",
			expectErr:    true,
		},
		{
			name:         "non-existing provider",
			taskID:       "test-task",
			providerName: "nonexistent",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CancelTask(context.Background(), tt.taskID, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	// Schedule a task first
	task := &types.Task{
		ID:   "test-task",
		Name: "Test Task",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 0 * * *",
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeHTTP,
			HTTP: &types.HTTPHandler{
				URL:    "http://example.com",
				Method: "GET",
			},
		},
	}

	_, err = manager.ScheduleTask(context.Background(), task, "test")
	require.NoError(t, err)

	tests := []struct {
		name         string
		taskID       string
		providerName string
		expectErr    bool
	}{
		{
			name:         "valid task ID",
			taskID:       "test-task",
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing task",
			taskID:       "nonexistent",
			providerName: "test",
			expectErr:    true,
		},
		{
			name:         "non-existing provider",
			taskID:       "test-task",
			providerName: "nonexistent",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedTask, err := manager.GetTask(context.Background(), tt.taskID, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, retrievedTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedTask)
				assert.Equal(t, tt.taskID, retrievedTask.ID)
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	// Schedule some tasks
	tasks := []*types.Task{
		{
			ID:   "task1",
			Name: "Task 1",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "GET",
				},
			},
		},
		{
			ID:   "task2",
			Name: "Task 2",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 1 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "POST",
				},
			},
		},
	}

	for _, task := range tasks {
		_, err := manager.ScheduleTask(context.Background(), task, "test")
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        *types.TaskFilter
		providerName  string
		expectErr     bool
		expectedCount int
	}{
		{
			name:          "no filter",
			filter:        nil,
			providerName:  "test",
			expectErr:     false,
			expectedCount: 2,
		},
		{
			name:          "with limit",
			filter:        &types.TaskFilter{Limit: 1},
			providerName:  "test",
			expectErr:     false,
			expectedCount: 1,
		},
		{
			name:          "non-existing provider",
			filter:        nil,
			providerName:  "nonexistent",
			expectErr:     true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedTasks, err := manager.ListTasks(context.Background(), tt.filter, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, retrievedTasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, retrievedTasks, tt.expectedCount)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	// Schedule a task first
	task := &types.Task{
		ID:   "test-task",
		Name: "Test Task",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 0 * * *",
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeHTTP,
			HTTP: &types.HTTPHandler{
				URL:    "http://example.com",
				Method: "GET",
			},
		},
	}

	_, err = manager.ScheduleTask(context.Background(), task, "test")
	require.NoError(t, err)

	// Update the task
	task.Name = "Updated Test Task"
	task.Description = "Updated description"

	tests := []struct {
		name         string
		task         *types.Task
		providerName string
		expectErr    bool
	}{
		{
			name:         "valid update",
			task:         task,
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing task",
			task:         &types.Task{ID: "nonexistent", Name: "Test"},
			providerName: "test",
			expectErr:    true,
		},
		{
			name:         "non-existing provider",
			task:         task,
			providerName: "nonexistent",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateTask(context.Background(), tt.task, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScheduleMultiple(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	tasks := []*types.Task{
		{
			ID:   "task1",
			Name: "Task 1",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "GET",
				},
			},
		},
		{
			ID:   "task2",
			Name: "Task 2",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 1 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "POST",
				},
			},
		},
	}

	tests := []struct {
		name          string
		tasks         []*types.Task
		providerName  string
		expectErr     bool
		expectedCount int
	}{
		{
			name:          "valid tasks",
			tasks:         tasks,
			providerName:  "test",
			expectErr:     false,
			expectedCount: 2,
		},
		{
			name:          "non-existing provider",
			tasks:         tasks,
			providerName:  "nonexistent",
			expectErr:     true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := manager.ScheduleMultiple(context.Background(), tt.tasks, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectedCount)
			}
		})
	}
}

func TestCancelMultiple(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	// Schedule some tasks first
	tasks := []*types.Task{
		{
			ID:   "task1",
			Name: "Task 1",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "GET",
				},
			},
		},
		{
			ID:   "task2",
			Name: "Task 2",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 1 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://example.com",
					Method: "POST",
				},
			},
		},
	}

	for _, task := range tasks {
		_, err := manager.ScheduleTask(context.Background(), task, "test")
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		taskIDs      []string
		providerName string
		expectErr    bool
	}{
		{
			name:         "valid task IDs",
			taskIDs:      []string{"task1", "task2"},
			providerName: "test",
			expectErr:    false,
		},
		{
			name:         "non-existing provider",
			taskIDs:      []string{"task1", "task2"},
			providerName: "nonexistent",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CancelMultiple(context.Background(), tt.taskIDs, tt.providerName)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetHealth(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	health, err := manager.GetHealth(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.Contains(t, health, "test")
	assert.Equal(t, types.StatusHealthy, health["test"].Status)
}

func TestGetMetrics(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider := mocks.NewMockSchedulingProvider("test")

	err := manager.RegisterProvider("test", mockProvider)
	require.NoError(t, err)

	// Connect the provider
	err = mockProvider.Connect(context.Background())
	require.NoError(t, err)

	metrics, err := manager.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "test")
	assert.NotNil(t, metrics["test"])
}

func TestConnectAll(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider1 := mocks.NewMockSchedulingProvider("provider1")
	mockProvider2 := mocks.NewMockSchedulingProvider("provider2")

	err := manager.RegisterProvider("provider1", mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider("provider2", mockProvider2)
	require.NoError(t, err)

	err = manager.ConnectAll(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockProvider1.IsConnected())
	assert.True(t, mockProvider2.IsConnected())
}

func TestDisconnectAll(t *testing.T) {
	manager := gateway.NewSchedulingManager(nil, nil)
	mockProvider1 := mocks.NewMockSchedulingProvider("provider1")
	mockProvider2 := mocks.NewMockSchedulingProvider("provider2")

	err := manager.RegisterProvider("provider1", mockProvider1)
	require.NoError(t, err)

	err = manager.RegisterProvider("provider2", mockProvider2)
	require.NoError(t, err)

	// Connect first
	err = manager.ConnectAll(context.Background())
	require.NoError(t, err)

	// Then disconnect
	err = manager.DisconnectAll(context.Background())
	assert.NoError(t, err)
	assert.False(t, mockProvider1.IsConnected())
	assert.False(t, mockProvider2.IsConnected())
}
