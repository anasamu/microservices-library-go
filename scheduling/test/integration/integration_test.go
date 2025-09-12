package integration

import (
	"context"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling"
	"github.com/anasamu/microservices-library-go/scheduling/providers/cron"
	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronProviderIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create cron provider
	cronConfig := &cron.CronConfig{
		Timezone:     "UTC",
		WithSeconds:  false,
		WithLocation: true,
	}

	cronProvider, err := cron.NewCronProvider(cronConfig, logger)
	require.NoError(t, err)

	// Create manager
	managerConfig := &gateway.ManagerConfig{
		DefaultProvider: "cron",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewSchedulingManager(managerConfig, logger)

	// Register provider
	err = manager.RegisterProvider("cron", cronProvider)
	require.NoError(t, err)

	// Connect provider
	err = cronProvider.Connect(context.Background())
	require.NoError(t, err)
	defer cronProvider.Disconnect(context.Background())

	t.Run("ScheduleTask", func(t *testing.T) {
		task := &types.Task{
			ID:          "integration-test-1",
			Name:        "Integration Test Task",
			Description: "A task for integration testing",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *", // Daily at midnight
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://httpbin.org/get",
					Method: "GET",
					Headers: map[string]string{
						"User-Agent": "scheduling-test",
					},
				},
			},
			RetryPolicy: &types.RetryPolicy{
				MaxAttempts: 3,
				Delay:       time.Minute,
				Backoff:     types.BackoffTypeExponential,
			},
			Timeout: time.Minute * 5,
			Metadata: map[string]string{
				"test": "integration",
			},
		}

		result, err := manager.ScheduleTask(context.Background(), task, "cron")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.ID, result.TaskID)
		assert.Equal(t, types.TaskStatusScheduled, result.Status)
	})

	t.Run("GetTask", func(t *testing.T) {
		retrievedTask, err := manager.GetTask(context.Background(), "integration-test-1", "cron")
		require.NoError(t, err)
		assert.NotNil(t, retrievedTask)
		assert.Equal(t, "integration-test-1", retrievedTask.ID)
		assert.Equal(t, "Integration Test Task", retrievedTask.Name)
	})

	t.Run("ListTasks", func(t *testing.T) {
		filter := &types.TaskFilter{
			Status: []types.TaskStatus{types.TaskStatusScheduled},
			Limit:  10,
		}

		tasks, err := manager.ListTasks(context.Background(), filter, "cron")
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)

		// Find our test task
		var found bool
		for _, task := range tasks {
			if task.ID == "integration-test-1" {
				found = true
				break
			}
		}
		assert.True(t, found, "Test task should be found in the list")
	})

	t.Run("UpdateTask", func(t *testing.T) {
		task, err := manager.GetTask(context.Background(), "integration-test-1", "cron")
		require.NoError(t, err)

		// Update task description
		task.Description = "Updated integration test task"
		task.Metadata["updated"] = "true"

		err = manager.UpdateTask(context.Background(), task, "cron")
		require.NoError(t, err)

		// Verify update
		updatedTask, err := manager.GetTask(context.Background(), "integration-test-1", "cron")
		require.NoError(t, err)
		assert.Equal(t, "Updated integration test task", updatedTask.Description)
		assert.Equal(t, "true", updatedTask.Metadata["updated"])
	})

	t.Run("ScheduleMultiple", func(t *testing.T) {
		tasks := []*types.Task{
			{
				ID:   "integration-batch-1",
				Name: "Batch Task 1",
				Schedule: &types.Schedule{
					Type:     types.ScheduleTypeCron,
					CronExpr: "0 1 * * *",
				},
				Handler: &types.TaskHandler{
					Type: types.HandlerTypeHTTP,
					HTTP: &types.HTTPHandler{
						URL:    "http://httpbin.org/get",
						Method: "GET",
					},
				},
			},
			{
				ID:   "integration-batch-2",
				Name: "Batch Task 2",
				Schedule: &types.Schedule{
					Type:     types.ScheduleTypeCron,
					CronExpr: "0 2 * * *",
				},
				Handler: &types.TaskHandler{
					Type: types.HandlerTypeHTTP,
					HTTP: &types.HTTPHandler{
						URL:    "http://httpbin.org/get",
						Method: "GET",
					},
				},
			},
		}

		results, err := manager.ScheduleMultiple(context.Background(), tasks, "cron")
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("CancelTask", func(t *testing.T) {
		err := manager.CancelTask(context.Background(), "integration-test-1", "cron")
		require.NoError(t, err)

		// Verify task is cancelled
		task, err := manager.GetTask(context.Background(), "integration-test-1", "cron")
		require.NoError(t, err)
		assert.Equal(t, types.TaskStatusCancelled, task.Status)
	})

	t.Run("CancelMultiple", func(t *testing.T) {
		taskIDs := []string{"integration-batch-1", "integration-batch-2"}
		err := manager.CancelMultiple(context.Background(), taskIDs, "cron")
		require.NoError(t, err)

		// Verify tasks are cancelled
		for _, taskID := range taskIDs {
			task, err := manager.GetTask(context.Background(), taskID, "cron")
			require.NoError(t, err)
			assert.Equal(t, types.TaskStatusCancelled, task.Status)
		}
	})

	t.Run("GetHealth", func(t *testing.T) {
		health, err := manager.GetHealth(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, health)
		assert.Contains(t, health, "cron")
		assert.Equal(t, types.StatusHealthy, health["cron"].Status)
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics, err := manager.GetMetrics(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "cron")
		assert.NotNil(t, metrics["cron"])
	})
}

func TestManagerWithMultipleProviders(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager
	managerConfig := &gateway.ManagerConfig{
		DefaultProvider: "cron",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewSchedulingManager(managerConfig, logger)

	// Create and register cron provider
	cronConfig := &cron.CronConfig{
		Timezone:     "UTC",
		WithSeconds:  false,
		WithLocation: true,
	}

	cronProvider, err := cron.NewCronProvider(cronConfig, logger)
	require.NoError(t, err)

	err = manager.RegisterProvider("cron", cronProvider)
	require.NoError(t, err)

	// Connect all providers
	err = manager.ConnectAll(context.Background())
	require.NoError(t, err)
	defer manager.DisconnectAll(context.Background())

	t.Run("ProviderManagement", func(t *testing.T) {
		// List providers
		providers := manager.ListProviders()
		assert.Contains(t, providers, "cron")

		// Get provider
		provider, err := manager.GetProvider("cron")
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "cron", provider.GetName())

		// Test non-existing provider
		_, err = manager.GetProvider("nonexistent")
		assert.Error(t, err)
	})

	t.Run("FallbackBehavior", func(t *testing.T) {
		task := &types.Task{
			ID:   "fallback-test",
			Name: "Fallback Test Task",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://httpbin.org/get",
					Method: "GET",
				},
			},
		}

		// Schedule with non-existing provider (should use fallback)
		result, err := manager.ScheduleTask(context.Background(), task, "nonexistent")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.ID, result.TaskID)
	})

	t.Run("HealthAndMetrics", func(t *testing.T) {
		// Get health for all providers
		health, err := manager.GetHealth(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, health)
		assert.Contains(t, health, "cron")

		// Get metrics for all providers
		metrics, err := manager.GetMetrics(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "cron")
	})
}

func TestTaskLifecycle(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create cron provider
	cronConfig := &cron.CronConfig{
		Timezone:     "UTC",
		WithSeconds:  false,
		WithLocation: true,
	}

	cronProvider, err := cron.NewCronProvider(cronConfig, logger)
	require.NoError(t, err)

	// Create manager
	manager := gateway.NewSchedulingManager(nil, logger)
	err = manager.RegisterProvider("cron", cronProvider)
	require.NoError(t, err)

	// Connect provider
	err = cronProvider.Connect(context.Background())
	require.NoError(t, err)
	defer cronProvider.Disconnect(context.Background())

	t.Run("CompleteTaskLifecycle", func(t *testing.T) {
		// 1. Create task
		task := &types.Task{
			ID:          "lifecycle-test",
			Name:        "Lifecycle Test Task",
			Description: "A task to test complete lifecycle",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://httpbin.org/get",
					Method: "GET",
				},
			},
			RetryPolicy: &types.RetryPolicy{
				MaxAttempts: 3,
				Delay:       time.Minute,
				Backoff:     types.BackoffTypeExponential,
			},
			Timeout: time.Minute * 5,
			Tags:    []string{"test", "lifecycle"},
			Metadata: map[string]string{
				"environment": "test",
			},
		}

		// 2. Schedule task
		result, err := manager.ScheduleTask(context.Background(), task, "cron")
		require.NoError(t, err)
		assert.Equal(t, types.TaskStatusScheduled, result.Status)

		// 3. Verify task exists
		retrievedTask, err := manager.GetTask(context.Background(), task.ID, "cron")
		require.NoError(t, err)
		assert.Equal(t, task.ID, retrievedTask.ID)
		assert.Equal(t, task.Name, retrievedTask.Name)
		assert.Equal(t, types.TaskStatusScheduled, retrievedTask.Status)

		// 4. Update task
		retrievedTask.Description = "Updated lifecycle test task"
		retrievedTask.Metadata["updated"] = "true"
		err = manager.UpdateTask(context.Background(), retrievedTask, "cron")
		require.NoError(t, err)

		// 5. Verify update
		updatedTask, err := manager.GetTask(context.Background(), task.ID, "cron")
		require.NoError(t, err)
		assert.Equal(t, "Updated lifecycle test task", updatedTask.Description)
		assert.Equal(t, "true", updatedTask.Metadata["updated"])

		// 6. List tasks with filter
		filter := &types.TaskFilter{
			Tags:  []string{"test"},
			Limit: 10,
		}
		tasks, err := manager.ListTasks(context.Background(), filter, "cron")
		require.NoError(t, err)
		assert.NotEmpty(t, tasks)

		// 7. Cancel task
		err = manager.CancelTask(context.Background(), task.ID, "cron")
		require.NoError(t, err)

		// 8. Verify cancellation
		cancelledTask, err := manager.GetTask(context.Background(), task.ID, "cron")
		require.NoError(t, err)
		assert.Equal(t, types.TaskStatusCancelled, cancelledTask.Status)
	})
}
