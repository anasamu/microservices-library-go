package mocks

import (
	"context"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling/types"
)

// MockSchedulingProvider is a mock implementation of SchedulingProvider
type MockSchedulingProvider struct {
	Name                  string
	SupportedFeatures     []types.SchedulingFeature
	ConnectionInfo        *types.ConnectionInfo
	Connected             bool
	Tasks                 map[string]*types.Task
	Results               map[string]*types.TaskResult
	HealthStatus          *types.HealthStatus
	Metrics               *types.Metrics
	ConnectError          error
	DisconnectError       error
	PingError             error
	ScheduleTaskError     error
	CancelTaskError       error
	GetTaskError          error
	ListTasksError        error
	UpdateTaskError       error
	ScheduleMultipleError error
	CancelMultipleError   error
	GetHealthError        error
	GetMetricsError       error
}

// NewMockSchedulingProvider creates a new mock scheduling provider
func NewMockSchedulingProvider(name string) *MockSchedulingProvider {
	return &MockSchedulingProvider{
		Name: name,
		SupportedFeatures: []types.SchedulingFeature{
			types.FeatureSchedule,
			types.FeatureCancel,
			types.FeatureList,
			types.FeatureUpdate,
			types.FeatureGet,
			types.FeatureHealth,
			types.FeatureMetrics,
		},
		ConnectionInfo: &types.ConnectionInfo{
			Host:     "localhost",
			Port:     0,
			Database: "mock",
			Username: "mock",
			Status:   types.StatusConnected,
		},
		Connected: false,
		Tasks:     make(map[string]*types.Task),
		Results:   make(map[string]*types.TaskResult),
		HealthStatus: &types.HealthStatus{
			Status:    types.StatusHealthy,
			Message:   "Mock provider is healthy",
			Timestamp: time.Now(),
		},
		Metrics: &types.Metrics{
			TotalTasks:     0,
			ScheduledTasks: 0,
			RunningTasks:   0,
			CompletedTasks: 0,
			FailedTasks:    0,
			CancelledTasks: 0,
			SuccessRate:    100.0,
			Timestamp:      time.Now(),
		},
	}
}

// GetName returns the provider name
func (m *MockSchedulingProvider) GetName() string {
	return m.Name
}

// GetSupportedFeatures returns supported features
func (m *MockSchedulingProvider) GetSupportedFeatures() []types.SchedulingFeature {
	return m.SupportedFeatures
}

// GetConnectionInfo returns connection information
func (m *MockSchedulingProvider) GetConnectionInfo() *types.ConnectionInfo {
	info := *m.ConnectionInfo
	if m.Connected {
		info.Status = types.StatusConnected
	} else {
		info.Status = types.StatusDisconnected
	}
	return &info
}

// Connect connects the provider
func (m *MockSchedulingProvider) Connect(ctx context.Context) error {
	if m.ConnectError != nil {
		return m.ConnectError
	}
	m.Connected = true
	return nil
}

// Disconnect disconnects the provider
func (m *MockSchedulingProvider) Disconnect(ctx context.Context) error {
	if m.DisconnectError != nil {
		return m.DisconnectError
	}
	m.Connected = false
	return nil
}

// Ping checks if the provider is responsive
func (m *MockSchedulingProvider) Ping(ctx context.Context) error {
	if m.PingError != nil {
		return m.PingError
	}
	if !m.Connected {
		return &MockError{Message: "provider not connected"}
	}
	return nil
}

// IsConnected returns connection status
func (m *MockSchedulingProvider) IsConnected() bool {
	return m.Connected
}

// ScheduleTask schedules a task
func (m *MockSchedulingProvider) ScheduleTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if m.ScheduleTaskError != nil {
		return nil, m.ScheduleTaskError
	}
	if !m.Connected {
		return nil, &MockError{Message: "provider not connected"}
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
	nextRun := time.Now().Add(time.Hour)
	task.NextRun = &nextRun

	// Store task
	m.Tasks[task.ID] = task

	// Create result
	result := &types.TaskResult{
		TaskID:    task.ID,
		Status:    types.TaskStatusScheduled,
		Message:   "Task scheduled successfully",
		Timestamp: time.Now(),
		NextRun:   &nextRun,
	}

	m.Results[task.ID] = result
	return result, nil
}

// CancelTask cancels a scheduled task
func (m *MockSchedulingProvider) CancelTask(ctx context.Context, taskID string) error {
	if m.CancelTaskError != nil {
		return m.CancelTaskError
	}
	if !m.Connected {
		return &MockError{Message: "provider not connected"}
	}

	task, exists := m.Tasks[taskID]
	if !exists {
		return &MockError{Message: "task not found"}
	}

	// Update task status
	task.Status = types.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	// Update result
	if result, exists := m.Results[taskID]; exists {
		result.Status = types.TaskStatusCancelled
		result.Message = "Task cancelled"
		result.Timestamp = time.Now()
	}

	return nil
}

// GetTask retrieves a task by ID
func (m *MockSchedulingProvider) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	if m.GetTaskError != nil {
		return nil, m.GetTaskError
	}
	if !m.Connected {
		return nil, &MockError{Message: "provider not connected"}
	}

	task, exists := m.Tasks[taskID]
	if !exists {
		return nil, &MockError{Message: "task not found"}
	}

	// Return a copy to prevent external modifications
	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks lists tasks with optional filtering
func (m *MockSchedulingProvider) ListTasks(ctx context.Context, filter *types.TaskFilter) ([]*types.Task, error) {
	if m.ListTasksError != nil {
		return nil, m.ListTasksError
	}
	if !m.Connected {
		return nil, &MockError{Message: "provider not connected"}
	}

	var tasks []*types.Task

	for _, task := range m.Tasks {
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
func (m *MockSchedulingProvider) UpdateTask(ctx context.Context, task *types.Task) error {
	if m.UpdateTaskError != nil {
		return m.UpdateTaskError
	}
	if !m.Connected {
		return &MockError{Message: "provider not connected"}
	}

	_, exists := m.Tasks[task.ID]
	if !exists {
		return &MockError{Message: "task not found"}
	}

	// Update task
	task.UpdatedAt = time.Now()
	m.Tasks[task.ID] = task

	return nil
}

// ScheduleMultiple schedules multiple tasks
func (m *MockSchedulingProvider) ScheduleMultiple(ctx context.Context, tasks []*types.Task) ([]*types.TaskResult, error) {
	if m.ScheduleMultipleError != nil {
		return nil, m.ScheduleMultipleError
	}

	var results []*types.TaskResult
	var errors []error

	for _, task := range tasks {
		result, err := m.ScheduleTask(ctx, task)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		results = append(results, result)
	}

	if len(errors) > 0 {
		return results, &MockError{Message: "some tasks failed to schedule"}
	}

	return results, nil
}

// CancelMultiple cancels multiple tasks
func (m *MockSchedulingProvider) CancelMultiple(ctx context.Context, taskIDs []string) error {
	if m.CancelMultipleError != nil {
		return m.CancelMultipleError
	}

	var errors []error

	for _, taskID := range taskIDs {
		if err := m.CancelTask(ctx, taskID); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return &MockError{Message: "some tasks failed to cancel"}
	}

	return nil
}

// GetHealth returns the health status
func (m *MockSchedulingProvider) GetHealth(ctx context.Context) (*types.HealthStatus, error) {
	if m.GetHealthError != nil {
		return nil, m.GetHealthError
	}

	status := *m.HealthStatus
	if !m.Connected {
		status.Status = types.StatusUnhealthy
		status.Message = "Mock provider is not connected"
	}

	return &status, nil
}

// GetMetrics returns provider metrics
func (m *MockSchedulingProvider) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	if m.GetMetricsError != nil {
		return nil, m.GetMetricsError
	}

	// Calculate metrics from stored tasks
	var scheduled, running, completed, failed, cancelled int64

	for _, task := range m.Tasks {
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

	total := int64(len(m.Tasks))
	var successRate float64
	if total > 0 {
		successRate = float64(completed) / float64(total) * 100
	}

	metrics := *m.Metrics
	metrics.TotalTasks = total
	metrics.ScheduledTasks = scheduled
	metrics.RunningTasks = running
	metrics.CompletedTasks = completed
	metrics.FailedTasks = failed
	metrics.CancelledTasks = cancelled
	metrics.SuccessRate = successRate
	metrics.Timestamp = time.Now()

	return &metrics, nil
}

// MockError represents a mock error
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// SetConnectError sets the error to return from Connect
func (m *MockSchedulingProvider) SetConnectError(err error) {
	m.ConnectError = err
}

// SetDisconnectError sets the error to return from Disconnect
func (m *MockSchedulingProvider) SetDisconnectError(err error) {
	m.DisconnectError = err
}

// SetPingError sets the error to return from Ping
func (m *MockSchedulingProvider) SetPingError(err error) {
	m.PingError = err
}

// SetScheduleTaskError sets the error to return from ScheduleTask
func (m *MockSchedulingProvider) SetScheduleTaskError(err error) {
	m.ScheduleTaskError = err
}

// SetCancelTaskError sets the error to return from CancelTask
func (m *MockSchedulingProvider) SetCancelTaskError(err error) {
	m.CancelTaskError = err
}

// SetGetTaskError sets the error to return from GetTask
func (m *MockSchedulingProvider) SetGetTaskError(err error) {
	m.GetTaskError = err
}

// SetListTasksError sets the error to return from ListTasks
func (m *MockSchedulingProvider) SetListTasksError(err error) {
	m.ListTasksError = err
}

// SetUpdateTaskError sets the error to return from UpdateTask
func (m *MockSchedulingProvider) SetUpdateTaskError(err error) {
	m.UpdateTaskError = err
}

// SetScheduleMultipleError sets the error to return from ScheduleMultiple
func (m *MockSchedulingProvider) SetScheduleMultipleError(err error) {
	m.ScheduleMultipleError = err
}

// SetCancelMultipleError sets the error to return from CancelMultiple
func (m *MockSchedulingProvider) SetCancelMultipleError(err error) {
	m.CancelMultipleError = err
}

// SetGetHealthError sets the error to return from GetHealth
func (m *MockSchedulingProvider) SetGetHealthError(err error) {
	m.GetHealthError = err
}

// SetGetMetricsError sets the error to return from GetMetrics
func (m *MockSchedulingProvider) SetGetMetricsError(err error) {
	m.GetMetricsError = err
}

// SetHealthStatus sets the health status to return
func (m *MockSchedulingProvider) SetHealthStatus(status *types.HealthStatus) {
	m.HealthStatus = status
}

// SetMetrics sets the metrics to return
func (m *MockSchedulingProvider) SetMetrics(metrics *types.Metrics) {
	m.Metrics = metrics
}

// ClearTasks clears all stored tasks
func (m *MockSchedulingProvider) ClearTasks() {
	m.Tasks = make(map[string]*types.Task)
	m.Results = make(map[string]*types.TaskResult)
}

// GetTaskCount returns the number of stored tasks
func (m *MockSchedulingProvider) GetTaskCount() int {
	return len(m.Tasks)
}
