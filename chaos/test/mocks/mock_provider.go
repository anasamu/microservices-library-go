package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/chaos/types"
)

// MockProvider is a mock implementation of ChaosProvider for testing
type MockProvider struct {
	experiments  map[string]*types.ExperimentResult
	mu           sync.RWMutex
	initialized  bool
	initError    error
	startError   error
	stopError    error
	statusError  error
	listError    error
	cleanupError error
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		experiments: make(map[string]*types.ExperimentResult),
	}
}

// SetInitError sets an error to be returned by Initialize
func (m *MockProvider) SetInitError(err error) {
	m.initError = err
}

// SetStartError sets an error to be returned by StartExperiment
func (m *MockProvider) SetStartError(err error) {
	m.startError = err
}

// SetStopError sets an error to be returned by StopExperiment
func (m *MockProvider) SetStopError(err error) {
	m.stopError = err
}

// SetStatusError sets an error to be returned by GetExperimentStatus
func (m *MockProvider) SetStatusError(err error) {
	m.statusError = err
}

// SetListError sets an error to be returned by ListExperiments
func (m *MockProvider) SetListError(err error) {
	m.listError = err
}

// SetCleanupError sets an error to be returned by Cleanup
func (m *MockProvider) SetCleanupError(err error) {
	m.cleanupError = err
}

// Initialize initializes the mock provider
func (m *MockProvider) Initialize(ctx context.Context, config map[string]interface{}) error {
	if m.initError != nil {
		return m.initError
	}
	m.initialized = true
	return nil
}

// StartExperiment starts a mock experiment
func (m *MockProvider) StartExperiment(ctx context.Context, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	if m.startError != nil {
		return nil, m.startError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	experimentID := fmt.Sprintf("mock-exp-%d", time.Now().UnixNano())
	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Mock experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":       string(config.Experiment),
			"chaos_type": string(config.Type),
			"duration":   config.Duration,
			"intensity":  config.Intensity,
			"target":     config.Target,
			"parameters": config.Parameters,
		},
	}

	m.experiments[experimentID] = result
	return result, nil
}

// StopExperiment stops a mock experiment
func (m *MockProvider) StopExperiment(ctx context.Context, experimentID string) error {
	if m.stopError != nil {
		return m.stopError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	experiment, exists := m.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	experiment.Status = "stopped"
	experiment.EndTime = time.Now().Format(time.RFC3339)

	return nil
}

// GetExperimentStatus gets the status of a mock experiment
func (m *MockProvider) GetExperimentStatus(ctx context.Context, experimentID string) (*types.ExperimentResult, error) {
	if m.statusError != nil {
		return nil, m.statusError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	experiment, exists := m.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// ListExperiments lists all mock experiments
func (m *MockProvider) ListExperiments(ctx context.Context) ([]*types.ExperimentResult, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	experiments := make([]*types.ExperimentResult, 0, len(m.experiments))
	for _, experiment := range m.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments, nil
}

// Cleanup cleans up the mock provider
func (m *MockProvider) Cleanup(ctx context.Context) error {
	if m.cleanupError != nil {
		return m.cleanupError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Mark all experiments as stopped
	for _, experiment := range m.experiments {
		if experiment.Status == "running" {
			experiment.Status = "stopped"
			experiment.EndTime = time.Now().Format(time.RFC3339)
		}
	}

	m.initialized = false
	return nil
}

// GetExperimentCount returns the number of experiments
func (m *MockProvider) GetExperimentCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.experiments)
}

// IsInitialized returns whether the provider is initialized
func (m *MockProvider) IsInitialized() bool {
	return m.initialized
}

// GetExperiment returns a specific experiment by ID
func (m *MockProvider) GetExperiment(experimentID string) (*types.ExperimentResult, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	experiment, exists := m.experiments[experimentID]
	return experiment, exists
}

// SetExperimentStatus sets the status of a specific experiment
func (m *MockProvider) SetExperimentStatus(experimentID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	experiment, exists := m.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	experiment.Status = status
	return nil
}
