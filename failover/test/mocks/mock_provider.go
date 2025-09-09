package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/failover/types"
)

// MockProvider is a mock implementation of FailoverProvider for testing
type MockProvider struct {
	name               string
	connected          bool
	endpoints          map[string]*types.ServiceEndpoint
	stats              *types.FailoverStats
	events             []*types.FailoverEvent
	config             *types.FailoverConfig
	mu                 sync.RWMutex
	connectError       error
	disconnectError    error
	pingError          error
	registerError      error
	deregisterError    error
	updateError        error
	getEndpointError   error
	listEndpointsError error
	selectError        error
	executeError       error
	healthCheckError   error
	configureError     error
	getConfigError     error
	getStatsError      error
	getEventsError     error
	closeError         error
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:      name,
		endpoints: make(map[string]*types.ServiceEndpoint),
		stats: &types.FailoverStats{
			LastUpdate: time.Now(),
		},
		events: make([]*types.FailoverEvent, 0),
		config: types.DefaultConfig,
	}
}

// GetName returns the provider name
func (m *MockProvider) GetName() string {
	return m.name
}

// GetSupportedFeatures returns supported features
func (m *MockProvider) GetSupportedFeatures() []types.FailoverFeature {
	return []types.FailoverFeature{
		types.FeatureHealthCheck,
		types.FeatureLoadBalancing,
		types.FeatureCircuitBreaker,
		types.FeatureRetry,
		types.FeatureTimeout,
		types.FeatureMonitoring,
	}
}

// GetConnectionInfo returns connection information
func (m *MockProvider) GetConnectionInfo() *types.ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := types.StatusDisconnected
	if m.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "mock-host",
		Port:     8080,
		Status:   status,
		Metadata: map[string]string{"provider": "mock"},
	}
}

// Connect connects to the mock provider
func (m *MockProvider) Connect(ctx context.Context) error {
	if m.connectError != nil {
		return m.connectError
	}

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()

	return nil
}

// Disconnect disconnects from the mock provider
func (m *MockProvider) Disconnect(ctx context.Context) error {
	if m.disconnectError != nil {
		return m.disconnectError
	}

	m.mu.Lock()
	m.connected = false
	m.mu.Unlock()

	return nil
}

// Ping tests the connection
func (m *MockProvider) Ping(ctx context.Context) error {
	if m.pingError != nil {
		return m.pingError
	}

	if !m.IsConnected() {
		return fmt.Errorf("not connected")
	}

	return nil
}

// IsConnected returns connection status
func (m *MockProvider) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// RegisterEndpoint registers an endpoint
func (m *MockProvider) RegisterEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	if m.registerError != nil {
		return m.registerError
	}

	m.mu.Lock()
	m.endpoints[endpoint.ID] = endpoint
	m.mu.Unlock()

	return nil
}

// DeregisterEndpoint deregisters an endpoint
func (m *MockProvider) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	if m.deregisterError != nil {
		return m.deregisterError
	}

	m.mu.Lock()
	delete(m.endpoints, endpointID)
	m.mu.Unlock()

	return nil
}

// UpdateEndpoint updates an endpoint
func (m *MockProvider) UpdateEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	if m.updateError != nil {
		return m.updateError
	}

	m.mu.Lock()
	m.endpoints[endpoint.ID] = endpoint
	m.mu.Unlock()

	return nil
}

// GetEndpoint gets an endpoint by ID
func (m *MockProvider) GetEndpoint(ctx context.Context, endpointID string) (*types.ServiceEndpoint, error) {
	if m.getEndpointError != nil {
		return nil, m.getEndpointError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	endpoint, exists := m.endpoints[endpointID]
	if !exists {
		return nil, fmt.Errorf("endpoint %s not found", endpointID)
	}

	return endpoint, nil
}

// ListEndpoints lists all endpoints
func (m *MockProvider) ListEndpoints(ctx context.Context) ([]*types.ServiceEndpoint, error) {
	if m.listEndpointsError != nil {
		return nil, m.listEndpointsError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	endpoints := make([]*types.ServiceEndpoint, 0, len(m.endpoints))
	for _, endpoint := range m.endpoints {
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// SelectEndpoint selects an endpoint
func (m *MockProvider) SelectEndpoint(ctx context.Context, config *types.FailoverConfig) (*types.FailoverResult, error) {
	if m.selectError != nil {
		return nil, m.selectError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find first healthy endpoint
	for _, endpoint := range m.endpoints {
		if endpoint.Health == types.HealthHealthy {
			return &types.FailoverResult{
				Endpoint: endpoint,
				Duration: 10 * time.Millisecond,
				Strategy: config.Strategy,
			}, nil
		}
	}

	return &types.FailoverResult{
		Error:    &types.FailoverError{Code: types.ErrCodeNoEndpoints, Message: "no healthy endpoints"},
		Duration: 10 * time.Millisecond,
		Strategy: config.Strategy,
	}, nil
}

// ExecuteWithFailover executes a function with failover
func (m *MockProvider) ExecuteWithFailover(ctx context.Context, config *types.FailoverConfig, fn func(*types.ServiceEndpoint) (interface{}, error)) (*types.FailoverResult, error) {
	if m.executeError != nil {
		return nil, m.executeError
	}

	// Select endpoint first
	result, err := m.SelectEndpoint(ctx, config)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return result, nil
	}

	// Execute function
	funcResult, err := fn(result.Endpoint)
	if err != nil {
		return &types.FailoverResult{
			Error:    err,
			Duration: 10 * time.Millisecond,
			Attempts: 1,
			Strategy: config.Strategy,
		}, nil
	}

	return &types.FailoverResult{
		Endpoint: result.Endpoint,
		Duration: 10 * time.Millisecond,
		Attempts: 1,
		Strategy: config.Strategy,
		Metadata: map[string]interface{}{
			"result": funcResult,
		},
	}, nil
}

// HealthCheck performs health check
func (m *MockProvider) HealthCheck(ctx context.Context, endpointID string) (types.HealthStatus, error) {
	if m.healthCheckError != nil {
		return types.HealthUnknown, m.healthCheckError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	endpoint, exists := m.endpoints[endpointID]
	if !exists {
		return types.HealthUnknown, fmt.Errorf("endpoint %s not found", endpointID)
	}

	return endpoint.Health, nil
}

// HealthCheckAll performs health check on all endpoints
func (m *MockProvider) HealthCheckAll(ctx context.Context) (map[string]types.HealthStatus, error) {
	if m.healthCheckError != nil {
		return nil, m.healthCheckError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]types.HealthStatus)
	for id, endpoint := range m.endpoints {
		statuses[id] = endpoint.Health
	}

	return statuses, nil
}

// Configure configures the provider
func (m *MockProvider) Configure(ctx context.Context, config *types.FailoverConfig) error {
	if m.configureError != nil {
		return m.configureError
	}

	m.mu.Lock()
	m.config = config
	m.mu.Unlock()

	return nil
}

// GetConfig returns the current configuration
func (m *MockProvider) GetConfig(ctx context.Context) (*types.FailoverConfig, error) {
	if m.getConfigError != nil {
		return nil, m.getConfigError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config, nil
}

// GetStats returns statistics
func (m *MockProvider) GetStats(ctx context.Context) (*types.FailoverStats, error) {
	if m.getStatsError != nil {
		return nil, m.getStatsError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	m.stats.LastUpdate = time.Now()
	return m.stats, nil
}

// GetEvents returns events
func (m *MockProvider) GetEvents(ctx context.Context, limit int) ([]*types.FailoverEvent, error) {
	if m.getEventsError != nil {
		return nil, m.getEventsError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.events) {
		limit = len(m.events)
	}

	start := len(m.events) - limit
	if start < 0 {
		start = 0
	}

	return m.events[start:], nil
}

// Close closes the provider
func (m *MockProvider) Close(ctx context.Context) error {
	if m.closeError != nil {
		return m.closeError
	}

	return m.Disconnect(ctx)
}

// Test helper methods

// SetConnectError sets an error to be returned by Connect
func (m *MockProvider) SetConnectError(err error) {
	m.connectError = err
}

// SetDisconnectError sets an error to be returned by Disconnect
func (m *MockProvider) SetDisconnectError(err error) {
	m.disconnectError = err
}

// SetPingError sets an error to be returned by Ping
func (m *MockProvider) SetPingError(err error) {
	m.pingError = err
}

// SetRegisterError sets an error to be returned by RegisterEndpoint
func (m *MockProvider) SetRegisterError(err error) {
	m.registerError = err
}

// SetDeregisterError sets an error to be returned by DeregisterEndpoint
func (m *MockProvider) SetDeregisterError(err error) {
	m.deregisterError = err
}

// SetUpdateError sets an error to be returned by UpdateEndpoint
func (m *MockProvider) SetUpdateError(err error) {
	m.updateError = err
}

// SetGetEndpointError sets an error to be returned by GetEndpoint
func (m *MockProvider) SetGetEndpointError(err error) {
	m.getEndpointError = err
}

// SetListEndpointsError sets an error to be returned by ListEndpoints
func (m *MockProvider) SetListEndpointsError(err error) {
	m.listEndpointsError = err
}

// SetSelectError sets an error to be returned by SelectEndpoint
func (m *MockProvider) SetSelectError(err error) {
	m.selectError = err
}

// SetExecuteError sets an error to be returned by ExecuteWithFailover
func (m *MockProvider) SetExecuteError(err error) {
	m.executeError = err
}

// SetHealthCheckError sets an error to be returned by HealthCheck
func (m *MockProvider) SetHealthCheckError(err error) {
	m.healthCheckError = err
}

// SetConfigureError sets an error to be returned by Configure
func (m *MockProvider) SetConfigureError(err error) {
	m.configureError = err
}

// SetGetConfigError sets an error to be returned by GetConfig
func (m *MockProvider) SetGetConfigError(err error) {
	m.getConfigError = err
}

// SetGetStatsError sets an error to be returned by GetStats
func (m *MockProvider) SetGetStatsError(err error) {
	m.getStatsError = err
}

// SetGetEventsError sets an error to be returned by GetEvents
func (m *MockProvider) SetGetEventsError(err error) {
	m.getEventsError = err
}

// SetCloseError sets an error to be returned by Close
func (m *MockProvider) SetCloseError(err error) {
	m.closeError = err
}

// AddEndpoint adds an endpoint for testing
func (m *MockProvider) AddEndpoint(endpoint *types.ServiceEndpoint) {
	m.mu.Lock()
	m.endpoints[endpoint.ID] = endpoint
	m.mu.Unlock()
}

// RemoveEndpoint removes an endpoint for testing
func (m *MockProvider) RemoveEndpoint(endpointID string) {
	m.mu.Lock()
	delete(m.endpoints, endpointID)
	m.mu.Unlock()
}

// SetConnected sets the connection status for testing
func (m *MockProvider) SetConnected(connected bool) {
	m.mu.Lock()
	m.connected = connected
	m.mu.Unlock()
}

// GetEndpointCount returns the number of registered endpoints
func (m *MockProvider) GetEndpointCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.endpoints)
}

// AddEvent adds an event for testing
func (m *MockProvider) AddEvent(event *types.FailoverEvent) {
	m.mu.Lock()
	m.events = append(m.events, event)
	m.mu.Unlock()
}

// UpdateStats updates statistics for testing
func (m *MockProvider) UpdateStats(success bool) {
	m.mu.Lock()
	m.stats.TotalRequests++
	if success {
		m.stats.SuccessfulRequests++
	} else {
		m.stats.FailedRequests++
	}
	m.mu.Unlock()
}
