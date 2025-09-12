package chaos

import (
	"context"
	"fmt"
	"sync"

	"github.com/anasamu/microservices-library-go/chaos/types"
)

// Re-export types for convenience
type ChaosType = types.ChaosType
type ExperimentType = types.ExperimentType
type ExperimentConfig = types.ExperimentConfig
type ExperimentResult = types.ExperimentResult
type ChaosProvider = types.ChaosProvider

// Re-export constants for convenience
const (
	ChaosTypeKubernetes = types.ChaosTypeKubernetes
	ChaosTypeHTTP       = types.ChaosTypeHTTP
	ChaosTypeMessaging  = types.ChaosTypeMessaging

	PodFailure     = types.PodFailure
	NetworkLatency = types.NetworkLatency
	CPUStress      = types.CPUStress
	MemoryStress   = types.MemoryStress

	HTTPLatency = types.HTTPLatency
	HTTPError   = types.HTTPError
	HTTPTimeout = types.HTTPTimeout

	MessageDelay   = types.MessageDelay
	MessageLoss    = types.MessageLoss
	MessageReorder = types.MessageReorder
)

// Manager manages chaos engineering experiments
type Manager struct {
	providers map[ChaosType]ChaosProvider
	mu        sync.RWMutex
}

// NewManager creates a new chaos manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[ChaosType]ChaosProvider),
	}
}

// RegisterProvider registers a chaos provider
func (m *Manager) RegisterProvider(chaosType ChaosType, provider ChaosProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[chaosType] = provider
}

// Initialize initializes all registered providers
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for chaosType, provider := range m.providers {
		if err := provider.Initialize(ctx, nil); err != nil {
			return fmt.Errorf("failed to initialize %s provider: %w", chaosType, err)
		}
	}

	return nil
}

// StartExperiment starts a chaos experiment
func (m *Manager) StartExperiment(ctx context.Context, config ExperimentConfig) (*ExperimentResult, error) {
	m.mu.RLock()
	provider, exists := m.providers[config.Type]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider for chaos type %s not found", config.Type)
	}

	return provider.StartExperiment(ctx, config)
}

// StopExperiment stops a chaos experiment
func (m *Manager) StopExperiment(ctx context.Context, chaosType ChaosType, experimentID string) error {
	m.mu.RLock()
	provider, exists := m.providers[chaosType]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("provider for chaos type %s not found", chaosType)
	}

	return provider.StopExperiment(ctx, experimentID)
}

// GetExperimentStatus gets the status of a chaos experiment
func (m *Manager) GetExperimentStatus(ctx context.Context, chaosType ChaosType, experimentID string) (*ExperimentResult, error) {
	m.mu.RLock()
	provider, exists := m.providers[chaosType]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider for chaos type %s not found", chaosType)
	}

	return provider.GetExperimentStatus(ctx, experimentID)
}

// ListExperiments lists all experiments for a chaos type
func (m *Manager) ListExperiments(ctx context.Context, chaosType ChaosType) ([]*ExperimentResult, error) {
	m.mu.RLock()
	provider, exists := m.providers[chaosType]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider for chaos type %s not found", chaosType)
	}

	return provider.ListExperiments(ctx)
}

// Cleanup cleans up all providers
func (m *Manager) Cleanup(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for chaosType, provider := range m.providers {
		if err := provider.Cleanup(ctx); err != nil {
			return fmt.Errorf("failed to cleanup %s provider: %w", chaosType, err)
		}
	}

	return nil
}

// GetAvailableProviders returns a list of available chaos providers
func (m *Manager) GetAvailableProviders() []ChaosType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]ChaosType, 0, len(m.providers))
	for chaosType := range m.providers {
		providers = append(providers, chaosType)
	}

	return providers
}

// DefaultManager creates a manager with default providers
// Note: This function requires the providers to be imported separately to avoid circular dependencies
func DefaultManager() *Manager {
	manager := NewManager()

	// Note: Providers should be registered by the caller to avoid import cycles
	// Example usage:
	// manager.RegisterProvider(ChaosTypeKubernetes, kubernetes.NewProvider())
	// manager.RegisterProvider(ChaosTypeHTTP, http.NewProvider())
	// manager.RegisterProvider(ChaosTypeMessaging, messaging.NewProvider())

	return manager
}
