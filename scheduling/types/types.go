package types

import (
	"time"
)

// SchedulingFeature represents a scheduling feature
type SchedulingFeature string

const (
	// Basic features
	FeatureSchedule SchedulingFeature = "schedule"
	FeatureCancel   SchedulingFeature = "cancel"
	FeatureList     SchedulingFeature = "list"
	FeatureUpdate   SchedulingFeature = "update"
	FeatureGet      SchedulingFeature = "get"
	FeatureHealth   SchedulingFeature = "health"
	FeatureMetrics  SchedulingFeature = "metrics"

	// Advanced features
	FeatureBatch         SchedulingFeature = "batch"
	FeatureRecurring     SchedulingFeature = "recurring"
	FeatureOneTime       SchedulingFeature = "one_time"
	FeatureCron          SchedulingFeature = "cron"
	FeatureInterval      SchedulingFeature = "interval"
	FeatureRetry         SchedulingFeature = "retry"
	FeatureTimeout       SchedulingFeature = "timeout"
	FeaturePersistence   SchedulingFeature = "persistence"
	FeatureClustering    SchedulingFeature = "clustering"
	FeatureMonitoring    SchedulingFeature = "monitoring"
	FeatureNotifications SchedulingFeature = "notifications"
)

// ConnectionInfo holds connection information for a scheduling provider
type ConnectionInfo struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Database string            `json:"database"`
	Username string            `json:"username"`
	Status   ConnectionStatus  `json:"status"`
	Metadata map[string]string `json:"metadata"`
}

// ConnectionStatus represents the connection status
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusError        ConnectionStatus = "error"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusPaused    TaskStatus = "paused"
)

// ScheduleType represents the type of schedule
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeOnce     ScheduleType = "once"
	ScheduleTypeInterval ScheduleType = "interval"
)

// HandlerType represents the type of task handler
type HandlerType string

const (
	HandlerTypeHTTP     HandlerType = "http"
	HandlerTypeCommand  HandlerType = "command"
	HandlerTypeFunction HandlerType = "function"
	HandlerTypeMessage  HandlerType = "message"
)

// BackoffType represents the type of retry backoff
type BackoffType string

const (
	BackoffTypeFixed       BackoffType = "fixed"
	BackoffTypeLinear      BackoffType = "linear"
	BackoffTypeExponential BackoffType = "exponential"
)

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Status    HealthStatusType       `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthStatusType represents the type of health status
type HealthStatusType string

const (
	StatusHealthy   HealthStatusType = "healthy"
	StatusUnhealthy HealthStatusType = "unhealthy"
	StatusDegraded  HealthStatusType = "degraded"
)

// Task represents a scheduled task
type Task struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Schedule    *Schedule         `json:"schedule"`
	Handler     *TaskHandler      `json:"handler"`
	RetryPolicy *RetryPolicy      `json:"retry_policy,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Status      TaskStatus        `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	NextRun     *time.Time        `json:"next_run,omitempty"`
	LastRun     *time.Time        `json:"last_run,omitempty"`
	RunCount    int64             `json:"run_count"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// Schedule represents a task schedule
type Schedule struct {
	Type      ScheduleType  `json:"type"`
	CronExpr  string        `json:"cron_expr,omitempty"`
	StartTime *time.Time    `json:"start_time,omitempty"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	Interval  time.Duration `json:"interval,omitempty"`
	Timezone  string        `json:"timezone,omitempty"`
}

// TaskHandler represents how a task should be executed
type TaskHandler struct {
	Type     HandlerType      `json:"type"`
	HTTP     *HTTPHandler     `json:"http,omitempty"`
	Command  *CommandHandler  `json:"command,omitempty"`
	Function *FunctionHandler `json:"function,omitempty"`
	Message  *MessageHandler  `json:"message,omitempty"`
}

// HTTPHandler represents an HTTP task handler
type HTTPHandler struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// CommandHandler represents a command task handler
type CommandHandler struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	WorkDir string            `json:"work_dir,omitempty"`
}

// FunctionHandler represents a function task handler
type FunctionHandler struct {
	FunctionName string                 `json:"function_name"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// MessageHandler represents a message task handler
type MessageHandler struct {
	Topic   string            `json:"topic"`
	Message string            `json:"message"`
	Headers map[string]string `json:"headers,omitempty"`
}

// RetryPolicy represents the retry policy for a task
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	Backoff     BackoffType   `json:"backoff"`
	MaxDelay    time.Duration `json:"max_delay,omitempty"`
}

// TaskResult represents the result of scheduling a task
type TaskResult struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Message   string     `json:"message,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	NextRun   *time.Time `json:"next_run,omitempty"`
}

// TaskFilter represents filters for listing tasks
type TaskFilter struct {
	Status        []TaskStatus `json:"status,omitempty"`
	Tags          []string     `json:"tags,omitempty"`
	CreatedAfter  *time.Time   `json:"created_after,omitempty"`
	CreatedBefore *time.Time   `json:"created_before,omitempty"`
	Limit         int          `json:"limit,omitempty"`
	Offset        int          `json:"offset,omitempty"`
}

// Metrics represents metrics from a scheduling provider
type Metrics struct {
	TotalTasks     int64                  `json:"total_tasks"`
	ScheduledTasks int64                  `json:"scheduled_tasks"`
	RunningTasks   int64                  `json:"running_tasks"`
	CompletedTasks int64                  `json:"completed_tasks"`
	FailedTasks    int64                  `json:"failed_tasks"`
	CancelledTasks int64                  `json:"cancelled_tasks"`
	AverageRunTime time.Duration          `json:"average_run_time"`
	SuccessRate    float64                `json:"success_rate"`
	Timestamp      time.Time              `json:"timestamp"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
}

// TaskExecution represents the execution of a task
type TaskExecution struct {
	ID        string        `json:"id"`
	TaskID    string        `json:"task_id"`
	Status    TaskStatus    `json:"status"`
	StartTime time.Time     `json:"start_time"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Error     string        `json:"error,omitempty"`
	Output    string        `json:"output,omitempty"`
	Attempt   int           `json:"attempt"`
}

// Notification represents a task notification
type Notification struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeTaskScheduled NotificationType = "task_scheduled"
	NotificationTypeTaskStarted   NotificationType = "task_started"
	NotificationTypeTaskCompleted NotificationType = "task_completed"
	NotificationTypeTaskFailed    NotificationType = "task_failed"
	NotificationTypeTaskCancelled NotificationType = "task_cancelled"
)

// TaskTemplate represents a template for creating tasks
type TaskTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Handler     *TaskHandler           `json:"handler"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// TaskGroup represents a group of related tasks
type TaskGroup struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	TaskIDs     []string          `json:"task_ids"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Dependency represents a task dependency
type Dependency struct {
	TaskID      string        `json:"task_id"`
	DependentOn string        `json:"dependent_on"`
	Condition   string        `json:"condition,omitempty"`
	WaitTimeout time.Duration `json:"wait_timeout,omitempty"`
}

// TaskWithDependencies represents a task with its dependencies
type TaskWithDependencies struct {
	Task         *Task         `json:"task"`
	Dependencies []*Dependency `json:"dependencies,omitempty"`
}
