package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling/gateway"
	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/sirupsen/logrus"
)

// Example demonstrates how to use the scheduling manager
func Example() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager configuration
	config := &gateway.ManagerConfig{
		DefaultProvider: "cron",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
		Metadata: map[string]string{
			"environment": "development",
			"version":     "1.0.0",
		},
	}

	// Create scheduling manager
	manager := gateway.NewSchedulingManager(config, logger)

	// Example 1: Schedule a simple task
	ctx := context.Background()

	// Create a simple task
	task := &types.Task{
		ID:          "task-001",
		Name:        "Daily Report Generation",
		Description: "Generate daily sales report",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 9 * * *", // Every day at 9 AM
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeHTTP,
			HTTP: &types.HTTPHandler{
				URL:    "http://api.example.com/reports/daily",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer token123",
				},
				Body: `{"report_type": "daily", "date": "{{.Date}}"}`,
			},
		},
		RetryPolicy: &types.RetryPolicy{
			MaxAttempts: 3,
			Delay:       time.Minute,
			Backoff:     types.BackoffTypeExponential,
		},
		Timeout: time.Minute * 5,
		Metadata: map[string]string{
			"department": "sales",
			"priority":   "high",
		},
	}

	// Schedule the task
	result, err := manager.ScheduleTask(ctx, task, "cron")
	if err != nil {
		log.Printf("Failed to schedule task: %v", err)
		return
	}

	fmt.Printf("Task scheduled successfully: %s\n", result.TaskID)

	// Example 2: Schedule a recurring task with complex schedule
	recurringTask := &types.Task{
		ID:          "task-002",
		Name:        "Data Backup",
		Description: "Backup database every 6 hours",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "0 */6 * * *", // Every 6 hours
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeCommand,
			Command: &types.CommandHandler{
				Command: "pg_dump",
				Args:    []string{"-h", "localhost", "-U", "postgres", "mydb"},
				Env: map[string]string{
					"PGPASSWORD": "secret",
				},
			},
		},
		RetryPolicy: &types.RetryPolicy{
			MaxAttempts: 2,
			Delay:       time.Minute * 2,
			Backoff:     types.BackoffTypeLinear,
		},
		Timeout: time.Minute * 30,
		Metadata: map[string]string{
			"database":    "production",
			"backup_type": "full",
		},
	}

	result2, err := manager.ScheduleTask(ctx, recurringTask, "cron")
	if err != nil {
		log.Printf("Failed to schedule recurring task: %v", err)
		return
	}

	fmt.Printf("Recurring task scheduled successfully: %s\n", result2.TaskID)

	// Example 3: Schedule a one-time task
	oneTimeTask := &types.Task{
		ID:          "task-003",
		Name:        "System Maintenance",
		Description: "Run system maintenance at specific time",
		Schedule: &types.Schedule{
			Type:      types.ScheduleTypeOnce,
			StartTime: time.Now().Add(time.Hour), // Run in 1 hour
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeFunction,
			Function: &types.FunctionHandler{
				FunctionName: "maintenance",
				Parameters: map[string]interface{}{
					"cleanup_temp": true,
					"optimize_db":  true,
				},
			},
		},
		RetryPolicy: &types.RetryPolicy{
			MaxAttempts: 1,
			Delay:       time.Second * 30,
			Backoff:     types.BackoffTypeFixed,
		},
		Timeout: time.Minute * 10,
	}

	result3, err := manager.ScheduleTask(ctx, oneTimeTask, "cron")
	if err != nil {
		log.Printf("Failed to schedule one-time task: %v", err)
		return
	}

	fmt.Printf("One-time task scheduled successfully: %s\n", result3.TaskID)

	// Example 4: List tasks
	filter := &types.TaskFilter{
		Status: []types.TaskStatus{types.TaskStatusScheduled, types.TaskStatusRunning},
		Limit:  10,
	}

	tasks, err := manager.ListTasks(ctx, filter, "cron")
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		return
	}

	fmt.Printf("Found %d tasks:\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("- %s: %s (Status: %s)\n", task.ID, task.Name, task.Status)
	}

	// Example 5: Get specific task
	retrievedTask, err := manager.GetTask(ctx, "task-001", "cron")
	if err != nil {
		log.Printf("Failed to get task: %v", err)
		return
	}

	fmt.Printf("Retrieved task: %s - %s\n", retrievedTask.ID, retrievedTask.Name)

	// Example 6: Update task
	retrievedTask.Description = "Generate daily sales report with enhanced analytics"
	err = manager.UpdateTask(ctx, retrievedTask, "cron")
	if err != nil {
		log.Printf("Failed to update task: %v", err)
		return
	}

	fmt.Printf("Task updated successfully: %s\n", retrievedTask.ID)

	// Example 7: Schedule multiple tasks
	batchTasks := []*types.Task{
		{
			ID:   "batch-001",
			Name: "Batch Task 1",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://api.example.com/batch1",
					Method: "POST",
				},
			},
		},
		{
			ID:   "batch-002",
			Name: "Batch Task 2",
			Schedule: &types.Schedule{
				Type:     types.ScheduleTypeCron,
				CronExpr: "0 1 * * *",
			},
			Handler: &types.TaskHandler{
				Type: types.HandlerTypeHTTP,
				HTTP: &types.HTTPHandler{
					URL:    "http://api.example.com/batch2",
					Method: "POST",
				},
			},
		},
	}

	results, err := manager.ScheduleMultiple(ctx, batchTasks, "cron")
	if err != nil {
		log.Printf("Failed to schedule multiple tasks: %v", err)
		return
	}

	fmt.Printf("Scheduled %d batch tasks successfully\n", len(results))

	// Example 8: Cancel task
	err = manager.CancelTask(ctx, "task-003", "cron")
	if err != nil {
		log.Printf("Failed to cancel task: %v", err)
		return
	}

	fmt.Printf("Task cancelled successfully: task-003\n")

	// Example 9: Get health status
	health, err := manager.GetHealth(ctx)
	if err != nil {
		log.Printf("Failed to get health status: %v", err)
		return
	}

	fmt.Printf("Provider health status:\n")
	for provider, status := range health {
		fmt.Printf("- %s: %s\n", provider, status.Status)
	}

	// Example 10: Get metrics
	metrics, err := manager.GetMetrics(ctx)
	if err != nil {
		log.Printf("Failed to get metrics: %v", err)
		return
	}

	fmt.Printf("Provider metrics:\n")
	for provider, metric := range metrics {
		fmt.Printf("- %s: %+v\n", provider, metric)
	}
}

// ExampleWithCustomProvider demonstrates how to use a custom provider
func ExampleWithCustomProvider() {
	logger := logrus.New()
	config := &gateway.ManagerConfig{
		DefaultProvider: "custom",
		RetryAttempts:   2,
		RetryDelay:      time.Second * 2,
		Timeout:         60 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewSchedulingManager(config, logger)

	// Note: In a real implementation, you would register a custom provider here
	// manager.RegisterProvider("custom", customProvider)

	ctx := context.Background()

	// Create a task with custom metadata
	task := &types.Task{
		ID:          "custom-task-001",
		Name:        "Custom Processing",
		Description: "Process data with custom logic",
		Schedule: &types.Schedule{
			Type:     types.ScheduleTypeCron,
			CronExpr: "*/15 * * * *", // Every 15 minutes
		},
		Handler: &types.TaskHandler{
			Type: types.HandlerTypeFunction,
			Function: &types.FunctionHandler{
				FunctionName: "customProcessor",
				Parameters: map[string]interface{}{
					"batch_size": 100,
					"timeout":    300,
					"retry":      true,
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
			"custom_field": "custom_value",
			"processor":    "advanced",
		},
	}

	// This would work once a custom provider is registered
	// result, err := manager.ScheduleTask(ctx, task, "custom")
	// if err != nil {
	//     log.Printf("Failed to schedule custom task: %v", err)
	//     return
	// }

	fmt.Printf("Custom task example prepared (requires custom provider implementation)\n")
}
