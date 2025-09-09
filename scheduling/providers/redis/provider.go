package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisProvider implements SchedulingProvider using Redis
type RedisProvider struct {
	client    *redis.Client
	tasks     map[string]*types.Task
	results   map[string]*types.TaskResult
	mu        sync.RWMutex
	config    *RedisConfig
	logger    *logrus.Logger
	connected bool
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Password     string            `json:"password"`
	Database     int               `json:"database"`
	PoolSize     int               `json:"pool_size"`
	MinIdleConns int               `json:"min_idle_conns"`
	MaxRetries   int               `json:"max_retries"`
	DialTimeout  time.Duration     `json:"dial_timeout"`
	ReadTimeout  time.Duration     `json:"read_timeout"`
	WriteTimeout time.Duration     `json:"write_timeout"`
	Metadata     map[string]string `json:"metadata"`
}

// NewRedisProvider creates a new Redis scheduling provider
func NewRedisProvider(config *RedisConfig, logger *logrus.Logger) (*RedisProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	return &RedisProvider{
		client:    client,
		tasks:     make(map[string]*types.Task),
		results:   make(map[string]*types.TaskResult),
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (rp *RedisProvider) GetName() string {
	return "redis"
}

// GetSupportedFeatures returns supported features
func (rp *RedisProvider) GetSupportedFeatures() []types.SchedulingFeature {
	return []types.SchedulingFeature{
		types.FeatureSchedule,
		types.FeatureCancel,
		types.FeatureList,
		types.FeatureUpdate,
		types.FeatureGet,
		types.FeatureHealth,
		types.FeatureMetrics,
		types.FeatureBatch,
		types.FeatureCron,
		types.FeatureRecurring,
		types.FeatureOneTime,
		types.FeatureInterval,
		types.FeatureRetry,
		types.FeatureTimeout,
		types.FeaturePersistence,
		types.FeatureClustering,
		types.FeatureMonitoring,
	}
}

// GetConnectionInfo returns connection information
func (rp *RedisProvider) GetConnectionInfo() *types.ConnectionInfo {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	status := types.StatusDisconnected
	if rp.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     rp.config.Host,
		Port:     rp.config.Port,
		Database: fmt.Sprintf("db%d", rp.config.Database),
		Username: "redis",
		Status:   status,
		Metadata: rp.config.Metadata,
	}
}

// Connect connects the provider
func (rp *RedisProvider) Connect(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.connected {
		return nil
	}

	// Test connection
	_, err := rp.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	rp.connected = true
	rp.logger.Info("Redis provider connected successfully")

	return nil
}

// Disconnect disconnects the provider
func (rp *RedisProvider) Disconnect(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if !rp.connected {
		return nil
	}

	err := rp.client.Close()
	if err != nil {
		return fmt.Errorf("failed to disconnect from Redis: %w", err)
	}

	rp.connected = false
	rp.logger.Info("Redis provider disconnected successfully")

	return nil
}

// Ping checks if the provider is responsive
func (rp *RedisProvider) Ping(ctx context.Context) error {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if !rp.connected {
		return fmt.Errorf("provider not connected")
	}

	_, err := rp.client.Ping(ctx).Result()
	return err
}

// IsConnected returns connection status
func (rp *RedisProvider) IsConnected() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.connected
}

// ScheduleTask schedules a task
func (rp *RedisProvider) ScheduleTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if task.Schedule == nil {
		return nil, fmt.Errorf("task schedule cannot be nil")
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()

	if !rp.connected {
		return nil, fmt.Errorf("provider not connected")
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
	nextRun, err := rp.calculateNextRun(task.Schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run time: %w", err)
	}
	task.NextRun = &nextRun

	// Store task in Redis
	taskKey := fmt.Sprintf("task:%s", task.ID)
	taskData, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	err = rp.client.Set(ctx, taskKey, taskData, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store task in Redis: %w", err)
	}

	// Store in local cache
	rp.tasks[task.ID] = task

	// Create result
	result := &types.TaskResult{
		TaskID:    task.ID,
		Status:    types.TaskStatusScheduled,
		Message:   "Task scheduled successfully",
		Timestamp: time.Now(),
		NextRun:   &nextRun,
	}

	// Store result in Redis
	resultKey := fmt.Sprintf("result:%s", task.ID)
	resultData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	err = rp.client.Set(ctx, resultKey, resultData, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store result in Redis: %w", err)
	}

	rp.results[task.ID] = result

	rp.logger.Infof("Task scheduled: %s", task.Name)
	return result, nil
}

// CancelTask cancels a scheduled task
func (rp *RedisProvider) CancelTask(ctx context.Context, taskID string) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if !rp.connected {
		return fmt.Errorf("provider not connected")
	}

	// Get task from Redis
	taskKey := fmt.Sprintf("task:%s", taskID)
	taskData, err := rp.client.Get(ctx, taskKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("task not found: %s", taskID)
		}
		return fmt.Errorf("failed to get task from Redis: %w", err)
	}

	var task types.Task
	err = json.Unmarshal([]byte(taskData), &task)
	if err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// Update task status
	task.Status = types.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	// Update task in Redis
	updatedTaskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal updated task: %w", err)
	}

	err = rp.client.Set(ctx, taskKey, updatedTaskData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update task in Redis: %w", err)
	}

	// Update local cache
	rp.tasks[taskID] = &task

	// Update result
	if result, exists := rp.results[taskID]; exists {
		result.Status = types.TaskStatusCancelled
		result.Message = "Task cancelled"
		result.Timestamp = time.Now()

		// Update result in Redis
		resultKey := fmt.Sprintf("result:%s", taskID)
		resultData, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		err = rp.client.Set(ctx, resultKey, resultData, 0).Err()
		if err != nil {
			return fmt.Errorf("failed to update result in Redis: %w", err)
		}
	}

	rp.logger.Infof("Task cancelled: %s", taskID)
	return nil
}

// GetTask retrieves a task by ID
func (rp *RedisProvider) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if !rp.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	// Try local cache first
	if task, exists := rp.tasks[taskID]; exists {
		taskCopy := *task
		return &taskCopy, nil
	}

	// Get from Redis
	taskKey := fmt.Sprintf("task:%s", taskID)
	taskData, err := rp.client.Get(ctx, taskKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task from Redis: %w", err)
	}

	var task types.Task
	err = json.Unmarshal([]byte(taskData), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// Update local cache
	rp.tasks[taskID] = &task

	// Return a copy to prevent external modifications
	taskCopy := task
	return &taskCopy, nil
}

// ListTasks lists tasks with optional filtering
func (rp *RedisProvider) ListTasks(ctx context.Context, filter *types.TaskFilter) ([]*types.Task, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if !rp.connected {
		return nil, fmt.Errorf("provider not connected")
	}

	// Get all task keys
	keys, err := rp.client.Keys(ctx, "task:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get task keys from Redis: %w", err)
	}

	var tasks []*types.Task

	for _, key := range keys {
		taskData, err := rp.client.Get(ctx, key).Result()
		if err != nil {
			rp.logger.Errorf("Failed to get task data for key %s: %v", key, err)
			continue
		}

		var task types.Task
		err = json.Unmarshal([]byte(taskData), &task)
		if err != nil {
			rp.logger.Errorf("Failed to unmarshal task for key %s: %v", key, err)
			continue
		}

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
		taskCopy := task
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
func (rp *RedisProvider) UpdateTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()

	if !rp.connected {
		return fmt.Errorf("provider not connected")
	}

	// Check if task exists
	taskKey := fmt.Sprintf("task:%s", task.ID)
	exists, err := rp.client.Exists(ctx, taskKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Update task
	task.UpdatedAt = time.Now()

	// Store updated task in Redis
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = rp.client.Set(ctx, taskKey, taskData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update task in Redis: %w", err)
	}

	// Update local cache
	rp.tasks[task.ID] = task

	rp.logger.Infof("Task updated: %s", task.ID)
	return nil
}

// ScheduleMultiple schedules multiple tasks
func (rp *RedisProvider) ScheduleMultiple(ctx context.Context, tasks []*types.Task) ([]*types.TaskResult, error) {
	var results []*types.TaskResult
	var errors []error

	for _, task := range tasks {
		result, err := rp.ScheduleTask(ctx, task)
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
func (rp *RedisProvider) CancelMultiple(ctx context.Context, taskIDs []string) error {
	var errors []error

	for _, taskID := range taskIDs {
		if err := rp.CancelTask(ctx, taskID); err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", taskID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some tasks failed to cancel: %v", errors)
	}

	return nil
}

// GetHealth returns the health status
func (rp *RedisProvider) GetHealth(ctx context.Context) (*types.HealthStatus, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	status := types.StatusHealthy
	message := "Redis provider is healthy"

	if !rp.connected {
		status = types.StatusUnhealthy
		message = "Redis provider is not connected"
	} else {
		// Test Redis connection
		_, err := rp.client.Ping(ctx).Result()
		if err != nil {
			status = types.StatusUnhealthy
			message = fmt.Sprintf("Redis connection failed: %v", err)
		}
	}

	// Get Redis info
	info, err := rp.client.Info(ctx, "server").Result()
	if err != nil {
		info = "Unable to get Redis info"
	}

	return &types.HealthStatus{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"total_tasks": len(rp.tasks),
			"host":        rp.config.Host,
			"port":        rp.config.Port,
			"database":    rp.config.Database,
			"redis_info":  info,
		},
	}, nil
}

// GetMetrics returns provider metrics
func (rp *RedisProvider) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	var scheduled, running, completed, failed, cancelled int64

	for _, task := range rp.tasks {
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

	total := int64(len(rp.tasks))
	var successRate float64
	if total > 0 {
		successRate = float64(completed) / float64(total) * 100
	}

	// Get Redis memory usage
	memoryInfo, err := rp.client.Info(ctx, "memory").Result()
	if err != nil {
		memoryInfo = "Unable to get memory info"
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
			"host":        rp.config.Host,
			"port":        rp.config.Port,
			"database":    rp.config.Database,
			"memory_info": memoryInfo,
		},
	}, nil
}

// calculateNextRun calculates the next run time for a schedule
func (rp *RedisProvider) calculateNextRun(schedule *types.Schedule) (time.Time, error) {
	now := time.Now()

	switch schedule.Type {
	case types.ScheduleTypeOnce:
		if schedule.StartTime != nil {
			return *schedule.StartTime, nil
		}
		return now.Add(time.Minute), nil

	case types.ScheduleTypeInterval:
		if schedule.Interval > 0 {
			return now.Add(schedule.Interval), nil
		}
		return now.Add(time.Minute), nil

	case types.ScheduleTypeCron:
		// For cron expressions, we would need a cron parser
		// This is a simplified implementation
		if schedule.CronExpr != "" {
			// In a real implementation, parse the cron expression
			// For now, return a default next run time
			return now.Add(time.Hour), nil
		}
		return now.Add(time.Minute), nil

	default:
		return now.Add(time.Minute), nil
	}
}
