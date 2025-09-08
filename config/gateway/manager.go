package gateway

import (
	"fmt"
	"sync"

	"github.com/anasamu/microservices-library-go/config/types"
)

// Manager manages configuration providers and provides a unified interface
type Manager struct {
	providers map[string]types.ConfigProvider
	current   types.ConfigProvider
	config    *types.Config
	mu        sync.RWMutex
	watchers  []func(*types.Config)
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]types.ConfigProvider),
		watchers:  make([]func(*types.Config), 0),
	}
}

// RegisterProvider registers a configuration provider
func (m *Manager) RegisterProvider(name string, provider types.ConfigProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = provider
}

// SetCurrentProvider sets the current active provider
func (m *Manager) SetCurrentProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	provider, exists := m.providers[name]
	if !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	m.current = provider
	return nil
}

// Load loads configuration from the current provider
func (m *Manager) Load() (*types.Config, error) {
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	if current == nil {
		return nil, fmt.Errorf("no current provider set")
	}

	config, err := current.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	m.mu.Lock()
	m.config = config
	m.mu.Unlock()

	return config, nil
}

// Save saves configuration to the current provider
func (m *Manager) Save(config *types.Config) error {
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	if current == nil {
		return fmt.Errorf("no current provider set")
	}

	err := current.Save(config)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	m.mu.Lock()
	m.config = config
	m.mu.Unlock()

	// Notify watchers
	m.notifyWatchers(config)

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *types.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Watch starts watching for configuration changes
func (m *Manager) Watch(callback func(*types.Config)) error {
	m.mu.Lock()
	m.watchers = append(m.watchers, callback)
	m.mu.Unlock()

	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	if current == nil {
		return fmt.Errorf("no current provider set")
	}

	return current.Watch(func(config *types.Config) {
		m.mu.Lock()
		m.config = config
		m.mu.Unlock()
		m.notifyWatchers(config)
	})
}

// notifyWatchers notifies all registered watchers
func (m *Manager) notifyWatchers(config *types.Config) {
	m.mu.RLock()
	watchers := make([]func(*types.Config), len(m.watchers))
	copy(watchers, m.watchers)
	m.mu.RUnlock()

	for _, watcher := range watchers {
		go watcher(config)
	}
}

// Close closes all providers
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, provider := range m.providers {
		if err := provider.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close provider %s: %w", name, err)
		}
	}

	return lastErr
}

// GetProvider returns a specific provider by name
func (m *Manager) GetProvider(name string) (types.ConfigProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns a list of all registered provider names
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}

	return names
}

// Reload reloads configuration from the current provider
func (m *Manager) Reload() (*types.Config, error) {
	return m.Load()
}

// UpdateConfig updates the current configuration and saves it
func (m *Manager) UpdateConfig(updater func(*types.Config)) error {
	m.mu.Lock()
	config := m.config
	m.mu.Unlock()

	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Create a copy to avoid race conditions
	newConfig := *config
	updater(&newConfig)

	return m.Save(&newConfig)
}
