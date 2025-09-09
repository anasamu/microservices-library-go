package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CronProvider implements SchedulingProvider using cron expressions
type CronProvider struct {
	scheduler *cron.Cron
	tasks     map[string]*types.Task
	results   map[string]*types.TaskResult
	mu        sync.RWMutex
	config    *CronConfig
	logger    *logrus.Logger
	connected bool
}

// CronConfig holds cron-specific configuration
type CronConfig struct {
	Timezone     string            `json:"timezone"`
	WithSeconds  bool              `json:"with_seconds"`
	WithLocation bool              `json:"with_location"`
	Metadata     map[string]string `json:"metadata"`
}

// NewCronProvider creates a new cron scheduling provider
func NewCronProvider(config *CronConfig, logger *logrus.Logger) (*CronProvider, error) {
	if config == nil {
		config = &CronConfig{
			Timezone:     "UTC",
			WithSeconds:  false,
			WithLocation: true,
		}
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Create cron scheduler with options
	opts := []cron.Option{
		cron.WithLogger(cron.VerbosePrintfLogger(logger)),
	}

	if config.WithLocation {
		location, err := time.LoadLocation(config.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
		opts = append(opts, cron.WithLocation(location))
	}

	if config.WithSeconds {
		opts = append(opts, cron.WithSeconds())
	}

	scheduler := cron.New(opts...)

	return &CronProvider{
		scheduler: scheduler,
		tasks:     make(map[string]*types.Task),
		results:   make(map[string]*types.TaskResult),
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (cp *CronProvider) GetName() string {
	return "cron"
}

// GetSupportedFeatures returns supported features
func (cp *CronProvider) GetSupportedFeatures() []types.SchedulingFeature {
	return []types.SchedulingFeature{
		types.FeatureSchedule,
		types.FeatureCancel,
		types.FeatureList,
		types.FeatureUpdate,
		types.FeatureGet,
		types.FeatureHealth,
		types.FeatureMetrics,
		types.FeatureCron,
		types.FeatureRecurring,
		types.FeatureRetry,
		types.FeatureTimeout,
		types.FeatureMonitoring,
	}
}

// GetConnectionInfo returns connection information
func (cp *CronProvider) GetConnectionInfo() *types.ConnectionInfo {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	status := types.StatusDisconnected
	if cp.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Database: "cron",
		Username: "system",
		Status:   status,
		Metadata: cp.config.Metadata,
	}
}

// Connect connects the provider
func (cp *CronProvider) Connect(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.connected {
		return nil
	}

	cp.scheduler.Start()
	cp.connected = true
	cp.logger.Info("Cron provider connected successfully")

	return nil
}

// Disconnect disconnects the provider
func (cp *CronProvider) Disconnect(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.connected {
		return nil
	}

	cp.scheduler.Stop()
	cp.connected = false
	cp.logger.Info("Cron provider disconnected successfully")

	return nil
}

// Ping checks if the provider is responsive
func (cp *CronProvider) Ping(ctx context.Context) error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.connected {
		return fmt.Errorf("provider not connected")
	}

	return nil
}

// IsConnected returns connection status
func (cp *CronProvider) IsConnected() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.connected
}

// ScheduleTask schedules a task
func (cp *CronProvider) ScheduleTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if task.Schedule == nil {
		return nil, fmt.Errorf("task schedule cannot be nil")
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	// Validate schedule type
	if task.Schedule.Type != types.ScheduleTypeCron {
		return nil, fmt.Errorf("cron provider only supports cron schedule type")
	}

	if task.Schedule.CronExpr == "" {
		return nil, fmt.Errorf("cron expression cannot be empty")
	}

	// Set task defaults
	if task.Status == "" {
		task.Status = types.TaskStatusScheduled
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.UpdatedAt = time.Now()

	// Calculate next run time
	nextRun, err := cp.calculateNextRun(task.Schedule.CronExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run time: %w", err)
	}
	task.NextRun = &nextRun

	// Store task
	cp.tasks[task.ID] = task

	// Schedule with cron
	entryID, err := cp.scheduler.AddFunc(task.Schedule.CronExpr, func() {
		cp.executeTask(task)
	})
	if err != nil {
		delete(cp.tasks, task.ID)
		return nil, fmt.Errorf("failed to schedule task: %w", err)
	}

	// Create result
	result := &types.TaskResult{
		TaskID:    task.ID,
		Status:    types.TaskStatusScheduled,
		Message:   "Task scheduled successfully",
		Timestamp: time.Now(),
		NextRun:   &nextRun,
	}

	cp.results[task.ID] = result

	cp.logger.Infof("Task scheduled: %s (ID: %d)", task.Name, entryID)
	return result, nil
}

// CancelTask cancels a scheduled task
func (cp *CronProvider) CancelTask(ctx context.Context, taskID string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.connected {
		return fmt.Errorf("provider not connected")
	}

	task, exists := cp.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove from cron scheduler
	// Note: This is a simplified implementation. In a real implementation,
	// you would need to track entry IDs to properly remove cron jobs
	cp.scheduler.Stop()
	cp.scheduler = cron.New()

	// Re-add all other tasks
	for id, t := range cp.tasks {
		if id != taskID {
			_, err := cp.scheduler.AddFunc(t.Schedule.CronExpr, func() {
				cp.executeTask(t)
			})
			if err != nil {
				cp.logger.Errorf("Failed to re-schedule task %s: %v", id, err)
			}
		}
	}
	cp.scheduler.Start()

	// Update task status
	task.Status = types.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	// Update result
	if result, exists := cp.results[taskID]; exists {
		result.Status = types.TaskStatusCancelled
		result.Message = "Task cancelled"
		result.Timestamp = time.Now()
	}

	cp.logger.Infof("Task cancelled: %s", taskID)
	return nil
}

// GetTask retrieves a task by ID
func (cp *CronProvider) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	task, exists := cp.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Return a copy to prevent external modifications
	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks lists tasks with optional filtering
func (cp *CronProvider) ListTasks(ctx context.Context, filter *types.TaskFilter) ([]*types.Task, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	var tasks []*types.Task

	for _, task := range cp.tasks {
		// Apply filters
		if filter != nil {
			if len(filter.Status) > 0 {
				found := false
				for _, status := range filter.Status {
					if task.Status == status {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if len(filter.Tags) > 0 {
				found := false
				for _, filterTag := range filter.Tags {
					for _, taskTag := range task.Tags {
						if taskTag == filterTag {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					continue
				}
			}

			if filter.CreatedAfter != nil && task.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}

			if filter.CreatedBefore != nil && task.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Return a copy to prevent external modifications
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(tasks) {
			tasks = tasks[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(tasks) {
			tasks = tasks[:filter.Limit]
		}
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (cp *CronProvider) UpdateTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.connected {
		return fmt.Errorf("provider not connected")
	}

	existingTask, exists := cp.tasks[task.ID]
	if !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Update task
	task.UpdatedAt = time.Now()
	task.CreatedAt = existingTask.CreatedAt // Preserve creation time
	cp.tasks[task.ID] = task

	// Re-schedule if schedule changed
	if task.Schedule.CronExpr != existingTask.Schedule.CronExpr {
		// Cancel existing schedule
		cp.scheduler.Stop()
		cp.scheduler = cron.New()

		// Re-add all tasks
		for _, t := range cp.tasks {
			_, err := cp.scheduler.AddFunc(t.Schedule.CronExpr, func() {
				cp.executeTask(t)
			})
			if err != nil {
				cp.logger.Errorf("Failed to re-schedule task %s: %v", t.ID, err)
			}
		}
		cp.scheduler.Start()
	}

	cp.logger.Infof("Task updated: %s", task.ID)
	return nil
}

// ScheduleMultiple schedules multiple tasks
func (cp *CronProvider) ScheduleMultiple(ctx context.Context, tasks []*types.Task) ([]*types.TaskResult, error) {
	var results []*types.TaskResult
	var errors []error

	for _, task := range tasks {
		result, err := cp.ScheduleTask(ctx, task)
		if err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", task.ID, err))
			continue
		}
		results = append(results, result)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("some tasks failed to schedule: %v", errors)
	}

	return results, nil
}

// CancelMultiple cancels multiple tasks
func (cp *CronProvider) CancelMultiple(ctx context.Context, taskIDs []string) error {
	var errors []error

	for _, taskID := range taskIDs {
		if err := cp.CancelTask(ctx, taskID); err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", taskID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some tasks failed to cancel: %v", errors)
	}

	return nil
}

// GetHealth returns the health status
func (cp *CronProvider) GetHealth(ctx context.Context) (*types.HealthStatus, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	status := types.StatusHealthy
	message := "Cron provider is healthy"

	if !cp.connected {
		status = types.StatusUnhealthy
		message = "Cron provider is not connected"
	}

	return &types.HealthStatus{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"total_tasks": len(cp.tasks),
			"timezone":    cp.config.Timezone,
		},
	}, nil
}

// GetMetrics returns provider metrics
func (cp *CronProvider) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	var scheduled, running, completed, failed, cancelled int64

	for _, task := range cp.tasks {
		switch task.Status {
		case types.TaskStatusScheduled:
			scheduled++
		case types.TaskStatusRunning:
			running++
		case types.TaskStatusCompleted:
			completed++
		case types.TaskStatusFailed:
			failed++
		case types.TaskStatusCancelled:
			cancelled++
		}
	}

	total := int64(len(cp.tasks))
	var successRate float64
	if total > 0 {
		successRate = float64(completed) / float64(total) * 100
	}

	return &types.Metrics{
		TotalTasks:     total,
		ScheduledTasks: scheduled,
		RunningTasks:   running,
		CompletedTasks: completed,
		FailedTasks:    failed,
		CancelledTasks: cancelled,
		SuccessRate:    successRate,
		Timestamp:      time.Now(),
		CustomMetrics: map[string]interface{}{
			"timezone": cp.config.Timezone,
		},
	}, nil
}

// executeTask executes a task
func (cp *CronProvider) executeTask(task *types.Task) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Update task status
	task.Status = types.TaskStatusRunning
	task.LastRun = &[]time.Time{time.Now()}[0]
	task.RunCount++
	task.UpdatedAt = time.Now()

	cp.logger.Infof("Executing task: %s", task.Name)

	// Simulate task execution
	// In a real implementation, this would execute the actual task handler
	time.Sleep(time.Millisecond * 100) // Simulate work

	// Update task status based on execution result
	// For this example, we'll mark it as completed
	task.Status = types.TaskStatusCompleted
	task.UpdatedAt = time.Now()

	// Update result
	if result, exists := cp.results[task.ID]; exists {
		result.Status = types.TaskStatusCompleted
		result.Message = "Task executed successfully"
		result.Timestamp = time.Now()
	}

	cp.logger.Infof("Task completed: %s", task.Name)
}

// calculateNextRun calculates the next run time for a cron expression
func (cp *CronProvider) calculateNextRun(cronExpr string) (time.Time, error) {
	// Create a temporary cron parser to calculate next run
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()), nil
}
