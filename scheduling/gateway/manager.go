package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/scheduling/types"
	"github.com/sirupsen/logrus"
)

// SchedulingManager manages multiple task scheduling providers
type SchedulingManager struct {
	providers map[string]SchedulingProvider
	logger    *logrus.Logger
	config    *ManagerConfig
	mu        sync.RWMutex
}

// ManagerConfig holds scheduling manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// SchedulingProvider interface for task scheduling backends
type SchedulingProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.SchedulingFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Task scheduling operations
	ScheduleTask(ctx context.Context, task *types.Task) (*types.TaskResult, error)
	CancelTask(ctx context.Context, taskID string) error
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	ListTasks(ctx context.Context, filter *types.TaskFilter) ([]*types.Task, error)
	UpdateTask(ctx context.Context, task *types.Task) error

	// Batch operations
	ScheduleMultiple(ctx context.Context, tasks []*types.Task) ([]*types.TaskResult, error)
	CancelMultiple(ctx context.Context, taskIDs []string) error

	// Health and monitoring
	GetHealth(ctx context.Context) (*types.HealthStatus, error)
	GetMetrics(ctx context.Context) (*types.Metrics, error)
}

// NewSchedulingManager creates a new scheduling manager
func NewSchedulingManager(config *ManagerConfig, logger *logrus.Logger) *SchedulingManager {
	if config == nil {
		config = &ManagerConfig{
			DefaultProvider: "default",
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &SchedulingManager{
		providers: make(map[string]SchedulingProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a scheduling provider
func (sm *SchedulingManager) RegisterProvider(name string, provider SchedulingProvider) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	sm.providers[name] = provider
	sm.logger.Infof("Registered scheduling provider: %s", name)
	return nil
}

// GetProvider returns a scheduling provider by name
func (sm *SchedulingManager) GetProvider(name string) (SchedulingProvider, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if name == "" {
		name = sm.config.DefaultProvider
	}

	provider, exists := sm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns all registered providers
func (sm *SchedulingManager) ListProviders() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	providers := make([]string, 0, len(sm.providers))
	for name := range sm.providers {
		providers = append(providers, name)
	}

	return providers
}

// ScheduleTask schedules a task using the specified provider
func (sm *SchedulingManager) ScheduleTask(ctx context.Context, task *types.Task, providerName string) (*types.TaskResult, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	// Retry logic
	var result *types.TaskResult
	var lastErr error

	for attempt := 0; attempt <= sm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			sm.logger.Warnf("Retrying task scheduling (attempt %d/%d)", attempt+1, sm.config.RetryAttempts+1)
			time.Sleep(sm.config.RetryDelay)
		}

		result, lastErr = provider.ScheduleTask(ctx, task)
		if lastErr == nil {
			sm.logger.Infof("Task scheduled successfully: %s", task.ID)
			return result, nil
		}

		sm.logger.Errorf("Task scheduling failed (attempt %d): %v", attempt+1, lastErr)
	}

	// Try fallback provider if enabled
	if sm.config.FallbackEnabled && providerName != sm.config.DefaultProvider {
		sm.logger.Warnf("Attempting fallback to default provider: %s", sm.config.DefaultProvider)
		fallbackProvider, err := sm.GetProvider(sm.config.DefaultProvider)
		if err == nil {
			result, err = fallbackProvider.ScheduleTask(ctx, task)
			if err == nil {
				sm.logger.Infof("Task scheduled successfully using fallback provider: %s", task.ID)
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to schedule task after %d attempts: %w", sm.config.RetryAttempts+1, lastErr)
}

// CancelTask cancels a task using the specified provider
func (sm *SchedulingManager) CancelTask(ctx context.Context, taskID string, providerName string) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	err = provider.CancelTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	sm.logger.Infof("Task cancelled successfully: %s", taskID)
	return nil
}

// GetTask retrieves a task using the specified provider
func (sm *SchedulingManager) GetTask(ctx context.Context, taskID string, providerName string) (*types.Task, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	task, err := provider.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// ListTasks lists tasks using the specified provider
func (sm *SchedulingManager) ListTasks(ctx context.Context, filter *types.TaskFilter, providerName string) ([]*types.Task, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	tasks, err := provider.ListTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTask updates a task using the specified provider
func (sm *SchedulingManager) UpdateTask(ctx context.Context, task *types.Task, providerName string) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	err = provider.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	sm.logger.Infof("Task updated successfully: %s", task.ID)
	return nil
}

// ScheduleMultiple schedules multiple tasks using the specified provider
func (sm *SchedulingManager) ScheduleMultiple(ctx context.Context, tasks []*types.Task, providerName string) ([]*types.TaskResult, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	results, err := provider.ScheduleMultiple(ctx, tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule multiple tasks: %w", err)
	}

	sm.logger.Infof("Scheduled %d tasks successfully", len(tasks))
	return results, nil
}

// CancelMultiple cancels multiple tasks using the specified provider
func (sm *SchedulingManager) CancelMultiple(ctx context.Context, taskIDs []string, providerName string) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Add timeout to context
	if sm.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.Timeout)
		defer cancel()
	}

	err = provider.CancelMultiple(ctx, taskIDs)
	if err != nil {
		return fmt.Errorf("failed to cancel multiple tasks: %w", err)
	}

	sm.logger.Infof("Cancelled %d tasks successfully", len(taskIDs))
	return nil
}

// GetHealth returns the health status of all providers
func (sm *SchedulingManager) GetHealth(ctx context.Context) (map[string]*types.HealthStatus, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	health := make(map[string]*types.HealthStatus)

	for name, provider := range sm.providers {
		status, err := provider.GetHealth(ctx)
		if err != nil {
			health[name] = &types.HealthStatus{
				Status:    types.StatusUnhealthy,
				Message:   err.Error(),
				Timestamp: time.Now(),
			}
		} else {
			health[name] = status
		}
	}

	return health, nil
}

// GetMetrics returns metrics from all providers
func (sm *SchedulingManager) GetMetrics(ctx context.Context) (map[string]*types.Metrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := make(map[string]*types.Metrics)

	for name, provider := range sm.providers {
		metric, err := provider.GetMetrics(ctx)
		if err != nil {
			sm.logger.Errorf("Failed to get metrics from provider %s: %v", name, err)
			continue
		}
		metrics[name] = metric
	}

	return metrics, nil
}

// ConnectAll connects all registered providers
func (sm *SchedulingManager) ConnectAll(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(sm.providers))

	for name, provider := range sm.providers {
		wg.Add(1)
		go func(name string, provider SchedulingProvider) {
			defer wg.Done()
			if err := provider.Connect(ctx); err != nil {
				sm.logger.Errorf("Failed to connect provider %s: %v", name, err)
				errChan <- fmt.Errorf("provider %s: %w", name, err)
			} else {
				sm.logger.Infof("Connected provider: %s", name)
			}
		}(name, provider)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect some providers: %v", errors)
	}

	return nil
}

// DisconnectAll disconnects all registered providers
func (sm *SchedulingManager) DisconnectAll(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(sm.providers))

	for name, provider := range sm.providers {
		wg.Add(1)
		go func(name string, provider SchedulingProvider) {
			defer wg.Done()
			if err := provider.Disconnect(ctx); err != nil {
				sm.logger.Errorf("Failed to disconnect provider %s: %v", name, err)
				errChan <- fmt.Errorf("provider %s: %w", name, err)
			} else {
				sm.logger.Infof("Disconnected provider: %s", name)
			}
		}(name, provider)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to disconnect some providers: %v", errors)
	}

	return nil
}
